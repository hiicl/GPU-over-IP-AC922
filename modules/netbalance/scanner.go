package netbalance

import (
    "bufio"
    "bytes"
    "context"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "runtime"
    "strconv"
    "strings"
    "time"
)

// Scanner 结构体用于扫描网络接口信息
type Scanner struct {
    InterfaceName string   // 网络接口名称
    Timeout       time.Duration // 扫描超时时间
}

// NetDeviceInfo 存储网络设备信息
type NetDeviceInfo struct {
    Name     string // 接口名称
    NUMANode int    // NUMA节点
    IsVirtual bool  // 是否为虚拟接口
}

// NewScanner 创建新的Scanner实例
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

// detectDefaultInterface 检测默认网络接口（返回第一个物理网卡）
// 仍然保留原逻辑，返回首个非虚拟网卡
func detectDefaultInterface() string {
    infos := detectAllPhysicalInterfaces()
    if len(infos) > 0 {
        return infos[0].Name
    }
    return fallbackInterface()
}

// detectAllPhysicalInterfaces 检测所有物理网卡并返回其信息
// 返回所有非虚拟网卡及其 NUMA 编号
func detectAllPhysicalInterfaces() []NetDeviceInfo {
    out, err := exec.Command("ip", "-o", "link", "show").Output()
    if err != nil {
        fmt.Printf("[netbalance] failed to run ip link: %v\n", err)
        return nil
    }

    scanner := bufio.NewScanner(bytes.NewReader(out))
    re := regexp.MustCompile(`^\d+: ([^:@]+)[:@]`)

    var results []NetDeviceInfo

    for scanner.Scan() {
        line := scanner.Text()
        matches := re.FindStringSubmatch(line)
        if len(matches) < 2 {
            continue
        }

        iface := matches[1]
        if isVirtualInterface(iface) {
            continue
        }

        numaNode := GetInterfaceNUMANode(iface)
        results = append(results, NetDeviceInfo{
            Name: iface,
            NUMANode: numaNode,
            IsVirtual: false,
        })
    }

    return results
}

// GetInterfaceNUMANode 获取网络接口的NUMA节点
// 读取 sysfs 获取网卡对应 NUMA 编号
func GetInterfaceNUMANode(iface string) int {
    path := filepath.Join("/sys/class/net", iface, "device", "numa_node")
    data, err := os.ReadFile(path)
    if err != nil {
        return -1
    }

    val := strings.TrimSpace(string(data))
    numa, err := strconv.Atoi(val)
    if err != nil || numa < 0 {
        return -1
    }
    return numa
}

// isVirtualInterface 判断是否为虚拟网络接口
func isVirtualInterface(iface string) bool {
    skipPrefixes := []string{"lo", "docker", "br", "veth", "dummy", "ip6tnl", "virbr"}
    for _, prefix := range skipPrefixes {
        if strings.HasPrefix(iface, prefix) {
            return true
        }
    }
    return false
}

// fallbackInterface 回退接口（根据架构返回默认接口名）
func fallbackInterface() string {
    if runtime.GOARCH == "ppc64le" {
        return "ens1f0"
    }
    return "eth0"
}

// ExecEthtoolSpeed 执行ethtool命令获取接口速度
func (s *Scanner) ExecEthtoolSpeed() (int, error) {
    ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
    defer cancel()

    cmd := exec.CommandContext(ctx, "ethtool", s.InterfaceName)
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return 0, fmt.Errorf("failed to get stdout pipe: %w", err)
    }

    if err := cmd.Start(); err != nil {
        return 0, fmt.Errorf("failed to start ethtool: %w", err)
    }

    scanner := bufio.NewScanner(stdout)
    var speedMbps int

    // 扫描ethtool输出，查找速度信息
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if strings.HasPrefix(line, "Speed:") {
            fields := strings.Fields(line)
            if len(fields) < 2 {
                continue
            }

            speedStr := fields[1]
            speed, err := parseSpeed(speedStr)
            if err != nil {
                return 0, fmt.Errorf("parse speed failed: %w", err)
            }
            speedMbps = speed
            break
        }
    }

    if err := scanner.Err(); err != nil {
        return 0, fmt.Errorf("reading ethtool output failed: %w", err)
    }

    if err := cmd.Wait(); err != nil {
        return 0, fmt.Errorf("ethtool command failed: %w", err)
    }

    if speedMbps == 0 {
        return 0, errors.New("speed info not found in ethtool output")
    }
    return speedMbps, nil
}

// parseSpeed 解析速度字符串为整数（单位：Mbps）
func parseSpeed(speedStr string) (int, error) {
    speedStr = strings.ToLower(speedStr)
    var multiplier int

    // 根据单位确定倍率
    switch {
    case strings.HasSuffix(speedStr, "gb/s"):
        multiplier = 1000
        speedStr = strings.TrimSuffix(speedStr, "gb/s")
    case strings.HasSuffix(speedStr, "mb/s"):
        multiplier = 1
        speedStr = strings.TrimSuffix(speedStr, "mb/s")
    default:
        return 0, fmt.Errorf("unsupported speed unit in %s", speedStr)
    }

    var val int
    _, err := fmt.Sscanf(speedStr, "%d", &val)
    if err != nil {
        return 0, fmt.Errorf("failed to parse speed value: %w", err)
    }

    return val * multiplier, nil
}
