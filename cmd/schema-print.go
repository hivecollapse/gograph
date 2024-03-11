package cmd

import (
	"fmt"
	"gograph/internal/log"
	"gograph/internal/schema"

	"github.com/spf13/cobra"
)

var (
	ouput string
)

// saveCmd represents the save command
var saveCmd = &cobra.Command{
	Use:   "print",
	Short: "Print schema",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		schema, err := schema.LoadSchemaFromGlob(SchemaPath)
		if err != nil {
			log.Fatalln("Unable to load schema")
		}

		if len(ouput) > 0 {
			err := schema.Save(ouput)
			if err != nil {
				log.Fatal("Unable to save schema", err)
			}
			log.Println("Schema saved", ouput)
		} else {
			fmt.Println(schema.PrintIndentString())
		}
	},
}

func init() {
	schemaCmd.AddCommand(saveCmd)
	saveCmd.Flags().StringVarP(&ouput, "output", "o", "", "Where to save the result")
}
