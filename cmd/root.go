package cmd

import (
	"log"
	"os"

	"github.com/jmcharter/lumaca/config"
	"github.com/spf13/cobra"
)

var cfg config.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lumaca",
	Short: "Lumaca, the simple static site generator.",
	Long:  `Lumaca is a fast and simple static site generator that transforms your markdown files into a static website.`,
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

	var err error
	cfg, err = config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}
}
