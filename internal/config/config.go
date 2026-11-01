// Package config reads the JSON configuration of instagram-stats.
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

// Config is the configuration of a run. Every option of the file is optional.
type Config struct {
	// ResultDir is the folder the graphs and rankings are written to. It is created
	// when missing.
	ResultDir string `json:"resultDir"`
	// WordsToSearchFor are words or phrases counted in the messages, regardless of case.
	// The empty ones are left out.
	WordsToSearchFor []string `json:"wordsToSearchFor"`
	// OutputTextRanking also writes the text rankings, next to the graphs.
	OutputTextRanking bool `json:"outputTextRanking"`
}

// Default returns the configuration used when no file is given, which also holds the
// default value of the options a file leaves out.
func Default() Config {
	return Config{
		ResultDir:         "Results",
		WordsToSearchFor:  []string{},
		OutputTextRanking: true,
	}
}

// Load reads the configuration file at path. An option missing from the file keeps
// its default value, and an unknown option is an error, to catch typos.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return Config{}, fmt.Errorf("reading the configuration file: %w", err)
	}

	cfg := Default()
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("invalid configuration file %s: %w", path, err)
	}

	// An empty word would be found in every message.
	cfg.WordsToSearchFor = slices.DeleteFunc(cfg.WordsToSearchFor, func(word string) bool { return word == "" })

	if cfg.ResultDir == "" {
		return Config{}, fmt.Errorf("invalid configuration file %s: resultDir is empty", path)
	}

	return cfg, nil
}
