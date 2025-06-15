// Package netbalance 提供网络接口带宽检测功能
// 包含以下核心功能：
//  1. 自动检测系统默认网络接口
//  2. 执行ethtool命令获取接口带宽信息
//  3. 解析ethtool输出中的速度值
package netbalance

import (
    "bufio"
    "bytes"
    "context"
    "errors"
    "fmt"
    "os/exec"
    "regexp"
    "runtime"
    "strings"
    "time"
)

// Scanner 用于扫描网络接口的带宽信息
// 字段说明：
//   InterfaceName: 要扫描的网络接口名称
//   Timeout:       ethtool命令执行超时时间
type Scanner struct {
    InterfaceName string
    Timeout       time.Duration
}

// NewScanner 创建并初始化Scanner实例
// 参数说明：
//   iface:   网络接口名称，如果为空则自动检测默认接口
//   timeout: ethtool命令超时时间，如果为0则使用默认3秒超时
func NewScanner(iface string, timeout time.Duration) *Scanner {
    if iface == "" {
        iface = detectDefaultInterface()
    }
    if timeout == 0 {
        timeout = 3 * time.Second
    }
    return &Scanner{
        InterfaceName: iface,
        Timeout:       timeout,
    }
}

// detectDefaultInterface 检测系统默认网络接口
// 实现逻辑：
//   1. 执行 `ip -o link show` 命令获取所有接口列表
//   2. 使用正则表达式解析接口名称
//   3. 跳过虚拟接口（lo, docker, veth等）
//   4. 返回第一个非虚拟接口
//   5. 如果失败则回退到fallbackInterface()
func detectDefaultInterface() string {
    // 执行ip命令获取网络接口信息
    out, err := exec.Command("ip", "-o", "link", "show").Output()
    if err != nil {
        fmt.Printf("[netbalance] failed to run ip link: %v\n", err)
        return fallbackInterface()
    }

    scanner := bufio.NewScanner(bytes.NewReader(out))
    // 正则匹配接口名称，格式示例: "1: eth0: <BROADCAST,MULTICAST> ..."
    re := regexp.MustCompile(`^\d+: ([^:@]+)[:@]`)

    for scanner.Scan() {
        line := scanner.Text()
        matches := re.FindStringSubmatch(line)
        if len(matches) < 2 {
            continue // 跳过不匹配的行
        }

        iface := matches[1]
        if isVirtualInterface(iface) {
            continue // 跳过虚拟接口
        }

        return iface // 找到有效接口
    }

    return fallbackInterface() // 未找到则使用回退接口
}

// isVirtualInterface 判断是否为虚拟接口
// 检查接口名称是否以虚拟接口前缀开头
func isVirtualInterface(iface string) bool {
    skipPrefixes := []string{"lo", "docker", "br", "veth", "dummy", "ip6tnl", "virbr"}
    for _, prefix := range skipPrefixes {
        if strings.HasPrefix(iface, prefix) {
            return true
        }
    }
    return false
}

// fallbackInterface 回退接口选择逻辑
// 根据系统架构返回默认接口名称
func fallbackInterface() string {
    if runtime.GOARCH == "ppc64le" {
        return "ens1f0" // PowerPC架构默认接口
    }
    return "eth0" // 其他架构默认接口
}

// ExecEthtoolSpeed 执行ethtool命令获取接口速度
// 返回值: 速度(Mbps), 错误信息
// 实现流程:
//   1. 创建带超时的上下文
//   2. 执行ethtool命令并获取输出管道
//   3. 逐行扫描输出，查找"Speed:"开头的行
//   4. 解析速度值并转换单位
//   5. 处理各种错误情况
func (s *Scanner) ExecEthtoolSpeed() (int, error) {
    // 创建带超时的上下文
    ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
    defer cancel() // 确保取消上下文释放资源

    // 准备ethtool命令
    cmd := exec.CommandContext(ctx, "ethtool", s.InterfaceName)
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return 0, fmt.Errorf("failed to get stdout pipe: %w", err)
    }

    // 启动命令
    if err := cmd.Start(); err != nil {
        return 0, fmt.Errorf("failed to start ethtool: %w", err)
    }

    scanner := bufio.NewScanner(stdout)
    var speedMbps int
    found := false // 标记是否找到速度信息

    // 扫描命令输出
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if strings.HasPrefix(line, "Speed:") {
            fields := strings.Fields(line)
            if len(fields) < 2 {
                continue // 跳过格式不正确的行
            }

            speedStr := fields[1] // 获取速度字符串(如"1000Mb/s")
            speed, err := parseSpeed(speedStr)
            if err != nil {
                return 0, fmt.Errorf("parse speed failed: %w", err)
            }
            speedMbps = speed
            found = true
            break // 找到速度信息后退出循环
        }
    }

    // 处理扫描错误
    if err := scanner.Err(); err != nil {
        return 0, fmt.Errorf("reading ethtool output failed: %w", err)
    }

    // 等待命令完成
    if err := cmd.Wait(); err != nil {
        return 0, fmt.Errorf("ethtool command failed: %w", err)
    }

    // 验证是否找到有效速度
    if !found || speedMbps == 0 {
        return 0, errors.New("speed info not found in ethtool output")
    }
    return speedMbps, nil
}

// parseSpeed 解析速度字符串并转换为Mbps
// 支持单位: Gb/s, Mb/s
// 示例:
//   "1Gb/s" -> 1000
//   "100Mb/s" -> 100
func parseSpeed(speedStr string) (int, error) {
    speedStr = strings.ToLower(speedStr)
    var multiplier int

    // 确定单位转换系数
    switch {
    case strings.HasSuffix(speedStr, "gb/s"):
        multiplier = 1000 // Gb/s -> 转换为Mbps
        speedStr = strings.TrimSuffix(speedStr, "gb/s")
    case strings.HasSuffix(speedStr, "mb/s"):
        multiplier = 1 // Mb/s 保持不变
        speedStr = strings.TrimSuffix(speedStr, "mb/s")
    default:
        return 0, fmt.Errorf("unsupported speed unit in %s", speedStr)
    }

    // 提取数值部分
    var val int
    _, err := fmt.Sscanf(speedStr, "%d", &val)
    if err != nil {
        return 0, fmt.Errorf("failed to parse speed value: %w", err)
    }

    return val * multiplier, nil
}
