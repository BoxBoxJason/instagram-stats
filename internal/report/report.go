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
	value         func(*stats.Participant) int
	total         func(*stats.Stats) int
	chart         string
	title         string
	ranking       string
	rankingTitle  string
	unit          string
	searchedWords bool
}

const wordsFound = "Number of words found"

// counts returns the counts drawn and ranked, in the order of the report.
func counts() []count {
	return []count{
		{
			chart: wordsFound, title: wordsFound,
			ranking: wordsFound, rankingTitle: wordsFound,
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
	all := counts()
	files := make([]string, 0, 2*len(all)+activityCharts)

	for _, entry := range all {
		files = append(files, entry.chart+".png")
	}

	files = append(files, weekChart+".png", daysChart+".png", monthChart+".png")
	for _, entry := range all {
		files = append(files, entry.ranking+".txt")
	}

	return files
}

// activityCharts is the number of graphs of the activity over time.
const activityCharts = 3

// rankingMode is the permissions of the ranking files.
const rankingMode = 0o600

// Write writes the graphs and rankings of statistics to dir, which must exist.
func Write(dir string, statistics *stats.Stats, options Options, logger *zap.Logger) error {
	for _, entry := range counts() {
		if entry.searchedWords && !options.SearchedWords {
			continue
		}

		total := entry.totalOf(statistics)

		err := writePie(dir, statistics, entry, total, logger)
		if err != nil {
			return err
		}

		if options.TextRankings {
			err = writeRanking(dir, statistics, entry, total, logger)
			if err != nil {
				return err
			}
		}
	}

	return writeActivity(dir, statistics, logger)
}

// writeRanking writes the text ranking of entry to dir.
func writeRanking(dir string, statistics *stats.Stats, entry count, total int, logger *zap.Logger) error {
	path := filepath.Join(dir, entry.ranking+".txt")
	logger.Debug("writing ranking", zap.String("file", path))

	title := fmt.Sprintf("%s (Total %d):", entry.rankingTitle, total)
	ranking := formatRanking(title, entry.unit, statistics.Ranking(entry.value))

	err := os.WriteFile(path, []byte(ranking), rankingMode)
	if err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}

// writeActivity writes the graphs of the activity over time, unless there are no messages.
func writeActivity(dir string, statistics *stats.Stats, logger *zap.Logger) error {
	days := statistics.Days()
	if len(days) == 0 {
		logger.Warn("no messages: skipping the activity graphs")

		return nil
	}

	err := writeChart(dir, weekChart, logger, func(canvas draw.Canvas) {
		drawHeatmap(canvas, "Activity by time and day", statistics.AverageWeek())
	})
	if err != nil {
		return err
	}

	err = writeChart(dir, daysChart, logger, func(canvas draw.Canvas) {
		drawBars(canvas, "Messages per day", "Date", "Number of messages", days, dayBarColor())
	})
	if err != nil {
		return err
	}

	return writeChart(dir, monthChart, logger, func(canvas draw.Canvas) {
		drawBars(canvas, "Messages per month", "Month", "Number of messages", statistics.Months(), monthBarColor())
	})
}

func (entry count) totalOf(statistics *stats.Stats) int {
	if entry.total != nil {
		return entry.total(statistics)
	}

	total := 0
	for _, participant := range statistics.Participants() {
		total += entry.value(participant)
	}

	return total
}

func writePie(dir string, statistics *stats.Stats, entry count, total int, logger *zap.Logger) error {
	pie := pieSlices(statistics.Participants(), entry.value)
	if len(pie) == 0 {
		logger.Info("skipping pie chart: every count is 0", zap.String("chart", entry.chart))

		return nil
	}

	return writeChart(dir, entry.chart, logger, func(canvas draw.Canvas) {
		drawPie(canvas, fmt.Sprintf("%s (Total %s)", entry.title, formatCount(total)), entry.unit, pie)
	})
}

// writeChart draws a graph with drawChart and writes it to dir, in the PNG file name.
func writeChart(dir, name string, logger *zap.Logger, drawChart func(draw.Canvas)) error {
	path := filepath.Join(dir, name+".png")
	logger.Debug("writing graph", zap.String("file", path))

	img, canvas := newCanvas()
	drawChart(draw.Crop(canvas, margin, -margin, margin, -margin))

	return savePNG(img, path)
}

// formatRanking returns the text of a ranking: its title, then a line per rank.
func formatRanking(title, unit string, ranking []stats.Rank) string {
	var out strings.Builder
	out.WriteString(title + "\n")

	for i, rank := range ranking {
		fmt.Fprintf(&out, "%d: %d %s: %s\n", i+1, rank.Count, unit, strings.Join(rank.Names, ", "))
	}

	return out.String()
}
