package cli

import (
	"context"
	"fmt"
	"gate-limiter/config/settings"
	"gate-limiter/internal/admin"
	"gate-limiter/internal/app"
	"gate-limiter/internal/metrics"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var daemonMode bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the rate limiter server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

func init() {
	runCmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "Run server in background (daemon mode)")
	rootCmd.AddCommand(runCmd)
}

func runServer() error {
	cfgPath := resolveConfigPath()

	if daemonMode {
		return runDaemon(cfgPath)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	limitHandler, config, err := app.InitRateLimitHandler(ctx, cfgPath)
	if err != nil {
		return fmt.Errorf("error initializing rate limiter handler: %w", err)
	}

	server := initMainServer(config, limitHandler)
	adminServer := initAdminServer(config)

	go func() {
		log.Printf("Admin page: http://localhost:%d\n", config.RateLimiter.AdminPort)
		if err := adminServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("admin server error: %v", err)
		}
	}()

	waitForShutdown(ctx, server, adminServer)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	log.Println("server stopped")
	return nil
}

func initMainServer(config *settings.RootRateLimiterConfig, handler http.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", handler)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.RateLimiter.Port),
		Handler: metrics.WithMetrics(mux),
	}
}

func initAdminServer(config *settings.RootRateLimiterConfig) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", admin.NewStatusHandler(config))

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.RateLimiter.AdminPort),
		Handler: mux,
	}
}

func waitForShutdown(ctx context.Context, servers ...*http.Server) {
	go func() {
		<-ctx.Done()
		log.Println("shutting down server...")
		for _, s := range servers {
			if err := s.Shutdown(context.Background()); err != nil {
				log.Printf("server shutdown error (%s): %v", s.Addr, err)
			}
		}
	}()
}

func runDaemon(cfgPath string) error {
	config, err := settings.ParseAndValidateConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error resolving executable path: %w", err)
	}

	cmd := exec.Command(exe, "run", "-c", cfgPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	logFile, err := os.OpenFile("gl.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening log file: %w", err)
	}
	defer logFile.Close()

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting daemon: %w", err)
	}

	pid := cmd.Process.Pid
	if err := os.WriteFile("gl.pid", []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("error writing pid file: %w", err)
	}

	fmt.Printf("gate-limiter started (PID: %d)\n", pid)
	fmt.Printf("  server    : http://localhost:%d\n", config.RateLimiter.Port)
	fmt.Printf("  admin page: http://localhost:%d\n", config.RateLimiter.AdminPort)
	fmt.Printf("  log       : gl.log\n")
	fmt.Printf("  pid       : gl.pid\n")

	return nil
}
