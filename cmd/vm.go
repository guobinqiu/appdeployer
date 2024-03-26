package cmd

import (
	"github.com/spf13/cobra"
)

var VMCmd = &cobra.Command{
	Use:   "vm",
	Short: "Deploy to VM",
	Run: func(cmd *cobra.Command, args []string) {
		//TODO
	},
}
