package netbalance

import (
    "encoding/json"
    "errors"
    "sync"
)

// netbalance 包提供网络负载均衡功能，管理分布式GPU节点的状态信息
// 实现节点信息的存储、更新和序列化/反序列化操作

// GPUInfo 表示单个GPU的状态信息
// 用于在节点间交换GPU状态数据
type GPUInfo struct {
    ID          string `json:"id"`          // GPU的唯一标识符
    Utilization int    `json:"utilization"` // GPU利用率百分比（0-100）
    MemoryUsed  int    `json:"memory_used"` // 已使用内存（单位：MB）
}

// NodeInfo 表示一个节点（服务器）的状态
// 包含节点的IP地址和该节点上所有GPU的状态信息
// 使用互斥锁确保并发安全
type NodeInfo struct {
    IP   string    `json:"ip"`   // 节点的IP地址
    GPUs []GPUInfo `json:"gpus"` // 该节点上的GPU状态列表
    mu   sync.RWMutex            // 读写锁，保护GPUs字段的并发访问
}

// NewNodeInfo 创建并初始化一个新的节点信息结构体
// ip: 节点的IP地址
// gpuCount: 该节点上的GPU数量
// 返回初始化后的NodeInfo指针
func NewNodeInfo(ip string, gpuCount int) *NodeInfo {
    return &NodeInfo{
        IP:   ip,
        GPUs: make([]GPUInfo, gpuCount), // 初始化GPU状态列表（指定长度）
    }
}

// UpdateGPU 更新指定GPU索引的状态信息
// index: 要更新的GPU在节点中的索引位置（从0开始）
// gpu: 新的GPU状态信息
// 返回错误如果索引超出范围
// 注意：此操作是线程安全的
func (n *NodeInfo) UpdateGPU(index int, gpu GPUInfo) error {
    // 加写锁确保并发安全
    n.mu.Lock()
    defer n.mu.Unlock() // 确保函数返回时解锁

    // 检查索引是否有效
    if index < 0 || index >= len(n.GPUs) {
        return errors.New("gpu index out of range")
    }
    
    // 更新指定索引处的GPU信息
    n.GPUs[index] = gpu
    return nil
}

// GetGPU 获取指定GPU索引的状态信息
// index: 要获取的GPU在节点中的索引位置
// 返回GPU状态信息和可能的错误
// 注意：此操作是线程安全的
func (n *NodeInfo) GetGPU(index int) (GPUInfo, error) {
    // 加读锁确保并发安全
    n.mu.RLock()
    defer n.mu.RUnlock() // 确保函数返回时解锁

    // 检查索引是否有效
    if index < 0 || index >= len(n.GPUs) {
        return GPUInfo{}, errors.New("gpu index out of range")
    }
    
    // 返回指定索引处的GPU信息
    return n.GPUs[index], nil
}

// ToJSON 将节点信息序列化为JSON字符串
// 返回JSON字符串和可能的错误
// 注意：此操作是线程安全的（使用读锁）
func (n *NodeInfo) ToJSON() (string, error) {
    // 加读锁确保并发安全
    n.mu.RLock()
    defer n.mu.RUnlock() // 确保函数返回时解锁

    // 将节点信息序列化为JSON
    data, err := json.Marshal(n)
    if err != nil {
        return "", err
    }
    
    // 返回JSON字符串
    return string(data), nil
}

// FromJSON 从JSON字符串反序列化为NodeInfo对象
// data: 包含节点信息的JSON字符串
// 返回NodeInfo指针和可能的错误
func FromJSON(data string) (*NodeInfo, error) {
    var node NodeInfo
    // 将JSON字符串反序列化为NodeInfo结构
    err := json.Unmarshal([]byte(data), &node)
    if err != nil {
        return nil, err
    }
    return &node, nil
}
