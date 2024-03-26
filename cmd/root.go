package cmd

import (
	"github.com/spf13/cobra"
)

func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "deploy",
	Short: "App deployer",
	Long:  "App deployer is used to deploy your application to any kubernetes clusters as well as VMs via ansible",
}

func init() {
	// 将子命令添加到根命令
	rootCmd.AddCommand(kubeCmd)
	rootCmd.AddCommand(VMCmd)
}
