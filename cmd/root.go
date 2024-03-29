package cmd

import (
	"gograph/internal/global"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gograph",
	Short: "A brief description of your application",
	Long:  `A longer description that spans multiple lines`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&global.Verbose, "verbose", "v", false, "Verbose")
	rootCmd.PersistentFlags().BoolVarP(&global.Debug, "debug", "d", false, "Debug")
}
