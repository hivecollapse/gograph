package cmd

import (
	"fmt"
	"gograph/internal/log"
	"gograph/internal/schema"
	"gograph/internal/util"

	"github.com/spf13/cobra"
)

// genvarCmd represents the genvar command
var genvarCmd = &cobra.Command{
	Use:   "genvar",
	Short: "Generate GraphQL query variables",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatalln("Specify operation")
		}
		operationName := args[0]

		log.Println("gen called with operation", operationName)

		userSchema, err := schema.LoadSchemaFromGlob(SchemaPath)
		if err != nil {
			log.Fatalln("Unable to load schema")
		}

		operation := userSchema.FindOperationByName(operationName)

		if operation == nil {
			log.Fatalln("operation not found")
		}

		log.Println("operation:", operation.String())

		variables := operation.Variables()
		fmt.Println(util.PrettyPrint(variables))

	},
}

func init() {
	queryCmd.AddCommand(genvarCmd)
}
