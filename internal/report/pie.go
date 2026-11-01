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
	label string
	value int
	color color.Color
}

// pieSlices returns the slices of the participants whose count is not 0, from the
// biggest to the smallest.
func pieSlices(participants []*stats.Participant, count func(*stats.Participant) int) []slice {
	var pie []slice
	for i, p := range participants {
		if n := count(p); n > 0 {
			pie = append(pie, slice{label: drawable(p.Name), value: n, color: participantColor(i)})
		}
	}
	slices.SortStableFunc(pie, func(a, b slice) int { return cmp.Compare(b.value, a.value) })
	if len(pie) > maxSlices {
		others := slice{label: fmt.Sprintf("%d others", len(pie)-maxSlices+1), color: othersColor}
		for _, s := range pie[maxSlices-1:] {
			others.value += s.value
		}
		pie = append(pie[:maxSlices-1], others)
	}
	return pie
}

// drawPie draws a pie chart of slices, titled title, with a legend giving the count of
// each slice in unit.
func drawPie(dc draw.Canvas, title, unit string, pie []slice) {
	width := dc.Max.X - dc.Min.X

	titleStyle := textStyle(22, true, textColor)
	titleStyle.XAlign, titleStyle.YAlign = draw.XCenter, draw.YTop
	dc.FillText(titleStyle, vg.Point{X: dc.Min.X + width/2, Y: dc.Max.Y}, title)

	top, bottom := dc.Max.Y-0.85*vg.Inch, dc.Min.Y+0.15*vg.Inch
	radius := (top - bottom) / 2
	center := vg.Point{X: dc.Min.X + 0.35*vg.Inch + radius, Y: (top + bottom) / 2}

	sum := 0
	for _, s := range pie {
		sum += s.value
	}

	start := math.Pi / 2
	for _, s := range pie {
		sweep := 2 * math.Pi * float64(s.value) / float64(sum)
		var wedge vg.Path
		wedge.Move(center)
		wedge.Arc(center, radius, start, -sweep)
		wedge.Close()
		dc.SetColor(s.color)
		dc.Fill(wedge)
		if len(pie) > 1 {
			dc.SetLineStyle(draw.LineStyle{Color: backgroundColor, Width: vg.Points(1.5)})
			dc.Stroke(wedge)
		}

		if share := float64(s.value) / float64(sum); share >= minPercentLabel {
			middle := start - sweep/2
			style := textStyle(13, true, contrastingText(s.color))
			style.XAlign, style.YAlign = draw.XCenter, draw.YCenter
			at := vg.Point{
				X: center.X + vg.Length(0.68*math.Cos(middle))*radius,
				Y: center.Y + vg.Length(0.68*math.Sin(middle))*radius,
			}
			dc.FillText(style, at, fmt.Sprintf("%.1f%%", 100*share))
		}
		start -= sweep
	}

	legendLeft := center.X + radius + 0.8*vg.Inch
	legendWidth := dc.Max.X - legendLeft
	rowHeight := min(0.42*vg.Inch, (top-bottom)/vg.Length(len(pie)))
	fontSize := max(9, min(16, rowHeight*0.55))
	style := textStyle(fontSize, false, textColor)
	style.YAlign = draw.YCenter
	y := center.Y + rowHeight*vg.Length(len(pie))/2 - rowHeight/2
	for _, s := range pie {
		swatch := vg.Rectangle{
			Min: vg.Point{X: legendLeft, Y: y - fontSize/2},
			Max: vg.Point{X: legendLeft + fontSize, Y: y + fontSize/2},
		}
		dc.SetColor(s.color)
		dc.Fill(swatch.Path())

		label := fmt.Sprintf("%s: %s %s (%.1f%%)", s.label, formatCount(s.value), unit,
			100*float64(s.value)/float64(sum))
		textLeft := legendLeft + 1.6*fontSize
		dc.FillText(style, vg.Point{X: textLeft, Y: y}, fit(style, label, legendWidth-(textLeft-legendLeft)))
		y -= rowHeight
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

// contrastingText returns the color of a text readable on background.
func contrastingText(background color.Color) color.Color {
	r, g, b, _ := background.RGBA()
	luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 0xFFFF
	if luminance > 0.6 {
		return color.Black
	}
	return color.White
}
