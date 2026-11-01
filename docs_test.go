package main

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/boxboxjason/instagram-stats/internal/config"
	"github.com/boxboxjason/instagram-stats/internal/report"
)

// TestDocumentation fails when the documentation misses a configuration option or an
// output file. The documented ones are the first cell of a row of a table.
func TestDocumentation(t *testing.T) {
	configuration := readPage(t, "docs/content/docs/configuration.md")
	outputs := readPage(t, "docs/content/docs/outputs.md")

	options := reflect.TypeFor[config.Config]()
	for i := range options.NumField() {
		name, _, _ := strings.Cut(options.Field(i).Tag.Get("json"), ",")
		if !strings.Contains(configuration, "| `"+name+"` |") {
			t.Errorf("configuration option %q is not documented in docs/content/docs/configuration.md", name)
		}
	}
	for _, file := range report.Files() {
		if !strings.Contains(outputs, "| `"+file+"` |") {
			t.Errorf("output file %q is not documented in docs/content/docs/outputs.md", file)
		}
	}
}

func readPage(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
