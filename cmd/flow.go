package cmd

import (
	"github.com/spf13/cobra"
)

// flowCmd represents the flow command
var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "Operates on flow",
	Long:  ``,
}

func init() {
	rootCmd.AddCommand(flowCmd)
}
