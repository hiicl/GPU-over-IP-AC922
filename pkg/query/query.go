package query

import (
    "bytes"
    "encoding/csv"
    "log"
    "os/exec"
    "strconv"
    "strings"

    pb "github.com/hiicl/GPU-over-IP-AC922/proto"
)

// query 包提供GPU查询功能，使用nvidia-smi命令行工具获取GPU信息

// GPUStatus 表示GPU的当前使用状态
// UsedMemory: 当前已使用内存（单位：MB）
// Utilization: GPU利用率百分比（0-100）
type GPUStatus struct {
    UsedMemory  int64
    Utilization int32
}

// ListGPUs 查询系统中所有可用的NVIDIA GPU信息
// 返回包含GPU UUID、名称和总内存的GPUInfo对象列表
// 使用nvidia-smi命令查询GPU信息：
//   --query-gpu=uuid,name,memory.total: 查询GPU的UUID、名称和总内存
//   --format=csv,noheader,nounits: 输出CSV格式，无标题行和单位
func ListGPUs() []*pb.GPUInfo {
    // 创建执行nvidia-smi的命令
    cmd := exec.Command("nvidia-smi",
        "--query-gpu=uuid,name,memory.total",
        "--format=csv,noheader,nounits")
    
    // 执行命令并获取输出
    out, err := cmd.Output()
    if err != nil {
        util.Log("nvidia-smi failed: %v", err)
        return nil
    }

    // 创建CSV读取器
    r := csv.NewReader(bytes.NewReader(out))
    r.Comma = ',' // 设置分隔符为逗号

    var result []*pb.GPUInfo
    lines, _ := r.ReadAll() // 读取所有CSV记录
    
    // 遍历每条记录
    for _, line := range lines {
        // 去除每个字段的空格
        for i := range line {
            line[i] = strings.TrimSpace(line[i])
        }
        // 将内存字符串转换为int64
        mem, _ := strconv.ParseInt(line[2], 10, 64)
        
        // 创建GPUInfo对象并添加到结果列表
        result = append(result, &pb.GPUInfo{
            Uuid:        line[0],  // UUID
            Name:        line[1],  // GPU名称
            TotalMemory: mem,      // 总内存（MB）
        })
    }

    return result
}

// GetGPUStatus 根据GPU的UUID获取其当前使用状态
// uuid: 要查询的GPU的唯一标识符
// 返回GPUStatus结构体，包含已使用内存和利用率信息
// 使用nvidia-smi命令查询GPU状态：
//   --query-gpu=uuid,memory.used,utilization.gpu: 查询GPU的UUID、已使用内存和利用率
//   --format=csv,noheader,nounits: 输出CSV格式，无标题行和单位
func GetGPUStatus(uuid string) GPUStatus {
    // 创建执行nvidia-smi的命令
    cmd := exec.Command("nvidia-smi",
        "--query-gpu=uuid,memory.used,utilization.gpu",
        "--format=csv,noheader,nounits")
    
    // 执行命令并获取输出
    out, err := cmd.Output()
    if err != nil {
        util.Log("nvidia-smi status failed: %v", err)
        return GPUStatus{}
    }

    // 将输出按行分割
    lines := strings.Split(string(out), "\n")
    
    // 遍历每行输出
    for _, line := range lines {
        // 按逗号分割字段
        fields := strings.Split(line, ",")
        if len(fields) < 3 {
            continue // 跳过字段不足的行
        }
        
        // 检查UUID是否匹配
        if strings.TrimSpace(fields[0]) == uuid {
            // 解析已使用内存
            used, _ := strconv.ParseInt(strings.TrimSpace(fields[1]), 10, 64)
            // 解析利用率
            util, _ := strconv.Atoi(strings.TrimSpace(fields[2]))
            
            // 返回匹配的GPU状态
            return GPUStatus{
                UsedMemory:  used,         // 已使用内存（MB）
                Utilization: int32(util),  // GPU利用率（0-100）
            }
        }
    }

    // 未找到匹配的GPU，返回空状态
    return GPUStatus{}
}
