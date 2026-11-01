// Command instagram-stats creates statistics from Instagram conversations.
//
// It reads the messages of an Instagram data export (JSON format), and writes a set
// of graphs and text rankings: messages, reactions, reels, photos and videos and
// vocabulary of each participant, and activity per day, month and hour of the week.
//
// Usage:
//
//	instagram-stats [-config <file>] [-verbose] [-log-file <file>] <source>...
//
// A source is a conversation folder of the export, a folder of conversations such as
// its inbox folder, or a message_<n>.json file. Several sources are analysed together.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/boxboxjason/instagram-stats/internal/config"
	"github.com/boxboxjason/instagram-stats/internal/export"
	"github.com/boxboxjason/instagram-stats/internal/report"
	"github.com/boxboxjason/instagram-stats/internal/stats"
)

func main() {
	err := run(os.Args[1:], os.Stderr)
	if err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			_, _ = fmt.Fprintln(os.Stderr, "instagram-stats:", err)
		}

		os.Exit(1)
	}
}

const usageText = `Usage: instagram-stats [flags] <source>...

A source is a conversation folder of an Instagram export, a folder of
conversations (such as messages/inbox), or a message_<n>.json file.
Several sources are analysed together.

Flags:
`

const (
	logFileMode   = 0o600
	resultDirMode = 0o750
)

func run(args []string, stderr io.Writer) error {
	flags := flag.NewFlagSet("instagram-stats", flag.ContinueOnError)
	flags.SetOutput(stderr)
	configPath := flags.String("config", "", "path to the JSON configuration `file` (default: the default configuration)")
	verbose := flags.Bool("verbose", false, "log every file read and written")
	logPath := flags.String("log-file", "", "also write the logs to `file`")

	flags.Usage = func() {
		_, _ = fmt.Fprint(stderr, usageText)

		flags.PrintDefaults()
	}

	err := flags.Parse(args)
	if err != nil {
		return fmt.Errorf("parsing the flags: %w", err)
	}

	if flags.NArg() == 0 {
		flags.Usage()

		return errors.New("no source given")
	}

	var logFile io.Writer

	if *logPath != "" {
		file, openErr := os.OpenFile(filepath.Clean(*logPath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, logFileMode)
		if openErr != nil {
			return fmt.Errorf("opening the log file: %w", openErr)
		}

		defer func() { _ = file.Close() }()

		logFile = file
	}

	logger := newLogger(stderr, logFile, *verbose)
	// Syncing a terminal can fail on some systems, there is nothing to do about it.
	defer func() { _ = logger.Sync() }()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		return err
	}

	begin := time.Now()

	statistics, err := readStats(flags.Args(), cfg, logger)
	if err != nil {
		return err
	}

	err = writeReport(cfg, statistics, logger)
	if err != nil {
		return err
	}

	logger.Info("results written", zap.String("folder", cfg.ResultDir),
		zap.Duration("runtime", time.Since(begin).Round(time.Millisecond)))

	return nil
}

// loadConfig reads the configuration file at path, or returns the default configuration
// when path is empty.
func loadConfig(path string) (config.Config, error) {
	if path == "" {
		return config.Default(), nil
	}

	cfg, err := config.Load(path)
	if err != nil {
		return config.Config{}, fmt.Errorf("loading the configuration: %w", err)
	}

	return cfg, nil
}

// writeReport writes the graphs and rankings of statistics, in the result folder.
func writeReport(cfg config.Config, statistics *stats.Stats, logger *zap.Logger) error {
	err := os.MkdirAll(cfg.ResultDir, resultDirMode)
	if err != nil {
		return fmt.Errorf("creating the result folder: %w", err)
	}

	options := report.Options{
		SearchedWords: len(cfg.WordsToSearchFor) > 0,
		TextRankings:  cfg.OutputTextRanking,
	}

	err = report.Write(cfg.ResultDir, statistics, options, logger)
	if err != nil {
		return fmt.Errorf("writing the report: %w", err)
	}

	return nil
}

// readStats reads the conversations of sources and returns their statistics.
func readStats(sources []string, cfg config.Config, logger *zap.Logger) (*stats.Stats, error) {
	logger.Info("reading conversations", zap.Strings("sources", sources))

	conversations, err := export.Load(sources, logger)
	if err != nil {
		return nil, fmt.Errorf("reading the conversations: %w", err)
	}

	if len(conversations) == 0 {
		return nil, fmt.Errorf("no conversation file found in %v", sources)
	}

	statistics := stats.New(time.Now().Location(), cfg.WordsToSearchFor)
	messages := 0

	for _, conversation := range conversations {
		statistics.Add(conversation)
		messages += len(conversation.Messages)
	}

	logger.Info("read conversations", zap.Int("files", len(conversations)), zap.Int("messages", messages),
		zap.Int("participants", len(statistics.Participants())))

	return statistics, nil
}

// newLogger returns a logger writing to console in a human readable format and, when
// file is not nil, to file in JSON lines. Debug messages are only logged when verbose.
func newLogger(console, file io.Writer, verbose bool) *zap.Logger {
	level := zapcore.InfoLevel
	if verbose {
		level = zapcore.DebugLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	consoleConfig := encoderConfig
	consoleConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	cores := []zapcore.Core{
		zapcore.NewCore(zapcore.NewConsoleEncoder(consoleConfig), zapcore.AddSync(console), level),
	}
	if file != nil {
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(file), level))
	}

	return zap.New(zapcore.NewTee(cores...))
}
