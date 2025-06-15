package netbalance

import (
    "fmt"
    "sort"
    "strings"
    "sync"
    "time"
)

// netbalance 包提供网络负载均衡功能
// balance.go 文件实现节点选择和目标地址生成功能

// GetBestDialTarget 获取最佳连接目标节点列表
// 功能：筛选最近活跃的节点，按网络带宽降序排序，生成逗号分隔的目标地址列表
// 参数：port - 目标节点的服务端口号
// 返回值：逗号分隔的目标地址字符串（格式：IP1:PORT,IP2:PORT,...）
func GetBestDialTarget(port string) string {
    lock.Lock()         // 加锁确保并发安全
    defer lock.Unlock() // 函数返回时解锁

    now := time.Now().Unix() // 获取当前时间戳（秒）
    var active []NodeInfo    // 存储活跃节点列表

    // 遍历所有发现的节点，筛选最近15秒内活跃的节点
    for _, node := range discovered {
        if now-node.LastSeen <= 15 {
            active = append(active, node)
        }
    }

    // 按网络带宽降序排序（带宽大的节点优先）
    sort.SliceStable(active, func(i, j int) bool {
        return active[i].NetBandwidth > active[j].NetBandwidth
    })

    var targets []string
    // 将节点信息转换为IP:PORT格式
    for _, node := range active {
        targets = append(targets, fmt.Sprintf("%s:%s", node.IP, port))
    }

    // 返回逗号分隔的目标地址列表
    return strings.Join(targets, ",")
}
