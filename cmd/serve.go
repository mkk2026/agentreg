package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/mkk2026/agentreg/internal/registry"
	"github.com/mkk2026/agentreg/internal/server"
	"github.com/mkk2026/agentreg/internal/verify"
	"github.com/spf13/cobra"
)

var (
	servePort     int
	serveStore    string
	serveInterval time.Duration
	serveTimeout  time.Duration
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the agentreg registry daemon",
	RunE: func(_ *cobra.Command, _ []string) error {
		logger := log.New(os.Stderr, "agentreg ", log.LstdFlags)

		store, err := registry.NewMemoryStore(serveStore)
		if err != nil {
			return fmt.Errorf("open store: %w", err)
		}
		srv := server.New(store, verify.NewHealthVerifier(serveTimeout), logger)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		go srv.StartHeartbeat(ctx, serveInterval)

		httpServer := &http.Server{Addr: fmt.Sprintf(":%d", servePort), Handler: srv.Handler()}
		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = httpServer.Shutdown(shutdownCtx)
		}()

		loaded, _ := store.List()
		printServeBanner(servePort, len(loaded), serveInterval)
		logger.Printf("listening on :%d (store=%q, heartbeat=%s)", servePort, serveStore, serveInterval)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		logger.Print("shut down cleanly")
		return nil
	},
}

func printServeBanner(port, agents int, interval time.Duration) {
	fmt.Printf("%s %s\n",
		paint("agentreg", ansiBold+ansiCyan),
		paint("· DNS for AI agents", ansiDim))
	fmt.Printf("%s http://localhost:%d  %s\n",
		paint("→", ansiGreen),
		port,
		paint(fmt.Sprintf("%d agent%s loaded · heartbeat %s", agents, plural(agents), interval), ansiDim))
}

func defaultStorePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "agentreg.json"
	}
	return filepath.Join(home, ".agentreg", "registry.json")
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "port to listen on")
	serveCmd.Flags().StringVar(&serveStore, "store", defaultStorePath(), "path to the JSON persistence file")
	serveCmd.Flags().DurationVar(&serveInterval, "heartbeat-interval", 15*time.Second, "how often to health-check agents")
	serveCmd.Flags().DurationVar(&serveTimeout, "probe-timeout", 3*time.Second, "per-agent health probe timeout")
	rootCmd.AddCommand(serveCmd)
}
