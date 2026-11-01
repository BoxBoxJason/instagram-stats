package report

import (
	"bufio"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	stdfont "golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/text/unicode/norm"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/text"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

// Size and resolution of the images: 1600×900 pixels.
const (
	imageWidth  = 16 * vg.Inch
	imageHeight = 9 * vg.Inch
	imageDPI    = 100
)

// Channel values of the colors of the dark theme of the graphs.
const (
	opaque     = 0xFF
	labelGray  = 0xA0
	axisGray   = 0x80
	gridGray   = 0x30
	othersGray = 0x70
	dayRed     = 0x4C
	dayGreen   = 0x8B
	dayBlue    = 0xF5
	monthRed   = 0xA8
	monthGreen = 0x6C
	monthBlue  = 0xF0
	titleFont  = 22
	tickFont   = 11
	labelFont  = 14
	drawnFont  = 12
	titlePad   = 0.25 * vg.Inch
	labelPad   = 0.1 * vg.Inch
)

// Colors of the dark theme of the graphs.
func backgroundColor() color.Color { return color.Black }
func textColor() color.Color       { return color.White }
func labelColor() color.Color      { return color.Gray{Y: labelGray} }
func axisColor() color.Color       { return color.Gray{Y: axisGray} }
func gridColor() color.Color       { return color.Gray{Y: gridGray} }
func othersColor() color.Color     { return color.Gray{Y: othersGray} }
func dayBarColor() color.Color     { return color.RGBA{R: dayRed, G: dayGreen, B: dayBlue, A: opaque} }
func monthBarColor() color.Color {
	return color.RGBA{R: monthRed, G: monthGreen, B: monthBlue, A: opaque}
}

// margin is the space left around the graphs.
const margin = 0.25 * vg.Inch

// drawable returns name with only the characters the font of the graphs can draw.
// Names often hold emoji, or letters styled with mathematical symbols (𝕹𝖆𝖒𝖊) that
// are first turned back into plain letters. A name left empty becomes "(unnamed)".
func drawable(name string) string {
	face := font.DefaultCache.Lookup(font.Font{Typeface: "Liberation", Variant: "Sans"}, drawnFont)

	var (
		buf sfnt.Buffer
		out strings.Builder
	)

	for _, r := range norm.NFKC.String(name) {
		index, err := face.Face.GlyphIndex(&buf, r)
		if err == nil && index != 0 || r == ' ' {
			out.WriteRune(r)
		}
	}

	if drawn := strings.Join(strings.Fields(out.String()), " "); drawn != "" {
		return drawn
	}

	return "(unnamed)"
}

// textStyle returns the style of a text in the font of the graphs.
func textStyle(size vg.Length, bold bool, clr color.Color) text.Style {
	var style text.Style
	restyle(&style, size, bold, clr)
	style.Handler = plot.DefaultTextHandler

	return style
}

// restyle sets the font and color of style, keeping its alignment and rotation.
func restyle(style *text.Style, size vg.Length, bold bool, clr color.Color) {
	f := font.Font{Typeface: "Liberation", Variant: "Sans"}
	if bold {
		f.Weight = stdfont.WeightBold
	}

	style.Font = font.From(f, size)
	style.Color = clr
}

// newPlot returns a plot in the dark theme of the graphs.
func newPlot(title, xLabel, yLabel string) *plot.Plot {
	plt := plot.New()
	plt.BackgroundColor = backgroundColor()
	plt.Title.Text = title
	restyle(&plt.Title.TextStyle, titleFont, true, textColor())
	plt.Title.Padding = titlePad

	for _, axis := range []*plot.Axis{&plt.X, &plt.Y} {
		axis.Color = axisColor()
		axis.Tick.Color = axisColor()
		restyle(&axis.Tick.Label, tickFont, false, textColor())
		restyle(&axis.Label.TextStyle, labelFont, true, labelColor())
		axis.Label.Padding = labelPad
	}

	plt.X.Label.Text = xLabel
	plt.Y.Label.Text = yLabel

	return plt
}

// newCanvas returns an image to draw a graph on, filled with the background color.
func newCanvas() (*vgimg.Canvas, draw.Canvas) {
	img := vgimg.NewWith(vgimg.UseWH(imageWidth, imageHeight), vgimg.UseDPI(imageDPI),
		vgimg.UseBackgroundColor(backgroundColor()))

	return img, draw.New(img)
}

// savePNG writes img to path, as a PNG image.
func savePNG(img *vgimg.Canvas, path string) (err error) {
	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()

	writer := bufio.NewWriter(file)

	_, err = (vgimg.PngCanvas{Canvas: img}).WriteTo(writer)
	if err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}

// Parameters of the colors of the participants, and of the conversion from HSL to RGB.
const (
	goldenAngle  = 137.508
	fullCircle   = 360
	participantS = 0.65
	hueSector    = 60
	hueSectors   = 6
	pairMod      = 2
)

// participantColor returns the color of the participant at index, the same in every
// graph. The hues are spread by the golden angle, so close indexes get distinct colors.
func participantColor(index int) color.Color {
	hue := math.Mod(float64(index)*goldenAngle, fullCircle)
	lightness := []float64{0.55, 0.45, 0.65}[index%3]

	return hsl(hue, participantS, lightness)
}

// hsl returns the color of hue (in degrees), saturation and lightness (in [0, 1]).
func hsl(hue, saturation, lightness float64) color.Color {
	chroma := (1 - math.Abs(2*lightness-1)) * saturation
	second := chroma * (1 - math.Abs(math.Mod(hue/hueSector, pairMod)-1))

	rgb := [hueSectors][3]float64{
		{chroma, second, 0}, {second, chroma, 0}, {0, chroma, second},
		{0, second, chroma}, {second, 0, chroma}, {chroma, 0, second},
	}[int(hue/hueSector)%hueSectors]

	match := lightness - chroma*half
	channel := func(v float64) uint8 { return uint8(math.Round((v + match) * opaque)) }

	return color.RGBA{R: channel(rgb[0]), G: channel(rgb[1]), B: channel(rgb[2]), A: opaque}
}

// formatCount formats n with a comma between groups of three digits.
func formatCount(n int) string {
	digits := strconv.Itoa(n)
	if n < 0 {
		return "-" + formatCount(-n)
	}

	var out []byte

	for i, digit := range []byte(digits) {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out = append(out, ',')
		}

		out = append(out, digit)
	}

	return string(out)
}
