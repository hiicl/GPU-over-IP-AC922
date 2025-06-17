package gpu

import (
    "context"
    "errors"
    "fmt"
    "os/exec"
    "strings"
    "time"
)

// gpu 包提供GPU信息查询功能，使用nvidia-smi命令行工具获取GPU详细信息

// GPUInfo 结构体定义GPU的详细状态信息
// UUID: GPU的唯一标识符
// Name: GPU型号名称
// MemoryUsed: 当前已使用内存（单位：MB）
// MemoryTotal: GPU总内存（单位：MB）
// Utilization: GPU利用率百分比（0-100）
type GPUInfo struct {
    UUID        string // GPU唯一标识符
    Name        string // GPU型号名称
    MemoryUsed  int    // 已使用内存（MB）
    MemoryTotal int    // 总内存（MB）
    Utilization int    // GPU利用率百分比（0-100）
}

// execCommandWithTimeout 执行命令行命令，带有超时控制
// timeout: 命令执行的最大等待时间（超时返回错误）
// name: 要执行的命令名称
// args: 命令参数
// 返回命令输出和可能的错误
// 注意：使用context.WithTimeout确保命令不会无限期阻塞
func execCommandWithTimeout(timeout time.Duration, name string, args ...string) (string, error) {
    // 创建带超时的上下文（超时自动取消）
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel() // 确保在函数返回时取消上下文

    // 创建命令对象（绑定到上下文）
    cmd := exec.CommandContext(ctx, name, args...)
    // 获取命令输出（标准输出和错误输出合并）
    outputBytes, err := cmd.CombinedOutput()

    // 检查是否因超时而取消
    if ctx.Err() == context.DeadlineExceeded {
        return "", errors.New("command timed out")
    }

    // 检查其他执行错误
    if err != nil {
        // 返回包含错误信息和原始输出的详细错误
        return "", fmt.Errorf("command execution failed: %v, output: %s", err, string(outputBytes))
    }

    // 返回命令输出（转换为字符串）
    return string(outputBytes), nil
}

// QueryGPUs 查询系统中所有NVIDIA GPU的当前状态信息
// 返回GPUInfo切片（每个GPU一个）和可能的错误
// 功能：通过nvidia-smi命令获取GPU的UUID、名称、内存使用情况和利用率
// 安全：使用预定义参数避免命令注入风险
func QueryGPUs() ([]GPUInfo, error) {
    // 调用带超时的nvidia-smi命令查询GPU信息
    output, err := execCommandWithTimeout(3*time.Second, // 设置3秒超时
        "nvidia-smi",
        "--query-gpu=uuid,name,memory.used,memory.total,utilization.gpu", // 查询GPU关键指标
        "--format=csv,noheader,nounits") // CSV格式输出，无表头/单位
    if err != nil {
        return nil, err
    }

    // 按行分割输出（并去除首尾空白）
    lines := strings.Split(strings.TrimSpace(output), "\n")
    var gpus []GPUInfo // 存储解析后的GPU信息

    // 处理每一行数据（每行对应一个GPU）
    for _, line := range lines {
        // 按逗号分割字段（注意nvidia-smi使用", "分隔）
        fields := strings.Split(line, ", ")
        if len(fields) != 5 {
            // 跳过字段数量不正确的行（格式错误）
            continue
        }

        // 解析各个数值字段（内存使用、总内存、利用率）
        memUsed, err1 := parseInt(fields[2]) // 已使用内存
        memTotal, err2 := parseInt(fields[3]) // 总内存
        util, err3 := parseInt(fields[4])     // 利用率

        // 跳过任何解析失败的行
        if err1 != nil || err2 != nil || err3 != nil {
            continue
        }

        // 创建GPUInfo对象并初始化
        gpu := GPUInfo{
            UUID:        fields[0], // GPU UUID
            Name:        fields[1], // GPU型号名称
            MemoryUsed:  memUsed,   // 已使用内存(MB)
            MemoryTotal: memTotal,  // 总内存(MB)
            Utilization: util,      // 利用率百分比(0-100)
        }
        gpus = append(gpus, gpu) // 添加到结果切片
    }

    // 检查是否找到有效GPU信息
    if len(gpus) == 0 {
        return nil, errors.New("no valid GPU info found")
    }
    return gpus, nil
}

// parseInt 辅助函数，将字符串解析为整数
// s: 要解析的字符串（包含数字）
// 返回解析后的整数值和可能的错误
// 功能：简化字符串到整数的转换
func parseInt(s string) (int, error) {
    var i int
    // 使用fmt.Sscanf安全解析整数
    _, err := fmt.Sscanf(s, "%d", &i)
    if err != nil {
        return 0, err
    }
    return i, nil
}
