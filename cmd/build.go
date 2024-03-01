package cmd

import (
	"github.com/jmcharter/lumaca/builder"
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds and compiles the site's source files into static files ready for deployment.",
	Long:  `Compiles all source content, templates, and static assets into a complete set of optimized, static HTML and CSS and renders an RSS feed.`,
	Run: func(cmd *cobra.Command, args []string) {
		builder.Build()
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().Bool("watch", false, "NOT CURRENTLY IMPLEMENTED - Continuously watch source files for changes and rebuild automitically when changes are detected.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
