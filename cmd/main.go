package main

import (
	"fmt"
	"instagram-stats/internal/config"
	"instagram-stats/internal/stats"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	configPath string
	verbose    bool
	outputDir  string
	display    bool
	version    = "dev"
)

func main() {
	rootCmd := &cobra.Command{
		Version: version,
		Use:     "instagram-stats",
		Short:   "Fetch and analyze Instagram statistics",
		Run: func(cmd *cobra.Command, args []string) {
			startTime := time.Now()
			setupZapLogger(verbose)
			zap.L().Info("Starting instagram-stats", zap.Time("start_time", startTime))

			zap.L().Debug("Config path", zap.String("config", configPath))
			zap.L().Debug("Output directory", zap.String("output_dir", outputDir))
			zap.L().Debug("Display enabled", zap.Bool("display", display))

			config, err := config.GetConfig(configPath)
			if err != nil {
				zap.L().Error("Failed to load configuration", zap.Error(err))
				os.Exit(1)
			}
			zap.L().Debug("Loaded configuration", zap.Any("config", config))

			// Collect messages
			err = stats.ProcessConversations(config.CollectConfig, config.OutputConfig)
			if err != nil {
				zap.L().Error("Failed to process conversations", zap.Error(err))
				os.Exit(1)
			}

			elapsed := time.Since(startTime)
			zap.L().Info("instagram-stats finished successfully", zap.Duration("elapsed_time", elapsed))
		},
	}

	rootCmd.Flags().StringVarP(&configPath, "config-path", "c", "instagram-stats.json", "Path to configuration file")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "stats", "Directory for process output")
	rootCmd.Flags().BoolVarP(&display, "display", "d", false, "Open results in browser after processing")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setupZapLogger(verbose bool) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if verbose {
		config.Level.SetLevel(zapcore.DebugLevel)
	} else {
		config.Level.SetLevel(zapcore.InfoLevel)
	}

	logger, err := config.Build()
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}

	zap.ReplaceGlobals(logger)
}
