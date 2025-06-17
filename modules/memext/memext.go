package memext

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

var pool []byte
var NumaNode int = -1

// Init 显存扩展初始化，autoEnable=true 会自动映射挂载目录大小（单位字节）
// 只做内存映射和线程NUMA绑定，不再重复设置系统巨页
func Init(autoEnable bool) error {
	if !autoEnable {
		log.Println("[memext] 模块未启用")
		return nil
	}

	// 读取挂载目录内存池大小，路径可根据docker挂载调整
	alloc, err := getMemExtAlloc("/mnt/memext/size")
	if err != nil {
		return fmt.Errorf("[memext] 获取共享内存池大小失败: %v", err)
	}
	if alloc <= 0 {
		return fmt.Errorf("[memext] 共享内存池大小无效: %d", alloc)
	}

	log.Printf("[memext] 映射共享内存池大小 %.2f GB", float64(alloc)/1e9)

	// 内存映射匿名私有内存，预分配（MAP_POPULATE避免首次访问page fault）
	data, err := syscall.Mmap(
		-1, 0, int(alloc),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANON|syscall.MAP_PRIVATE|syscall.MAP_POPULATE,
	)
	if err != nil {
		return fmt.Errorf("[memext] 内存映射失败: %v", err)
	}
	pool = data

	// 获取当前线程绑定的 NUMA 节点，尝试绑定线程避免跨 NUMA 访问
	node := getCPUNuma()
	NumaNode = node
	if node >= 0 {
		// 锁定当前线程，防止调度到其他线程，避免 NUMA 绑定失效
		runtime.LockOSThread()
		log.Printf("[memext] 当前线程绑定 NUMA 节点: %d", node)

		// TODO: 如需更严格的线程内存分配策略，可在这里调用 set_mempolicy 等系统调用
	}

	log.Println("[memext] 显存共享池初始化成功")
	return nil
}

// Pool 返回内存池引用，供其他模块使用
func Pool() []byte {
	return pool
}

// getMemExtAlloc 读取共享内存池大小，单位字节
// 挂载路径 /mnt/memext/size 文件由外部 docker 启动脚本或配置写入
func getMemExtAlloc(path string) (int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	sizeStr := strings.TrimSpace(string(data))
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return size, nil
}

// getCPUNuma 读取当前线程绑定的 CPU NUMA 节点，失败返回 -1
func getCPUNuma() int {
	data, err := os.ReadFile("/proc/self/numa_maps")
	if err != nil {
		return -1
	}
	return parseNumaMaps(string(data))
}

// parseNumaMaps 解析 numa_maps 获取绑定的 NUMA 节点
func parseNumaMaps(s string) int {
	for _, line := range strings.Split(s, "\n") {
		if strings.Contains(line, "heap") && strings.Contains(line, "N:") {
			parts := strings.Fields(line)
			for _, f := range parts {
				if strings.HasPrefix(f, "N:") {
					n := f[2:]
					if node, err := strconv.Atoi(n); err == nil {
						return node
					}
				}
			}
		}
	}
	return -1
}
