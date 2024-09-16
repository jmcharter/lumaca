package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var portFlag int

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve static content for local development",
	Long:  `Start a server and serve static content for local development, optionally building content as well`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := serveStaticContent(portFlag); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	serveCmd.Flags().IntVarP(&portFlag, "port", "p", 8080, "Port to serve static content on")
}

func serveStaticContent(port int) error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.FileServer(http.Dir(cfg.Directories.Dist)),
	}
	go func() {
		fmt.Printf("Serving content on http://localhost%s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("\nShutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("Failed to gracefully shutdown server: %w", err)
	}
	fmt.Println("Server shutdown successfully.")

	return nil
}
