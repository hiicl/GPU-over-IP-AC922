package util

import (
    "log"
    "os"
    "time"
)

// util 包提供通用的工具函数，包括日志记录和错误处理

// debugMode 表示是否启用调试模式，由环境变量DEBUG="1"控制
var (
    debugMode = os.Getenv("DEBUG") == "1"
)

// Log 提供带时间戳的统一日志输出
// format: 日志格式字符串，支持fmt.Printf风格的格式化
// args: 格式化参数
// 示例：Log("Starting server on port %d", 8080)
func Log(format string, args ...interface{}) {
    // 生成当前时间前缀
    prefix := time.Now().Format("2006-01-02 15:04:05")
    // 输出带时间戳的日志
    log.Printf("["+prefix+"] "+format, args...)
}

// DebugLog 仅在调试模式启用时输出日志
// format: 日志格式字符串
// args: 格式化参数
// 注意：是否启用调试模式由DEBUG环境变量控制（DEBUG="1"时启用）
func DebugLog(format string, args ...interface{}) {
    // 检查是否启用调试模式
    if debugMode {
        // 在日志前添加[DEBUG]标记
        Log("[DEBUG] "+format, args...)
    }
}

// Fatal 记录致命错误日志并退出程序
// msg: 错误消息，可以使用%s等格式化标记
// 注意：此函数调用os.Exit(1)终止程序
func Fatal(msg string) {
    // 记录致命错误日志
    Log("[FATAL] %s", msg)
    // 退出程序，返回状态码1
    os.Exit(1)
}
