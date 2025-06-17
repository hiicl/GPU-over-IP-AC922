package main

import (
    "fmt"
    "os"

    "juice-server-cli/cmd/juice/config"
    "juice-server-cli/cmd/juice/utils"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "juice",
    Short: "Juice CLI 管理工具",
    Long:  `统一控制 Juice Server 的拓扑初始化、服务启动与模块配置`,
}

func main() {
    rootCmd.AddCommand(initCmd)
    rootCmd.AddCommand(startCmd)
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}