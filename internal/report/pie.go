package report

import (
	"cmp"
	"fmt"
	"image/color"
	"math"
	"slices"

	"gonum.org/v1/plot/text"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"

	"github.com/boxboxjason/instagram-stats/internal/stats"
)

// maxSlices is the most slices a pie chart has: past it, the smallest counts are
// grouped in a single slice, the text rankings giving the details.
const maxSlices = 20

// minPercentLabel is the smallest share of the pie whose slice shows its percentage.
const minPercentLabel = 0.04

// slice is a slice of a pie chart.
type slice struct {
	color color.Color
	label string
	value int
}

// pieSlices returns the slices of the participants whose count is not 0, from the
// biggest to the smallest.
func pieSlices(participants []*stats.Participant, count func(*stats.Participant) int) []slice {
	var pie []slice

	for i, participant := range participants {
		if total := count(participant); total > 0 {
			pie = append(pie, slice{label: drawable(participant.Name), value: total, color: participantColor(i)})
		}
	}

	slices.SortStableFunc(pie, func(a, b slice) int { return cmp.Compare(b.value, a.value) })

	if len(pie) > maxSlices {
		others := slice{label: fmt.Sprintf("%d others", len(pie)-maxSlices+1), color: othersColor()}
		for _, entry := range pie[maxSlices-1:] {
			others.value += entry.value
		}

		pie = append(pie[:maxSlices-1], others)
	}

	return pie
}

// half is a half, to take the middle of a length, and fullTurn the angle of a turn.
const (
	half     = 0.5
	fullTurn = 2 * math.Pi
)

// Dimensions of the pie charts.
const (
	titleSize        = 22
	percentSize      = 13
	percentRadius    = 0.68 // distance of the percentage of a slice from the center, in radiuses
	topSpace         = 0.85 * vg.Inch
	bottomSpace      = 0.15 * vg.Inch
	leftSpace        = 0.35 * vg.Inch
	legendGap        = 0.8 * vg.Inch
	legendRowHeight  = 0.42 * vg.Inch
	legendMinFont    = 9
	legendMaxFont    = 16
	legendFontRatio  = 0.55 // font size of the legend, in row heights
	legendTextOffset = 1.6  // space of the swatch before the text of the legend, in font sizes
	sliceBorder      = 1.5
	fullPercent      = 100
)

// drawPie draws a pie chart of slices, titled title, with a legend giving the count of
// each slice in unit.
func drawPie(canvas draw.Canvas, title, unit string, pie []slice) {
	width := canvas.Max.X - canvas.Min.X

	titleStyle := textStyle(titleSize, true, textColor())
	titleStyle.XAlign, titleStyle.YAlign = draw.XCenter, draw.YTop
	canvas.FillText(titleStyle, vg.Point{X: canvas.Min.X + width/2, Y: canvas.Max.Y}, title)

	top, bottom := canvas.Max.Y-topSpace, canvas.Min.Y+bottomSpace
	radius := (top - bottom) * half
	center := vg.Point{X: canvas.Min.X + leftSpace + radius, Y: (top + bottom) * half}

	sum := 0
	for _, entry := range pie {
		sum += entry.value
	}

	drawSlices(canvas, center, radius, pie, sum)

	legendLeft := center.X + radius + legendGap
	rowHeight := min(legendRowHeight, (top-bottom)/vg.Length(len(pie)))
	legendTop := center.Y + rowHeight*vg.Length(len(pie))*half - rowHeight*half

	drawLegend(canvas, legendLeft, legendTop, rowHeight, unit, pie, sum)
}

// drawSlices draws the slices of pie, whose values add up to sum, clockwise from the top.
func drawSlices(canvas draw.Canvas, center vg.Point, radius vg.Length, pie []slice, sum int) {
	start := math.Pi * half

	for _, entry := range pie {
		share := float64(entry.value) / float64(sum)
		sweep := fullTurn * share

		var wedge vg.Path
		wedge.Move(center)
		wedge.Arc(center, radius, start, -sweep)
		wedge.Close()
		canvas.SetColor(entry.color)
		canvas.Fill(wedge)

		if len(pie) > 1 {
			canvas.SetLineStyle(draw.LineStyle{Color: backgroundColor(), Width: vg.Points(sliceBorder)})
			canvas.Stroke(wedge)
		}

		if share >= minPercentLabel {
			middle := start - sweep*half
			style := textStyle(percentSize, true, contrastingText(entry.color))
			style.XAlign, style.YAlign = draw.XCenter, draw.YCenter
			at := vg.Point{
				X: center.X + vg.Length(percentRadius*math.Cos(middle))*radius,
				Y: center.Y + vg.Length(percentRadius*math.Sin(middle))*radius,
			}
			canvas.FillText(style, at, fmt.Sprintf("%.1f%%", fullPercent*share))
		}

		start -= sweep
	}
}

// drawLegend draws a row per slice of pie, from the row at height top, at legendLeft.
func drawLegend(canvas draw.Canvas, legendLeft, top, rowHeight vg.Length, unit string, pie []slice, sum int) {
	legendWidth := canvas.Max.X - legendLeft
	fontSize := max(legendMinFont, min(legendMaxFont, rowHeight*legendFontRatio))
	style := textStyle(fontSize, false, textColor())
	style.YAlign = draw.YCenter

	row := top
	for _, entry := range pie {
		swatch := vg.Rectangle{
			Min: vg.Point{X: legendLeft, Y: row - fontSize/2},
			Max: vg.Point{X: legendLeft + fontSize, Y: row + fontSize/2},
		}

		canvas.SetColor(entry.color)
		canvas.Fill(swatch.Path())

		label := fmt.Sprintf("%s: %s %s (%.1f%%)", entry.label, formatCount(entry.value), unit,
			fullPercent*float64(entry.value)/float64(sum))
		textLeft := legendLeft + legendTextOffset*fontSize
		canvas.FillText(style, vg.Point{X: textLeft, Y: row}, fit(style, label, legendWidth-(textLeft-legendLeft)))
		row -= rowHeight
	}
}

// fit returns label, shortened with an ellipsis when it is wider than width.
func fit(style text.Style, label string, width vg.Length) string {
	if style.Width(label) <= width {
		return label
	}

	runes := []rune(label)
	for len(runes) > 1 {
		runes = runes[:len(runes)-1]
		if shortened := string(runes) + "…"; style.Width(shortened) <= width {
			return shortened
		}
	}

	return string(runes)
}

// Weights of the channels in the luminance of a color, and luminance from which a color
// is light.
const (
	redWeight      = 0.299
	greenWeight    = 0.587
	blueWeight     = 0.114
	maxChannel     = 0xFFFF
	lightThreshold = 0.6
)

// contrastingText returns the color of a text readable on background.
func contrastingText(background color.Color) color.Color {
	red, green, blue, _ := background.RGBA()

	luminance := (redWeight*float64(red) + greenWeight*float64(green) + blueWeight*float64(blue)) / maxChannel
	if luminance > lightThreshold {
		return color.Black
	}

	return color.White
}
