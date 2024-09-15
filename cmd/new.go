package cmd

import (
	"github.com/jmcharter/lumaca/builder"
	"github.com/spf13/cobra"
)

var title string
var author string
var draft bool

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new page with basic metadata included.",
	Long:  `Create a new file representing a page or blog post. Metadata will be added to the top, populating information from system and config data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		builder.New(cfg, title, author, draft)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringVarP(&title, "title", "t", "", "Title for the new page (required)")
	newCmd.Flags().StringVarP(&author, "author", "a", "", "Author of the new page (optional)")
	newCmd.Flags().BoolVarP(&draft, "draft", "d", false, "Whether the new post is a draft. Defaults to false. (optional)")

	newCmd.MarkFlagRequired("title")
}
