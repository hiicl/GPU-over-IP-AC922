package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "net"
    "os"
    "strconv"

    "google.golang.org/grpc"
    pb "juice-refactor/proto"

    "juice-refactor/pkg/query"
    "juice-refactor/pkg/scheduler"
    "juice-refactor/pkg/gpu"
    "juice-refactor/modules/memext"
    "juice-refactor/modules/netbalance"
)

// 单个服务结构（绑定一组GPU）
type server struct {
    pb.UnimplementedGPUServiceServer
    boundGPUs map[string]bool
}

// 只处理绑定的GPU
func (s *server) ListGPUs(ctx context.Context, _ *pb.Void) (*pb.GPUList, error) {
    infos := query.ListGPUs()
    var filtered []*pb.GPUInfo
    for _, g := range infos {
        if s.boundGPUs[g.Uuid] {
            filtered = append(filtered, g)
        }
    }
    return &pb.GPUList{Gpus: filtered}, nil
}

func (s *server) GetGPUStatus(ctx context.Context, req *pb.GPURequest) (*pb.GPUStatus, error) {
    if !s.boundGPUs[req.Uuid] {
        return nil, fmt.Errorf("GPU %s not bound to this NUMA group", req.Uuid)
    }
    st := query.GetGPUStatus(req.Uuid)
    return &pb.GPUStatus{UsedMemory: st.UsedMemory, Utilization: st.Utilization}, nil
}

func (s *server) AcquireGPU(ctx context.Context, req *pb.GPURequest) (*pb.Ack, error) {
    if !s.boundGPUs[req.Uuid] {
        return &pb.Ack{Ok: false, Msg: "GPU not bound"}, nil
    }
    ok := scheduler.Acquire(req.Uuid)
    return &pb.Ack{Ok: ok, Msg: "acquire result"}
}

func (s *server) ReleaseGPU(ctx context.Context, req *pb.GPURequest) (*pb.Ack, error) {
    scheduler.Release(req.Uuid)
    return &pb.Ack{Ok: true, Msg: "released"}
}

func (s *server) RunCommand(ctx context.Context, req *pb.RunRequest) (*pb.RunResponse, error) {
    if !s.boundGPUs[req.Uuid] {
        return nil, fmt.Errorf("GPU %s not bound to this NUMA group", req.Uuid)
    }
    output, code := gpu.RunCommand(req.Uuid, req.Cmd)
    return &pb.RunResponse{ExitCode: int32(code), Output: output}, nil
}

// 启动多个 NUMA 分组的 gRPC 服务
func main() {
    flag.Parse()
    memext.Init()

    // 自动获取 NUMA 拓扑
    groups, err := netbalance.GetNUMATopology()
    if err != nil {
        log.Fatalf("[Fatal] Failed to get NUMA topology: %v", err)
    }

    basePort := 50051
    for i, group := range groups {
        port := basePort + i
        gpuUUIDs := query.GPUUUIDsByIDs(group.GPUIDs)

        go func(p int, gpus []string, nics []string) {
            log.Printf("[Launch] NUMA %d listening on :%d, GPUs=%v, NICs=%v", group.NUMANode, p, gpus, nics)
            runGRPCServer(p, gpus)
        }(port, gpuUUIDs, group.NetIfaces)
    }

    select {} // 阻塞主线程
}

// 启动一个 gRPC Server 并绑定 GPU UUIDs
func runGRPCServer(port int, gpuUUIDs []string) {
    bound := make(map[string]bool)
    for _, uuid := range gpuUUIDs {
        bound[uuid] = true
    }

    lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))
    if err != nil {
        log.Fatalf("[Fatal] Failed to listen on port %d: %v", port, err)
    }

    grpcServer := grpc.NewServer()
    pb.RegisterGPUServiceServer(grpcServer, &server{boundGPUs: bound})

    log.Printf("[OK] gRPC server ready on :%d", port)
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("[Fatal] Failed to serve gRPC: %v", err)
    }
}
