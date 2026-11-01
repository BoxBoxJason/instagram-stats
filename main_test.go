package main

import (
	"bytes"
	"github.com/boxboxjason/instagram-stats/internal/config"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testExport = "internal/export/testdata/messages"

func TestRun(t *testing.T) {
	dir := t.TempDir()
	results := filepath.Join(dir, "results", "nested")
	configPath := filepath.Join(dir, "config.json")
	configJSON := `{"resultDir": "` + filepath.ToSlash(results) + `", "wordsToSearchFor": ["drôle"]}`
	if err := os.WriteFile(configPath, []byte(configJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(dir, "run.log")

	var stderr bytes.Buffer
	if err := run([]string{"-config", configPath, "-log-file", logPath, testExport}, &stderr); err != nil {
		t.Fatalf("run: %v\n%s", err, stderr.String())
	}

	// The result folder is created, and the searched word was found in the repaired text.
	words, err := os.ReadFile(filepath.Join(results, "Number of words found.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(words), "1: 1 words found: Bob") {
		t.Errorf("words found ranking:\n%s", words)
	}
	logs, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logs), "results written") {
		t.Errorf("log file misses the end of the run:\n%s", logs)
	}
}

func TestRunErrors(t *testing.T) {
	tests := map[string][]string{
		"no source":            {},
		"missing source":       {"testdata/missing"},
		"missing config":       {"-config", "testdata/missing.json", testExport},
		"not a conversation":   {testExport + "/secret_conversations.json"},
		"no conversation file": {"internal/config"},
		"unknown flag":         {"-unknown", testExport},
	}
	for name, args := range tests {
		t.Run(name, func(t *testing.T) {
			if err := run(args, &bytes.Buffer{}); err == nil {
				t.Error("got no error")
			}
		})
	}
}

func TestWriteReportCreateError(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.ResultDir = filepath.Join(blocker, "results")

	if err := writeReport(cfg, nil, zap.NewNop()); err == nil {
		t.Fatal("expected an error when the result folder cannot be created")
	}
}

func TestRunFailures(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(cfgPath, []byte(`{"resultDir": "`+filepath.ToSlash(dir)+`/out"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := map[string][]string{
		"bad log file": {"-log-file", filepath.Join(dir, "missing", "x.log"), testExport},
		"bad source":   {"-config", cfgPath, filepath.Join(dir, "nope")},
	}
	for name, args := range tests {
		t.Run(name, func(t *testing.T) {
			var stderr bytes.Buffer
			if err := run(args, &stderr); err == nil {
				t.Fatal("expected an error")
			}
		})
	}

	var stderr bytes.Buffer
	if err := run([]string{"-verbose", "-config", cfgPath, testExport}, &stderr); err != nil {
		t.Fatalf("verbose run: %v", err)
	}
}
