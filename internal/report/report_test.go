package report

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/boxboxjason/instagram-stats/internal/export"
	"github.com/boxboxjason/instagram-stats/internal/stats"
)

var discard = zap.NewNop()

func sampleStats() *stats.Stats {
	s := stats.New(time.UTC, []string{"lol"})
	var messages []export.Message
	start := time.Date(2023, 11, 20, 0, 0, 0, 0, time.UTC)
	for i := range 400 {
		sender := []string{"Alice", "Bob", "Carol 😀"}[i%3]
		messages = append(messages, export.Message{
			SenderName:  export.Text(sender),
			TimestampMs: start.Add(time.Duration(i) * 7 * time.Hour).UnixMilli(),
			Content:     export.Text(fmt.Sprintf("lol message %d", i)),
			Reactions:   []export.Reaction{{Reaction: "❤", Actor: "Bob"}},
		})
	}
	s.Add(&export.Conversation{
		Participants: []export.Participant{{Name: "Alice"}, {Name: "Bob"}, {Name: "Carol 😀"}},
		Messages:     messages,
	})
	return s
}

func listDir(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	slices.Sort(names)
	return names
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		want    []string
	}{
		{
			name:    "every output",
			options: Options{SearchedWords: true, TextRankings: true},
			// No reels nor photos and videos: their pie charts are skipped.
			want: []string{
				"Activity by time and day.png", "Number of Messages.png", "Number of Reactions.png",
				"Number of messages per day.png", "Number of messages per month.png", "Number of messages.txt",
				"Number of photos and videos.txt", "Number of reactions.txt", "Number of reels.txt",
				"Number of words found.png", "Number of words found.txt", "Vocabulary richness.txt", "Vocabulary.png",
			},
		},
		{
			name:    "graphs only",
			options: Options{},
			want: []string{
				"Activity by time and day.png", "Number of Messages.png", "Number of Reactions.png",
				"Number of messages per day.png", "Number of messages per month.png", "Vocabulary.png",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := Write(dir, sampleStats(), test.options, discard); err != nil {
				t.Fatal(err)
			}
			if got := listDir(t, dir); !slices.Equal(got, test.want) {
				t.Errorf("files = %q, want %q", got, test.want)
			}
			for _, name := range listDir(t, dir) {
				if filepath.Ext(name) != ".png" {
					continue
				}
				file, err := os.Open(filepath.Join(dir, name))
				if err != nil {
					t.Fatal(err)
				}
				config, err := png.DecodeConfig(file)
				file.Close()
				if err != nil {
					t.Errorf("%s: %v", name, err)
				} else if config.Width != 1600 || config.Height != 900 {
					t.Errorf("%s: %d×%d image, want 1600×900", name, config.Width, config.Height)
				}
			}
		})
	}
}

func TestWriteRanking(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, sampleStats(), Options{TextRankings: true}, discard); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "Number of reactions.txt"))
	if err != nil {
		t.Fatal(err)
	}
	want := "Number of reactions ranking (Total 400):\n1: 400 reactions: Bob\n2: 0 reactions: Alice, Carol 😀\n"
	if string(got) != want {
		t.Errorf("ranking:\n%s\nwant:\n%s", got, want)
	}
}

func TestWriteWithoutMessages(t *testing.T) {
	s := stats.New(time.UTC, nil)
	s.Add(&export.Conversation{Participants: []export.Participant{{Name: "Alice"}}})
	dir := t.TempDir()
	if err := Write(dir, s, Options{TextRankings: true}, discard); err != nil {
		t.Fatal(err)
	}
	for _, name := range listDir(t, dir) {
		if filepath.Ext(name) == ".png" {
			t.Errorf("wrote graph %s for statistics without messages", name)
		}
	}
}

func TestFiles(t *testing.T) {
	files := Files()
	if len(files) != 15 {
		t.Errorf("got %d files, want 15: %q", len(files), files)
	}
	seen := map[string]bool{}
	for _, file := range files {
		if seen[file] {
			t.Errorf("file %q listed twice", file)
		}
		seen[file] = true
	}
}

func TestPieSlices(t *testing.T) {
	var participants []*stats.Participant
	for i := range maxSlices + 5 {
		participants = append(participants, &stats.Participant{Name: fmt.Sprint("p", i), Messages: i})
	}
	pie := pieSlices(participants, func(p *stats.Participant) int { return p.Messages })
	// p0 sends nothing: 24 participants for 20 slices, the 5 smallest grouped.
	if len(pie) != maxSlices {
		t.Fatalf("got %d slices, want %d", len(pie), maxSlices)
	}
	if pie[0].label != "p24" || pie[0].value != 24 {
		t.Errorf("first slice = %+v, want p24", pie[0])
	}
	if others := pie[len(pie)-1]; others.label != "5 others" || others.value != 1+2+3+4+5 {
		t.Errorf("last slice = %+v, want 5 others with 15", others)
	}
}

func TestDrawable(t *testing.T) {
	tests := map[string]string{
		"Alice":         "Alice",
		"LB 🐆":          "LB",
		"𝕹𝖆𝖙𝖍𝖆𝖓 🦎":      "Nathan",
		"🔥🔥":            "(unnamed)",
		"Élodie Müller": "Élodie Müller",
	}
	for name, want := range tests {
		if got := drawable(name); got != want {
			t.Errorf("drawable(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestTimeTicks(t *testing.T) {
	days := func(from time.Time, n int) []time.Time {
		var starts []time.Time
		for i := range n {
			starts = append(starts, from.AddDate(0, 0, i))
		}
		return starts
	}
	labels := func(starts []time.Time) []string {
		var got []string
		for _, tick := range timeTicks(starts).Ticks(0, 0) {
			got = append(got, tick.Label)
		}
		return got
	}

	// Five years of days: a label every 4 months, on the months divisible by 4.
	long := labels(days(time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC), 5*365))
	if len(long) > maxTimeLabels || long[0] != "May 2020" || long[1] != "Sep 2020" {
		t.Errorf("labels of five years = %q", long)
	}

	// Ten days: a label per day.
	short := labels(days(time.Date(2024, 2, 25, 0, 0, 0, 0, time.UTC), 10))
	if len(short) != 10 || short[0] != "25 Feb 2024" {
		t.Errorf("labels of ten days = %q", short)
	}
}

func TestFormatCount(t *testing.T) {
	tests := map[int]string{0: "0", 999: "999", 1000: "1,000", 92567: "92,567", 1234567: "1,234,567", -1500: "-1,500"}
	for n, want := range tests {
		if got := formatCount(n); got != want {
			t.Errorf("formatCount(%d) = %q, want %q", n, got, want)
		}
	}
}
