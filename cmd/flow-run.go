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
	runCmd.Flags().StringVarP(&until, "until", "u", "", "Run until the specified step")
}
