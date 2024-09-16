package cmd

import (
	"github.com/jmcharter/lumaca/builder"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new lumaca project",
	Long:  `Initialize a new lumaca project, creating the default directory structure and guiding you through the creation of a config file if one is not already detected.`,
	Run: func(cmd *cobra.Command, args []string) {
		builder.Initialise()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
