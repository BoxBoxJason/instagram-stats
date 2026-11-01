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
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/boxboxjason/instagram-stats/internal/config"
	"github.com/boxboxjason/instagram-stats/internal/export"
	"github.com/boxboxjason/instagram-stats/internal/report"
	"github.com/boxboxjason/instagram-stats/internal/stats"
)

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(os.Stderr, "instagram-stats:", err)
		}
		os.Exit(1)
	}
}

func run(args []string, stderr io.Writer) error {
	flags := flag.NewFlagSet("instagram-stats", flag.ContinueOnError)
	flags.SetOutput(stderr)
	configPath := flags.String("config", "", "path to the JSON configuration `file` (default: the default configuration)")
	verbose := flags.Bool("verbose", false, "log every file read and written")
	logPath := flags.String("log-file", "", "also write the logs to `file`")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: instagram-stats [flags] <source>...")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "A source is a conversation folder of an Instagram export, a folder of")
		fmt.Fprintln(stderr, "conversations (such as messages/inbox), or a message_<n>.json file.")
		fmt.Fprintln(stderr, "Several sources are analysed together.")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Flags:")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() == 0 {
		flags.Usage()
		return errors.New("no source given")
	}

	var logFile io.Writer
	if *logPath != "" {
		file, err := os.OpenFile(*logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return err
		}
		defer file.Close()
		logFile = file
	}
	logger := newLogger(stderr, logFile, *verbose)
	// Syncing a terminal can fail on some systems, there is nothing to do about it.
	defer func() { _ = logger.Sync() }()

	cfg := config.Default()
	if *configPath != "" {
		var err error
		if cfg, err = config.Load(*configPath); err != nil {
			return err
		}
	}

	begin := time.Now()
	logger.Info("reading conversations", zap.Strings("sources", flags.Args()))
	conversations, err := export.Load(flags.Args(), logger)
	if err != nil {
		return err
	}
	if len(conversations) == 0 {
		return fmt.Errorf("no conversation file found in %v", flags.Args())
	}

	statistics := stats.New(time.Local, cfg.WordsToSearchFor)
	messages := 0
	for _, conversation := range conversations {
		statistics.Add(conversation)
		messages += len(conversation.Messages)
	}
	logger.Info("read conversations", zap.Int("files", len(conversations)), zap.Int("messages", messages),
		zap.Int("participants", len(statistics.Participants())))

	if err := os.MkdirAll(cfg.ResultDir, 0o755); err != nil {
		return fmt.Errorf("creating the result folder: %w", err)
	}
	options := report.Options{
		SearchedWords: len(cfg.WordsToSearchFor) > 0,
		TextRankings:  cfg.OutputTextRanking,
	}
	if err := report.Write(cfg.ResultDir, statistics, options, logger); err != nil {
		return err
	}
	logger.Info("results written", zap.String("folder", cfg.ResultDir),
		zap.Duration("runtime", time.Since(begin).Round(time.Millisecond)))
	return nil
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
