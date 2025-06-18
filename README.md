# GPU-over-IP-AC922 项目

## 服务端部署指南 (ppc64le 环境)

### 环境要求
- 使用 IBM 认证的 CUDA 驱动 (≥418 版本)
- 推荐版本：440.x 到 470.x 系列（适配 POWER9 + NVLink）
- 配套使用 nvidia-peermem 模块（特定场景需要）

### 编译注意事项
```bash
# 必须启用的编译标志：
-DCMAKE_CUDA_ARCHITECTURES="70"  # V100 架构支持
-mcpu=power9 -mtune=power9       # CPU 优化
-DUSE_CUDA=ON                    # 启用 CUDA 和 Unified Memory
```
> **重要**：未启用上述标志可能导致 openCAPI 功能异常

### 部署步骤
1. **安装依赖**：
   ```bash
   sudo apt install -y git golang-go docker.io
   ```

2. **获取代码**：
   ```bash
   git clone https://github.com/hiicl/GPU-over-IP-AC922.git
   cd GPU-over-IP-AC922
   ```

3. **构建 Docker 镜像**：
   ```bash
   cd docker
   chmod +x build.sh
   ./build.sh ppc64le
   ```

4. **生成拓扑文件**（宿主机执行）：
   ```bash
   ./cmd/aitherion init
   ```

5. **运行服务容器**：
   ```bash
   docker run -d --name aitherion-server \
       --net host \
       --cpuset-cpus="0-1,4-5" \
       --env ENABLE_MEMEXT=true \
       --env ENABLE_NETBALANCE=true \
       --privileged \
       aitherion-server:latest
   ```

---

## 客户端使用指南 (Windows)

### 环境准备
1. 安装 [Go](https://golang.org/dl/) 和 Git

### 编译与运行
```cmd
:: 获取代码
git clone https://github.com/hiicl/GPU-over-IP-AC922.git
cd GPU-over-IP-AC922\cmd\client

:: 编译客户端
go build -o aitherion-client.exe main.go

:: 运行客户端（替换实际服务器IP）
set GRPC_SERVER=192.168.1.100:50051
aitherion-client.exe nvidia-smi
```

---

## 常见问题排查

### 服务端问题
- **拓扑生成失败**：  
  检查 `/sys/class/drm/card*/device/numa_node` 文件权限
- **容器启动失败**：  
  添加 `--privileged` 参数并验证内核模块加载

### 客户端问题
- **连接失败**：  
  1. 检查服务器防火墙设置（端口 50051）  
  2. 验证 `GRPC_SERVER` 环境变量设置  
  3. 确认网络可达性
