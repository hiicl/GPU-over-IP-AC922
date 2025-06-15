package main

import (
    "context"
    "flag"
    "log"
    "net"

    "google.golang.org/grpc"
    pb "juice-refactor/proto"

    "juice-refactor/pkg/query"
    "juice-refactor/pkg/scheduler"
    "juice-refactor/pkg/gpu"
    "juice-refactor/modules/memext"
    "juice-refactor/modules/netbalance"
)

// cmd/grpcserver/main.go 是gRPC服务的主入口文件
// 实现了GPU管理服务的gRPC接口，包括GPU列表查询、状态获取、资源占用和命令执行等功能

// server 结构体实现了GPUService服务的所有RPC方法
type server struct {
    pb.UnimplementedGPUServiceServer
}

// ListGPUs 实现gRPC的ListGPUs接口，返回系统中所有可用GPU的信息列表
func (s *server) ListGPUs(ctx context.Context, _ *pb.Void) (*pb.GPUList, error) {
    // 调用query包获取GPU信息
    infos := query.ListGPUs()
    return &pb.GPUList{Gpus: infos}, nil
}

// GetGPUStatus 实现gRPC的GetGPUStatus接口，返回指定GPU的当前使用状态
func (s *server) GetGPUStatus(ctx context.Context, req *pb.GPURequest) (*pb.GPUStatus, error) {
    // 调用query包获取指定GPU的状态
    st := query.GetGPUStatus(req.Uuid)
    return &pb.GPUStatus{
        UsedMemory:  st.UsedMemory,
        Utilization: st.Utilization,
    }, nil
}

// AcquireGPU 实现gRPC的AcquireGPU接口，尝试占用指定的GPU资源
func (s *server) AcquireGPU(ctx context.Context, req *pb.GPURequest) (*pb.Ack, error) {
    // 调用scheduler包尝试占用GPU
    ok := scheduler.Acquire(req.Uuid)
    if ok {
        return &pb.Ack{Ok: true, Msg: "GPU acquired"}, nil
    }
    return &pb.Ack{Ok: false, Msg: "GPU unavailable"}, nil
}

// ReleaseGPU 实现gRPC的ReleaseGPU接口，释放已占用的GPU资源
func (s *server) ReleaseGPU(ctx context.Context, req *pb.GPURequest) (*pb.Ack, error) {
    // 调用scheduler包释放GPU
    scheduler.Release(req.Uuid)
    return &pb.Ack{Ok: true, Msg: "GPU released"}, nil
}

// RunCommand 实现gRPC的RunCommand接口，在指定GPU上运行命令
func (s *server) RunCommand(ctx context.Context, req *pb.RunRequest) (*pb.RunResponse, error) {
    // 调用gpu包执行命令
    output, code := gpu.RunCommand(req.Uuid, req.Cmd)
    return &pb.RunResponse{
        ExitCode: int32(code),
        Output:   output,
    }, nil
}

// main 函数是gRPC服务的入口点
func main() {
    // 解析命令行参数
    flag.Parse()
    
    // 根据环境变量决定是否启动网络平衡模块
    if os.Getenv("ENABLE_NETBALANCE") == "true" {
        // 启动网络平衡扫描器
        go netbalance.StartScanner()
    }
    
    // 获取服务端口号，默认为50051
    port := os.Getenv("GRPC_PORT")
    if port == "" {
        port = "50051"
    }

    // 创建TCP监听器
    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
        util.Fatal(fmt.Sprintf("failed to listen: %v", err))
    }

    // 创建gRPC服务器实例
    grpcServer := grpc.NewServer()
    // 注册GPUService服务
    pb.RegisterGPUServiceServer(grpcServer, &server{})

    // 初始化显存扩展模块
    memext.Init()

    // 输出服务启动日志
    log.Println("gRPC server is running on port:" + port)

    // 启动gRPC服务
    if err := grpcServer.Serve(lis); err != nil {
        util.Fatal(fmt.Sprintf("failed to serve: %v", err))
    }
}
