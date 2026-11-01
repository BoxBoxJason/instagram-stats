package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name, content string
		want          Config
	}{
		{
			name:    "empty object keeps the defaults",
			content: `{}`,
			want:    Default(),
		},
		{
			name:    "every option",
			content: `{"resultDir": "/tmp/out", "wordsToSearchFor": ["lol", "see you"], "outputTextRanking": false}`,
			want:    Config{ResultDir: "/tmp/out", WordsToSearchFor: []string{"lol", "see you"}, OutputTextRanking: false},
		},
		{
			name:    "some options",
			content: `{"wordsToSearchFor": ["lol"]}`,
			want:    Config{ResultDir: "Results", WordsToSearchFor: []string{"lol"}, OutputTextRanking: true},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Load(writeConfig(t, test.content))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestLoadErrors(t *testing.T) {
	tests := map[string]string{
		"unknown option": `{"resultDirectory": "out"}`,
		"wrong type":     `{"outputTextRanking": "yes"}`,
		"invalid JSON":   `{`,
		"empty folder":   `{"resultDir": ""}`,
	}
	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Load(writeConfig(t, content)); err == nil {
				t.Error("got no error")
			}
		})
	}
	if _, err := Load(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Error("missing file: got no error")
	}
}
