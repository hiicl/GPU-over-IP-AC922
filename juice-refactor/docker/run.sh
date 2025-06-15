#!/bin/bash
set -e

IMAGE_NAME=juice-server
TAG=latest

echo "[AutoDetect] 正在检测 CUDA 路径..."

# 1. 自动查找 CUDA 路径
CUDA_PATHS=(
    "/usr/local/cuda/lib64"
    "/usr/local/cuda-*/lib64"
    "/usr/lib/powerpc64le-linux-gnu"
    "/usr/lib64"
)

FOUND_CUDA_LIB=""
for path in "${CUDA_PATHS[@]}"; do
    for real in $(ls -d $path 2>/dev/null); do
        if ls $real/libcudart.so* 1>/dev/null 2>&1; then
            FOUND_CUDA_LIB=$real
            echo "[✓] 找到 CUDA 库路径: $FOUND_CUDA_LIB"
            break 2
        fi
    done
done

if [ -z "$FOUND_CUDA_LIB" ]; then
    echo "[×] 未能找到 CUDA 库路径，请手动指定。"
    exit 1
fi

# 2. 检查 nvidia-smi 是否可用
if ! command -v nvidia-smi &> /dev/null; then
    echo "[×] 宿主机未找到 nvidia-smi 命令，可能未安装 NVIDIA 驱动。"
    exit 1
else
    NSMI_PATH=$(command -v nvidia-smi)
    echo "[✓] nvidia-smi 路径: $NSMI_PATH"
fi

# 3. 探测并挂载物理网卡（排除 lo、docker、br、veth、dummy）
echo "[AutoDetect] 正在挂载宿主机物理网卡..."
EXTRA_NET_VOLUMES=""
for netif in $(ls /sys/class/net/); do
    if [[ "$netif" =~ ^(eth|en|ens|eno)[0-9]+$ ]]; then
        echo "[✓] 挂载网卡: /sys/class/net/$netif"
        EXTRA_NET_VOLUMES+=" -v /sys/class/net/$netif:/sys/class/net/$netif:ro"
    else
        echo "[跳过] 非物理网卡: $netif"
    fi
done

# 4. 启动容器
GRPC_PORT=${GRPC_PORT:-50051}
echo "[Launch] 正在启动 Juice Server 容器，监听端口: $GRPC_PORT..."

docker run -it --rm \
    --runtime=nvidia \
    -e NVIDIA_VISIBLE_DEVICES=all \
    -e GRPC_PORT=$GRPC_PORT \
    -v ${FOUND_CUDA_LIB}:${FOUND_CUDA_LIB}:ro \
    -v $NSMI_PATH:$NSMI_PATH:ro \
    -v /dev:/dev \
    $EXTRA_NET_VOLUMES \
    -p ${GRPC_PORT}:${GRPC_PORT} \
    ${IMAGE_NAME}:${TAG}
