服务端部署教程（ppc64le环境）：

使用 IBM 认证的 CUDA 驱动（>= 418 版本）
建议版本：440.x 到 470.x 系列，适配 POWER9 + NVLink
配套使用 nvidia-peermem 模块（某些场景下）

编译时开启以下标志：
-DCMAKE_CUDA_ARCHITECTURES="70"（V100）
-mcpu=power9 -mtune=power9（CPU 优化）
-DUSE_CUDA=ON 并启用 Unified Memory 支持

否则从理论上可能实现不了openCAPI

1. 安装依赖
sudo apt install -y git golang-go docker.io  

2. 获取代码
git clone https://github.com/hiicl/GPU-over-IP-AC922.git
cd GPU-over-IP-AC922

3. 构建Docker镜像
cd docker
chmod +x build.sh
./build.sh ppc64le

4. 生成拓扑文件（需在宿主机执行）
./cmd/aitherion init

5. 运行容器
docker run -d --name aitherion-server \
    --net host \
    --cpuset-cpus="0-1,4-5" \
    --env ENABLE_MEMEXT=true \
    --env ENABLE_NETBALANCE=true \
    --privileged \
    aitherion-server:latest

客户端Windows编译教程：

cmd
1. 安装Go和Git
下载地址：https://golang.org/dl/

2. 获取代码
git clone https://github.com/hiicl/GPU-over-IP-AC922.git
cd GPU-over-IP-AC922\cmd\client

3. 编译客户端
go build -o aitherion-client.exe main.go

4. 运行客户端
set GRPC_SERVER=192.168.1.100:50051
aitherion-client.exe nvidia-smi

常见问题：
拓扑生成失败：检查/sys/class/drm/card*/device/numa_node权限
容器启动失败：添加--privileged参数
客户端连接失败：检查防火墙和端口(50051)
