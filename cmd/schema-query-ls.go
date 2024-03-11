package cmd

import (
	"fmt"
	"gograph/internal/log"
	"gograph/internal/schema"

	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List queries in a graphql document",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		s, err := schema.LoadSchemaFromGlob(SchemaPath)
		if err != nil {
			log.Fatalln("Unable to load schema")
		}
		operations := s.ListOperations(schema.Query, true)
		for _, o := range operations {
			fmt.Println("query: ", o.String())
		}

		operations = s.ListOperations(schema.Mutation, true)
		for _, o := range operations {
			fmt.Println("mutation:", o.String())
		}
	},
}

func init() {
	queryCmd.AddCommand(lsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
