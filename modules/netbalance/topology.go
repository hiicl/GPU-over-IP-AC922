package netbalance

import (
    "bufio"
    "bytes"
    "fmt"
    "os/exec"
    "regexp"
    "strconv"
    "strings"
)

type NUMAGroup struct {
    NUMANode int
    GPUIDs   []int
    NetIfs   []string
}

// 读取 GPU 的 NUMA 节点，返回 map[gpuID]numaNode
func parseGPUNumaMapping() (map[int]int, error) {
    gpuNumaMap := make(map[int]int)
    for gpuID := 0; ; gpuID++ {
        path := fmt.Sprintf("/sys/class/drm/card%d/device/numa_node", gpuID)
        content, err := os.ReadFile(path)
        if err != nil {
            // 没找到文件或读错误，可能是GPU编号越界，停止读取
            break
        }
        numaStr := strings.TrimSpace(string(content))
        numa, err := strconv.Atoi(numaStr)
        if err != nil {
            numa = -1
        }
        gpuNumaMap[gpuID] = numa
    }
    if len(gpuNumaMap) == 0 {
        return nil, fmt.Errorf("no GPU NUMA info found in sysfs")
    }
    return gpuNumaMap, nil
}

// MapNUMATopology 聚合 GPU 和物理网卡，按 NUMA 节点分类
func MapNUMATopology() ([]NUMAGroup, error) {
    // 1. 获取所有物理网卡及其NUMA节点
    netIfs, err := scanner.DetectAllPhysicalInterfaces()
    if err != nil {
        return nil, fmt.Errorf("detectAllPhysicalInterfaces error: %v", err)
    }

    // 2. 获取GPU NUMA映射
    gpuNumaMap, err := parseGPUNumaMapping()
    if err != nil {
        return nil, err
    }

    // 3. 聚合，key为numa节点，value为NUMAGroup
    groups := make(map[int]*NUMAGroup)

    // 添加GPU到对应NUMA组
    for gpuID, numaNode := range gpuNumaMap {
        if numaNode < 0 {
            numaNode = 0 // 无NUMA归为0
        }
        group, ok := groups[numaNode]
        if !ok {
            group = &NUMAGroup{
                NUMANode: numaNode,
                GPUIDs:   []int{},
                NetIfs:   []string{},
            }
            groups[numaNode] = group
        }
        group.GPUIDs = append(group.GPUIDs, gpuID)
    }

    // 添加物理网卡到对应NUMA组
    for _, iface := range netIfs {
        numaNode := iface.NUMANode
        if numaNode < 0 {
            numaNode = 0
        }
        group, ok := groups[numaNode]
        if !ok {
            group = &NUMAGroup{
                NUMANode: numaNode,
                GPUIDs:   []int{},
                NetIfs:   []string{},
            }
            groups[numaNode] = group
        }
        group.NetIfs = append(group.NetIfs, iface.Name)
    }

    // 转换成切片返回，按NUMA节点排序（简单实现）
    result := []NUMAGroup{}
    for _, group := range groups {
        result = append(result, *group)
    }

    return result, nil
}

// WriteNUMAMappingFiles 写出每个NUMA节点对应的GPU和网卡信息到文件，方便run.sh脚本使用
func WriteNUMAMappingFiles(baseDir string, groups []NUMAGroup) error {
    if err := os.MkdirAll(baseDir, 0755); err != nil {
        return err
    }
    for _, g := range groups {
        // 写 GPU 列表文件
        gpuFile := filepath.Join(baseDir, fmt.Sprintf("numa%d_gpus.txt", g.NUMANode))
        gpuStrs := []string{}
        for _, gpuID := range g.GPUIDs {
            gpuStrs = append(gpuStrs, fmt.Sprintf("%d", gpuID))
        }
        if err := os.WriteFile(gpuFile, []byte(strings.Join(gpuStrs, "\n")), 0644); err != nil {
            return err
        }

        // 写 网卡列表文件
        netFile := filepath.Join(baseDir, fmt.Sprintf("numa%d_ifaces.txt", g.NUMANode))
        if err := os.WriteFile(netFile, []byte(strings.Join(g.NetIfs, "\n")), 0644); err != nil {
            return err
        }
    }
    return nil
}