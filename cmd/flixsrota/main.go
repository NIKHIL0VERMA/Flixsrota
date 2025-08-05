package main

import (
	"fmt"
	"os"

	"github.com/flixsrota/flixsrota/internal/config"
	"github.com/flixsrota/flixsrota/internal/core"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	configFile string
	logLevel   string

	// Version information
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "flixsrota",
		Short: "A high-performance video processing backend service",
		Long: `Flixsrota is a modular, high-performance video processing backend service 
that interfaces with FFmpeg using gRPC APIs. It supports pluggable queue and 
storage systems and is designed to run as a backend service/daemon.`,
		Version: fmt.Sprintf("%s (Built: %s, Commit: %s)", Version, BuildTime, GitCommit),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set log level
			switch logLevel {
			case "debug":
				logger, _ = zap.NewDevelopment()
			case "info":
				logger, _ = zap.NewProduction()
			case "warn":
				logger, _ = zap.NewProduction(zap.IncreaseLevel(zap.WarnLevel))
			case "error":
				logger, _ = zap.NewProduction(zap.IncreaseLevel(zap.ErrorLevel))
			}
		},
	}

	// Add persistent flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.flixsrota.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")

	// Add commands
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(serveCmd())

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Flixsrota configuration",
		Long:  "Interactive CLI wizard for config file generation and management",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file",
		Long:  "Run interactive wizard to create a new configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.RunWizard(configFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Configuration file created successfully!")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		Long:  "Check if the configuration file is valid",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load(configFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ Configuration file is valid!")
			fmt.Printf("📡 Server will run on: %s:%d\n", cfg.GRPC.Address, cfg.GRPC.Port)
			fmt.Printf("📋 Queue adapter: %s\n", cfg.Queue.Adapter)
			fmt.Printf("💾 Storage adapter: %s\n", cfg.Storage.Adapter)
			fmt.Printf("🎥 FFmpeg path: %s\n", cfg.FFmpeg.ExecutablePath)
			fmt.Printf("👷 Workers: %d-%d\n", cfg.Worker.MinWorkers, cfg.Worker.MaxWorkers)
		},
	})

	return cmd
}

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Flixsrota server",
		Long:  "Run the gRPC server and start job processing workers",
		Run: func(cmd *cobra.Command, args []string) {
			// Load configuration
			cfg, err := config.Load(configFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
				os.Exit(1)
			}

			// Create and start the server
			server := core.NewServer(cfg)
			if err := server.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
