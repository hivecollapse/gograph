package cmd

import (
	"gograph/internal/global"

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

	flowCmd.PersistentFlags().BoolVarP(&global.Verbose, "verbose", "v", false, "Verbose")
	flowCmd.PersistentFlags().BoolVarP(&global.Debug, "debug", "d", false, "Debug")
}
