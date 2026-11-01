// Package report writes the graphs and text rankings of the statistics.
package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"gonum.org/v1/plot/vg/draw"

	"github.com/boxboxjason/instagram-stats/internal/stats"
)

// count is a count per participant, drawn as a pie chart and written as a ranking.
type count struct {
	// chart is the name of the pie chart file, without its extension.
	chart string
	// title is the title of the pie chart.
	title string
	// ranking is the name of the ranking file, without its extension.
	ranking string
	// rankingTitle is the first line of the ranking file.
	rankingTitle string
	// unit is what is counted, such as "messages".
	unit string
	// value returns the count of a participant.
	value func(*stats.Participant) int
	// total returns the total of the counts, their sum when nil.
	total func(*stats.Stats) int
	// searchedWords is set for the count that only exists when words are searched.
	searchedWords bool
}

var counts = []count{
	{
		chart: "Number of words found", title: "Number of words found",
		ranking: "Number of words found", rankingTitle: "Number of words found",
		unit:          "words found",
		value:         func(p *stats.Participant) int { return p.WordsFound },
		searchedWords: true,
	},
	{
		chart: "Vocabulary", title: "Vocabulary richness (words)",
		ranking: "Vocabulary richness", rankingTitle: "Vocabulary richness ranking",
		unit:  "different words",
		value: (*stats.Participant).Vocabulary,
		total: (*stats.Stats).VocabularySize,
	},
	{
		chart: "Number of Messages", title: "Number of Messages",
		ranking: "Number of messages", rankingTitle: "Number of messages ranking",
		unit:  "messages",
		value: func(p *stats.Participant) int { return p.Messages },
	},
	{
		chart: "Number of Reactions", title: "Reactions to messages",
		ranking: "Number of reactions", rankingTitle: "Number of reactions ranking",
		unit:  "reactions",
		value: func(p *stats.Participant) int { return p.Reactions },
	},
	{
		chart: "Number of Reels sent", title: "Reels sent",
		ranking: "Number of reels", rankingTitle: "Number of reels sent ranking",
		unit:  "reels",
		value: func(p *stats.Participant) int { return p.Reels },
	},
	{
		chart: "Number of photos and videos sent", title: "Photos and videos sent",
		ranking: "Number of photos and videos", rankingTitle: "Number of photos and videos ranking",
		unit:  "photos and videos",
		value: func(p *stats.Participant) int { return p.PhotosAndVideos },
	},
}

// Names of the files of the activity graphs, without their extension.
const (
	weekChart  = "Activity by time and day"
	daysChart  = "Number of messages per day"
	monthChart = "Number of messages per month"
)

// Options selects the outputs to write.
type Options struct {
	// SearchedWords writes the graph and ranking of the searched words.
	SearchedWords bool
	// TextRankings writes the text rankings, next to the graphs.
	TextRankings bool
}

// Files returns the names of all the files a report can write, whatever its options.
func Files() []string {
	var files []string
	for _, c := range counts {
		files = append(files, c.chart+".png")
	}
	files = append(files, weekChart+".png", daysChart+".png", monthChart+".png")
	for _, c := range counts {
		files = append(files, c.ranking+".txt")
	}
	return files
}

// Write writes the graphs and rankings of s to dir, which must exist.
func Write(dir string, s *stats.Stats, options Options, logger *zap.Logger) error {
	for _, c := range counts {
		if c.searchedWords && !options.SearchedWords {
			continue
		}
		total := c.totalOf(s)
		if err := writePie(dir, s, c, total, logger); err != nil {
			return err
		}
		if options.TextRankings {
			path := filepath.Join(dir, c.ranking+".txt")
			logger.Debug("writing ranking", zap.String("file", path))
			ranking := formatRanking(fmt.Sprintf("%s (Total %d):", c.rankingTitle, total), c.unit, s.Ranking(c.value))
			if err := os.WriteFile(path, []byte(ranking), 0o644); err != nil {
				return err
			}
		}
	}

	days := s.Days()
	if len(days) == 0 {
		logger.Warn("no messages: skipping the activity graphs")
		return nil
	}
	if err := writeChart(dir, weekChart, logger, func(dc draw.Canvas) {
		drawHeatmap(dc, "Activity by time and day", s.AverageWeek())
	}); err != nil {
		return err
	}
	if err := writeChart(dir, daysChart, logger, func(dc draw.Canvas) {
		drawBars(dc, "Messages per day", "Date", "Number of messages", days, dayBarColor)
	}); err != nil {
		return err
	}
	return writeChart(dir, monthChart, logger, func(dc draw.Canvas) {
		drawBars(dc, "Messages per month", "Month", "Number of messages", s.Months(), monthBarColor)
	})
}

func (c count) totalOf(s *stats.Stats) int {
	if c.total != nil {
		return c.total(s)
	}
	total := 0
	for _, p := range s.Participants() {
		total += c.value(p)
	}
	return total
}

func writePie(dir string, s *stats.Stats, c count, total int, logger *zap.Logger) error {
	pie := pieSlices(s.Participants(), c.value)
	if len(pie) == 0 {
		logger.Info("skipping pie chart: every count is 0", zap.String("chart", c.chart))
		return nil
	}
	return writeChart(dir, c.chart, logger, func(dc draw.Canvas) {
		drawPie(dc, fmt.Sprintf("%s (Total %s)", c.title, formatCount(total)), c.unit, pie)
	})
}

// writeChart draws a graph with drawChart and writes it to dir, in the PNG file name.
func writeChart(dir, name string, logger *zap.Logger, drawChart func(draw.Canvas)) error {
	path := filepath.Join(dir, name+".png")
	logger.Debug("writing graph", zap.String("file", path))
	img, dc := newCanvas()
	drawChart(draw.Crop(dc, margin, -margin, margin, -margin))
	return savePNG(img, path)
}

// formatRanking returns the text of a ranking: its title, then a line per rank.
func formatRanking(title, unit string, ranking []stats.Rank) string {
	var b strings.Builder
	b.WriteString(title + "\n")
	for i, rank := range ranking {
		fmt.Fprintf(&b, "%d: %d %s: %s\n", i+1, rank.Count, unit, strings.Join(rank.Names, ", "))
	}
	return b.String()
}
