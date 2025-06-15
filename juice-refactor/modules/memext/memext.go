package memext

import (
    "log"
    "os"
    "strconv"
)

// memext 模块提供显存扩展功能，允许将主机内存映射为GPU显存
// 通过环境变量控制是否启用扩展以及扩展内存大小

// Init 初始化显存扩展模块
// 根据环境变量ENABLE_MEM_EXT决定是否启用主机内存映射功能
// 如果启用，则调用enableUnifiedMemory函数进行配置
func Init() {
    util.Log("[MemExt] 初始化显存扩展模块...")

    // 检查是否启用了显存扩展功能
    if os.Getenv("ENABLE_MEM_EXT") == "1" {
        // 启用主机内存映射
        enableUnifiedMemory()
    } else {
        // 未启用显存扩展
        log.Println("[MemExt] 显存扩展未启用")
    }
}

// enableUnifiedMemory 启用主机内存映射功能（模拟实现）
// 在实际应用中，此函数会配置CUDA unified memory或MLU缓存管理
// 同时读取MEM_EXT_MB环境变量获取用户指定的扩展内存大小
func enableUnifiedMemory() {
    // 记录启用日志
    util.Log("[MemExt] 启用了主机内存映射（模拟）")

    // 获取用户设置的扩展内存大小（MB）
    memLimit := os.Getenv("MEM_EXT_MB")
    if memLimit != "" {
        // 将字符串转换为整数
        if mb, err := strconv.Atoi(memLimit); err == nil {
            // 记录扩展的内存大小
            util.Log("[MemExt] 已申请扩展 %d MiB 主机内存供 GPU 使用（模拟）", mb)
        }
    }
}
