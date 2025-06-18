package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hiicl/GPU-over-IP-AC922/cmd/aitherion/config"
	"github.com/hiicl/GPU-over-IP-AC922/cmd/aitherion/utils"

	"github.com/spf13/cobra"
)

var cfg config.CLIConfig

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "启动所有 NUMA 实例容器，默认绑定 NUMA 资源并自动分配端口",
	Run: func(cmd *cobra.Command, args []string) {
		// 先自动扫描当前系统 NUMA 数量（即有多少 numaX_gpus.txt 文件）
		numaFiles, err := filepath.Glob("/var/lib/aitherion/topology/numa[0-9]*_gpus.txt")
		if err != nil {
			fmt.Printf("扫描 NUMA 文件失败: %v\n", err)
			os.Exit(1)
		}
		if len(numaFiles) == 0 {
			fmt.Println("未检测到 NUMA GPU 拓扑文件，退出")
			os.Exit(1)
		}

		// 允许通过命令行参数限制启动 NUMA 数量，默认为全部
		startNum := len(numaFiles)
		if cfg.NumNUMA > 0 && cfg.NumNUMA < startNum {
			startNum = cfg.NumNUMA
		}

		fmt.Printf("检测到 %d 个 NUMA 节点，将启动 %d 个 NUMA 容器\n", len(numaFiles), startNum)

		// 确保默认启用 NUMA 绑定（除非显式禁用）
		if !cmd.Flags().Changed("no-numa-bind") {
			cfg.DisableNUMABinding = false
		}

		// 根据 NUMA 数量和起始端口自动分配 GRPC 端口
		for i := 0; i < startNum; i++ {
			cfg.GRPCBasePort = cfg.GRPCBasePort + i
			cfg.CurrentNUMAIndex = i // 新增字段，便于 docker.go 使用（可选）

			fmt.Printf("[start] 启动 NUMA %d 容器，GRPC 端口: %d\n", i, cfg.GRPCBasePort)
			if err := utils.StartContainerForNUMA(cfg, i); err != nil {
				fmt.Printf("启动 NUMA %d 容器失败: %v\n", i, err)
				os.Exit(1)
			}
		}

		fmt.Println("[start] 所有 NUMA 容器启动成功")
	},
}

func init() {
	// Memext 相关参数
	startCmd.Flags().BoolVar(&cfg.EnableMemExt, "memext", false, "启用 memext 模块")
	startCmd.Flags().IntVar(&cfg.MemExtSizeMB, "memext-size-mb", 0, "memext 分配内存大小 (MB)，0 表示自动计算")
	startCmd.Flags().Float64Var(&cfg.MemExtRatio, "memext-ratio", 0.9, "memext 使用空闲内存比例 (默认0.9)")
	startCmd.Flags().BoolVar(&cfg.DisableNUMABinding, "no-numa-bind", false, "禁用 NUMA 绑定（默认启用）")
	startCmd.Flags().BoolVar(&cfg.EnableHugePages, "hugepages", true, "启用临时巨页 (默认启用)")

	// 网卡调度
	startCmd.Flags().BoolVar(&cfg.EnableNetBalance, "netbalance", false, "启用网卡负载均衡模块")

	// 容器相关参数
	startCmd.Flags().IntVar(&cfg.GRPCBasePort, "port", 50051, "GRPC 起始端口")
	startCmd.Flags().StringVar(&cfg.ImageName, "image", "aitherion-server", "容器镜像名称")
	startCmd.Flags().StringVar(&cfg.Tag, "tag", "latest", "镜像 tag")
	startCmd.Flags().BoolVar(&cfg.DryRun, "dry-run", false, "仅打印 Docker 命令，不执行")

	// 新增：允许用户指定启动几个 NUMA 容器
	startCmd.Flags().IntVar(&cfg.NumNUMA, "num-numa", 0, "启动 NUMA 容器数量，0 表示全部")
}
