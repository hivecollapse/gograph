package cmd

import (
	"github.com/spf13/cobra"
)

var (
	SchemaPath string
)

// schemaCmd represents the schema command
var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "A brief description of your command",
	Long:  `A longer description that spans multiple `,
}

func init() {
	rootCmd.AddCommand(schemaCmd)

	schemaCmd.MarkPersistentFlagRequired("path")
	schemaCmd.PersistentFlags().StringVarP(&SchemaPath, "path", "p", "", "Glob path to the graphql schema files (required)")
}
