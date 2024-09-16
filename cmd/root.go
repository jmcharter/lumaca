package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/jmcharter/lumaca/config"
	"github.com/spf13/cobra"
)

var cfg config.Config
var initialised bool = false

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lumaca",
	Short: "Lumaca, the simple static site generator.",
	Long:  `Lumaca is a fast and simple static site generator that transforms your markdown files into a static website.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
//
//	func Execute() {
//		if initialised {
//			err := rootCmd.Execute()
//			if err != nil {
//				os.Exit(1)
//			}
//		}
//	}
func Execute() {
	if !initFileExists() {
		// Only register init command if init file does not exist
		rootCmd.AddCommand(initCmd)
	} else {
		rootCmd.AddCommand(initCmd)
		rootCmd.AddCommand(newCmd)
		rootCmd.AddCommand(buildCmd)
		rootCmd.AddCommand(serveCmd)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initFileExists() bool {
	_, err := os.Stat("config.toml")
	return !os.IsNotExist(err)
}
func init() {
	// Try to initialize configuration, but allow for it to not exist
	initConfig()

}

func initConfig() {
	var err error
	cfg, err = config.InitConfig()
	if err != nil {
		if errors.Is(err, config.DecodeFileError) {
			// Configuration file does not exist, log an info message and continue
			log.Println("Configuration file not found, please run init")
		} else {
			// Some other error occurred during config initialization
			log.Fatal(err)
		}
	}
	initialised = true
}
