package cmd

import (
	"fmt"
	"gograph/internal/log"
	"gograph/internal/schema"

	"github.com/spf13/cobra"
)

var (
	validate bool
	depth    int
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate GraphQL query",
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

		queryString := operation.QueryString(&schema.QuerySelectorOptions{
			IgnoreUnderscored: true,
			MaxDepth:          uint8(depth),
		})

		fmt.Println(queryString.Text)

		if validate {
			err = userSchema.Validate(queryString.Text)
			if err != nil {
				log.Fatalln(err)
			}
		}

	},
}

func init() {
	queryCmd.AddCommand(genCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// genCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	genCmd.Flags().BoolVarP(&validate, "validate", "v", false, "Validate the generated query")
	genCmd.Flags().IntVarP(&depth, "depth", "d", 3, "Depth for query generation")
}
