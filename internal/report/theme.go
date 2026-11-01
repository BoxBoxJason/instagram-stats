package report

import (
	"bufio"
	"fmt"
	"image/color"
	"math"
	"os"
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

// Colors of the dark theme of the graphs.
var (
	backgroundColor = color.Black
	textColor       = color.White
	labelColor      = color.Gray{Y: 0xA0}
	axisColor       = color.Gray{Y: 0x80}
	gridColor       = color.Gray{Y: 0x30}
	othersColor     = color.Gray{Y: 0x70}
	dayBarColor     = color.RGBA{R: 0x4C, G: 0x8B, B: 0xF5, A: 0xFF}
	monthBarColor   = color.RGBA{R: 0xA8, G: 0x6C, B: 0xF0, A: 0xFF}
)

// margin is the space left around the graphs.
const margin = 0.25 * vg.Inch

// drawable returns name with only the characters the font of the graphs can draw.
// Names often hold emoji, or letters styled with mathematical symbols (𝕹𝖆𝖒𝖊) that
// are first turned back into plain letters. A name left empty becomes "(unnamed)".
func drawable(name string) string {
	face := font.DefaultCache.Lookup(font.Font{Typeface: "Liberation", Variant: "Sans"}, 12)
	var buf sfnt.Buffer
	var b strings.Builder
	for _, r := range norm.NFKC.String(name) {
		if index, err := face.Face.GlyphIndex(&buf, r); err == nil && index != 0 || r == ' ' {
			b.WriteRune(r)
		}
	}
	if drawn := strings.Join(strings.Fields(b.String()), " "); drawn != "" {
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
	p := plot.New()
	p.BackgroundColor = backgroundColor
	p.Title.Text = title
	restyle(&p.Title.TextStyle, 22, true, textColor)
	p.Title.Padding = 0.25 * vg.Inch

	for _, axis := range []*plot.Axis{&p.X, &p.Y} {
		axis.Color = axisColor
		axis.Tick.Color = axisColor
		restyle(&axis.Tick.Label, 11, false, textColor)
		restyle(&axis.Label.TextStyle, 14, true, labelColor)
		axis.Label.Padding = 0.1 * vg.Inch
	}
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel
	return p
}

// newCanvas returns an image to draw a graph on, filled with the background color.
func newCanvas() (*vgimg.Canvas, draw.Canvas) {
	img := vgimg.NewWith(vgimg.UseWH(imageWidth, imageHeight), vgimg.UseDPI(imageDPI),
		vgimg.UseBackgroundColor(backgroundColor))
	return img, draw.New(img)
}

// savePNG writes img to path, as a PNG image.
func savePNG(img *vgimg.Canvas, path string) (err error) {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()
	writer := bufio.NewWriter(file)
	if _, err := (vgimg.PngCanvas{Canvas: img}).WriteTo(writer); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return writer.Flush()
}

// participantColor returns the color of the participant at index, the same in every
// graph. The hues are spread by the golden angle, so close indexes get distinct colors.
func participantColor(index int) color.Color {
	hue := math.Mod(float64(index)*137.508, 360)
	lightness := []float64{0.55, 0.45, 0.65}[index%3]
	return hsl(hue, 0.65, lightness)
}

// hsl returns the color of hue (in degrees), saturation and lightness (in [0, 1]).
func hsl(hue, saturation, lightness float64) color.Color {
	chroma := (1 - math.Abs(2*lightness-1)) * saturation
	x := chroma * (1 - math.Abs(math.Mod(hue/60, 2)-1))
	var r, g, b float64
	switch {
	case hue < 60:
		r, g = chroma, x
	case hue < 120:
		r, g = x, chroma
	case hue < 180:
		g, b = chroma, x
	case hue < 240:
		g, b = x, chroma
	case hue < 300:
		r, b = x, chroma
	default:
		r, b = chroma, x
	}
	m := lightness - chroma/2
	channel := func(v float64) uint8 { return uint8(math.Round((v + m) * 0xFF)) }
	return color.RGBA{R: channel(r), G: channel(g), B: channel(b), A: 0xFF}
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
