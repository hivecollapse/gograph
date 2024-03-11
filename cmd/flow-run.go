/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"gograph/internal/flow"
	"gograph/internal/log"
	"strings"

	"github.com/spf13/cobra"
)

var (
	until string
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a flow",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatalln("Specify flow file")
		}

		finalSummary := new(strings.Builder)

		for _, file := range args {
			flowDef, err := flow.LoadFlowDefinitionFile(file)
			if err != nil {
				log.Fatalln(err)
			}

			flowRunner := flow.NewFlowRunner(flowDef)

			options := &flow.RunOption{
				BreakAfterStepNamed: until,
			}
			flowRunner.Run(options)

			// Print the result
			localSummary := new(strings.Builder)
			flowRunner.Summarize(localSummary)
			log.Out(localSummary)

			// Save for final result
			finalSummary.WriteString(localSummary.String())
		}

		log.Outln("\nAll flows have completed")
		log.Out(finalSummary)

	},
}

func init() {
	flowCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	runCmd.Flags().StringVarP(&until, "until", "u", "", "Run until the specified step")
}
