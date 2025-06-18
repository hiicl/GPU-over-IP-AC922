package config

// CLIConfig 定义 aitherion CLI 启动参数
// 注意：config.json 将不再使用，由 CLI 参数或环境变量动态指定
type CLIConfig struct {
	// MemExt 模块参数
	EnableMemExt       bool    // 是否启用 memext 模块
	MemExtSizeMB       int     // 分配的共享显存池大小（MB），如为 0，则使用比率
	MemExtRatio        float64 // 使用内存占空闲比率（如 0.9 表示使用空闲内存的 90%）
	EnableHugePages    bool    // 是否启用临时 hugepages 设置
	DisableNUMABinding bool    // 是否禁止 NUMA 绑定（默认绑定）

	// NetBalance 模块参数
	EnableNetBalance bool // 是否启用网卡 NUMA 负载均衡

	// Docker 启动参数
	GRPCBasePort   int    // gRPC 起始端口
	ImageName      string // 容器镜像名
	Tag            string // 镜像 tag
	DryRun         bool   // 仅输出 docker 命令不执行
	
	// 新增字段
	NumNUMA         int // 启动的NUMA容器数量
	CurrentNUMAIndex int // 当前NUMA索引（用于容器启动时指定）
}

// 默认值（也可扩展 defaults.go 加载）
func DefaultCLIConfig() CLIConfig {
	return CLIConfig{
		EnableMemExt:       false,
		MemExtSizeMB:       0,         // 若为 0，则使用 MemExtRatio
		MemExtRatio:        0.9,       // 默认使用 90% 的空闲内存
		EnableHugePages:    true,      // 默认启用临时 hugepages
		DisableNUMABinding: false,     // 默认绑定 NUMA

		EnableNetBalance: false,
		GRPCBasePort:     50051,
		ImageName:        "aitherion-server",
		Tag:              "latest",
		DryRun:           false,
		
		// 新增字段默认值
		NumNUMA:         0,
		CurrentNUMAIndex: 0,
	}
}
