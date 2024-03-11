package cmd

import (
	"gograph/internal/log"
	"gograph/internal/schema"

	"github.com/spf13/cobra"
)

// var (
// 	ouputDump string
// )

// saveCmd represents the save command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump schema",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		schema, err := schema.LoadSchemaFromGlob(SchemaPath)
		if err != nil {
			log.Fatalln("Unable to load schema")
		}

		schema.Dump()
	},
}

func init() {
	schemaCmd.AddCommand(dumpCmd)
	// dumpCmd.Flags().StringVarP(&ouput, "output", "o", "", "Where to save the result")
}
