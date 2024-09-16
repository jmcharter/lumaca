package cmd

import (
	"github.com/jmcharter/lumaca/builder"
	"github.com/spf13/cobra"
)

var cfgAuthor string
var cfgTitle string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new lumaca project",
	Long:  `Initialize a new lumaca project, creating the default directory structure and guiding you through the creation of a config file if one is not already detected.`,
	Run: func(cmd *cobra.Command, args []string) {
		builder.Initialise(cfgAuthor, cfgTitle)
	},
}

func init() {
	initCmd.Flags().StringVarP(&cfgAuthor, "author", "a", "", "Name of the primary blog author (optional)")
	initCmd.Flags().StringVarP(&cfgTitle, "title", "t", "", "Title of the blog (optional)")
}
