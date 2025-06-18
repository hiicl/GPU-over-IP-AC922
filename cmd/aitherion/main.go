package main

import (
    "fmt"
    "os"

    "github.com/hiicl/GPU-over-IP-AC922/cmd/aitherion/config"
    "github.com/hiicl/GPU-over-IP-AC922/cmd/aitherion/utils"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "aitherion",
    Short: "aitherion CLI 管理工具",
    Long:  `统一控制 aitherion Server 的拓扑初始化、服务启动与模块配置`,
}

func main() {
    rootCmd.AddCommand(initCmd)
    rootCmd.AddCommand(startCmd)
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
