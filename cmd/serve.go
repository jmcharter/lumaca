package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var portFlag int

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve static content for local development",
	Long:  `Start a server and serve static content for local development, optionally building content as well`,
	Run: func(cmd *cobra.Command, args []string) {
		serveStaticContent(portFlag)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&portFlag, "port", "p", 8080, "Port to serve static content on")
}

func serveStaticContent(port int) {
	fs := http.FileServer(http.Dir(cfg.Directories.Dist))
	http.Handle("/", fs)
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Serving content on http://localhost%s\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
