package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "google.golang.org/grpc"
    pb "github.com/hiicl/GPU-over-IP-AC922/proto"

    "github.com/hiicl/GPU-over-IP-AC922/modules/netbalance"
)

// cmd/client/main.go 是gRPC服务的客户端入口文件
// 实现与GPU管理服务交互的命令行客户端，支持服务发现和负载均衡
func main() {
    // 1. 获取服务端地址（通过环境变量或自动发现）
    serverAddrs := os.Getenv("GRPC_SERVER") // 从环境变量读取服务器地址，支持多个逗号分隔

    // 如果没有指定服务器，且启用了自动负载均衡，则自动发现节点IP列表
    if serverAddrs == "" && os.Getenv("ENABLE_NETBALANCE") == "true" {
        serverAddrs = netbalance.GetBestDialTarget("50051") // 调用网络平衡模块获取最佳服务节点
        if serverAddrs == "" {
            log.Fatal("自动发现未找到可用节点，请设置 GRPC_SERVER")
        }
        log.Printf("[AutoDiscover] 使用自动发现的服务器列表: %s\n", serverAddrs)
    }

    // 默认回退到本地地址
    if serverAddrs == "" {
        serverAddrs = "localhost:50051"
    }

    // 2. 建立gRPC连接（支持负载均衡）
    // 使用round_robin负载均衡策略连接多个服务器
    // dns:/// 前缀启用DNS解析多个地址
    // WithDefaultServiceConfig 设置负载均衡策略为轮询
    // WithInsecure 禁用TLS（生产环境应使用TLS）
    // WithBlock 等待连接建立
    conn, err := grpc.Dial(
        fmt.Sprintf("dns:///%s", serverAddrs),
        grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
        grpc.WithInsecure(),
        grpc.WithBlock(),
    )
    if err != nil {
        log.Fatalf("Did not connect: %v", err)
    }
    defer conn.Close()

    // 3. 调用ListGPUs接口获取GPU列表
    client := pb.NewGPUServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // 设置5秒超时
    defer cancel()

    listResp, err := client.ListGPUs(ctx, &pb.Void{})
    if err != nil {
        log.Fatalf("Failed to list GPUs: %v", err)
    }

    if len(listResp.Gpus) == 0 {
        log.Println("No GPUs found.")
        return
    }

    // 显示可用GPU列表
    fmt.Println("Available GPUs:")
    for _, g := range listResp.Gpus {
        fmt.Printf("- %s (%s)\n", g.Name, g.Uuid) // 显示GPU名称和UUID
    }

    // 4. 显示第一个GPU的状态
    target := listResp.Gpus[0].Uuid // 选择第一个GPU作为目标
    statResp, err := client.GetGPUStatus(ctx, &pb.GPURequest{Uuid: target})
    if err != nil {
        log.Fatalf("Failed to get GPU status: %v", err)
    }

    // 打印GPU状态信息
    fmt.Printf("Status for GPU %s:\n", target)
    fmt.Printf("  Used Memory: %d MiB\n", statResp.UsedMemory)  // 显示已使用内存
    fmt.Printf("  Utilization: %d%%\n", statResp.Utilization)   // 显示GPU利用率

    // 5. 如果命令行有参数，则将其作为命令在第一个GPU上执行
    if len(os.Args) > 1 {
        cmd := os.Args[1] // 获取命令行参数作为要执行的命令
        runResp, err := client.RunCommand(ctx, &pb.RunRequest{Uuid: target, Cmd: cmd})
        if err != nil {
            log.Fatalf("Command run failed: %v", err)
        }
        // 打印命令输出和退出码
        fmt.Printf("Output:\n%s\nExit Code: %d\n", runResp.Output, runResp.ExitCode)
    }
}
