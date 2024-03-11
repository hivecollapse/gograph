package cmd

import (
	"github.com/spf13/cobra"
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "List queries in a schema",
	Long:  ``,
	// Run: func(cmd *cobra.Command, args []string) {
	// },
}

func init() {
	schemaCmd.AddCommand(queryCmd)
}
