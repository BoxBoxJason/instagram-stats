package report

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette/moreland"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"

	"github.com/boxboxjason/instagram-stats/internal/stats"
)

// maxTimeLabels is the most labels the time axis of a bar graph has.
const maxTimeLabels = 16

// weekdays are the days of the week, in the order of [stats.Stats.AverageWeek].
var weekdays = []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

// bars is a plotter drawing a bar per value, at x = 0, 1, 2...
type bars struct {
	values []int
	color  color.Color
}

// Plot draws the bars, separated by a gap while they are wide enough to see it.
func (b bars) Plot(c draw.Canvas, p *plot.Plot) {
	toX, toY := p.Transforms(&c)
	half := 0.5
	if len(b.values) <= 120 {
		half = 0.4
	}
	for i, value := range b.values {
		if value == 0 {
			continue
		}
		left, right := toX(float64(i)-half), toX(float64(i)+half)
		bottom, top := toY(0), toY(float64(value))
		c.FillPolygon(b.color, []vg.Point{{X: left, Y: bottom}, {X: right, Y: bottom}, {X: right, Y: top}, {X: left, Y: top}})
	}
}

// DataRange returns the range covered by the bars.
func (b bars) DataRange() (xmin, xmax, ymin, ymax float64) {
	for _, value := range b.values {
		ymax = max(ymax, float64(value))
	}
	return -0.5, float64(len(b.values)) - 0.5, 0, max(ymax, 1)
}

// drawBars draws a bar graph of counts, one bar per period.
func drawBars(dc draw.Canvas, title, xLabel, yLabel string, counts []stats.Count, clr color.Color) {
	p := newPlot(title, xLabel, yLabel)
	values := make([]int, len(counts))
	starts := make([]time.Time, len(counts))
	for i, count := range counts {
		values[i] = count.Messages
		starts[i] = count.Start
	}

	grid := plotter.NewGrid()
	grid.Vertical.Color = nil
	grid.Horizontal.Color = gridColor
	p.Add(grid, bars{values: values, color: clr})
	p.X.Tick.Marker = timeTicks(starts)
	p.X.Padding, p.Y.Padding = 0, 0
	p.Y.Min = 0
	p.Draw(dc)
}

// timeTicks returns the ticks of a time axis whose values are the indexes of starts,
// the first days of consecutive periods (days or months).
//
// The labels are on the first day of the months, every one, two, three... months so
// there are at most maxTimeLabels of them, or on days when that makes too few.
func timeTicks(starts []time.Time) plot.Ticker {
	var ticks []plot.Tick
	var monthStarts []int
	for i, start := range starts {
		if start.Day() == 1 {
			monthStarts = append(monthStarts, i)
		}
	}

	if len(monthStarts) >= 3 {
		step := 1
		for _, candidate := range []int{1, 2, 3, 4, 6, 12, 24, 36, 60, 120} {
			step = candidate
			if (len(monthStarts)+candidate-1)/candidate <= maxTimeLabels {
				break
			}
		}
		for _, i := range monthStarts {
			if monthNumber(starts[i])%step == 0 {
				ticks = append(ticks, plot.Tick{Value: float64(i), Label: starts[i].Format("Jan 2006")})
			}
		}
		return plot.ConstantTicks(ticks)
	}

	step := (len(starts) + maxTimeLabels - 1) / maxTimeLabels
	for i := 0; i < len(starts); i += max(step, 1) {
		ticks = append(ticks, plot.Tick{Value: float64(i), Label: starts[i].Format("2 Jan 2006")})
	}
	return plot.ConstantTicks(ticks)
}

// monthNumber returns the number of months from year 0 to the month of t.
func monthNumber(t time.Time) int {
	return t.Year()*12 + int(t.Month()) - 1
}

// weekGrid is the grid of a heatmap of a week: a column per day, a row per hour.
type weekGrid [7][24]float64

func (g *weekGrid) Dims() (c, r int)   { return 7, 24 }
func (g *weekGrid) Z(c, r int) float64 { return g[c][r] }
func (g *weekGrid) X(c int) float64    { return float64(c) }
func (g *weekGrid) Y(r int) float64    { return float64(r) }
func (g *weekGrid) Min() float64       { return g.extreme(math.Min) }
func (g *weekGrid) Max() float64       { return g.extreme(math.Max) }

func (g *weekGrid) extreme(pick func(a, b float64) float64) float64 {
	extreme := g[0][0]
	for _, hours := range g {
		for _, value := range hours {
			extreme = pick(extreme, value)
		}
	}
	return extreme
}

// drawHeatmap draws the heatmap of the messages of an average week, with a color bar.
func drawHeatmap(dc draw.Canvas, title string, week [7][24]float64) {
	grid := weekGrid(week)
	colors := moreland.SmoothBlueRed()
	low, high := grid.Min(), grid.Max()
	if high <= low {
		high = low + 1
	}
	colors.SetMin(low)
	colors.SetMax(high)

	heatmap := plotter.NewHeatMap(&grid, colors.Palette(256))
	heatmap.Min, heatmap.Max = low, high

	p := newPlot(title, "Day", "Time")
	p.Add(heatmap)
	p.X.Padding, p.Y.Padding = 0, 0
	// Midnight at the top, as on a timetable.
	p.Y.Scale = plot.InvertedScale{Normalizer: plot.LinearScale{}}
	var dayTicks, hourTicks []plot.Tick
	for i, day := range weekdays {
		dayTicks = append(dayTicks, plot.Tick{Value: float64(i), Label: day})
	}
	for hour := range 24 {
		hourTicks = append(hourTicks, plot.Tick{Value: float64(hour), Label: fmt.Sprintf("%02d:00", hour)})
	}
	p.X.Tick.Marker = plot.ConstantTicks(dayTicks)
	p.Y.Tick.Marker = plot.ConstantTicks(hourTicks)

	const barWidth = 1.3 * vg.Inch
	heatmapCanvas := draw.Crop(dc, 0, -barWidth, 0, 0)
	p.Draw(heatmapCanvas)
	data := p.DataCanvas(heatmapCanvas)

	bar := newPlot("", "", "Messages per hour")
	bar.Add(&plotter.ColorBar{ColorMap: colors, Vertical: true})
	bar.HideX()
	bar.Y.Padding = 0
	// Align the bar with the data area of the heatmap.
	barCanvas := draw.Crop(dc, dc.Max.X-dc.Min.X-barWidth+0.2*vg.Inch, -0.2*vg.Inch, data.Min.Y-dc.Min.Y, data.Max.Y-dc.Max.Y)
	bar.Draw(barCanvas)
}
