// Package dchart2 makes charts using the deck markup
package dchart2

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/ajstarks/deck/generate"
)

// NameValue is a name,value pair
type NameValue struct {
	Label string
	Note  string
	Value float64
}

// ChartBox holds the essential data for making a chart
type ChartBox struct {
	Data       []NameValue
	Title      string
	DataFormat string
	DataColor  string
	LabelColor string
	ValueColor string
	Opacity    float64
	TextSize   float64
	Top        float64
	Bottom     float64
	Left       float64
	Right      float64
	Minvalue   float64
	Maxvalue   float64
	Zerobased  bool
}

// Flags define chart on/off switches
type Flags struct {
	DataMinimum,
	FullDeck,
	ReadCSV,
	ShowAxis,
	ShowBar,
	ShowDonut,
	ShowVDot,
	ShowHDot,
	ShowFrame,
	ShowGrid,
	ShowHBar,
	ShowLine,
	ShowNote,
	ShowPercentage,
	ShowPGrid,
	ShowPMap,
	ShowRadial,
	ShowRegressionLine,
	ShowScatter,
	ShowSlope,
	ShowSpokes,
	ShowTitle,
	ShowValues,
	ShowVolume,
	ShowWBar,
	ShowXLast,
	ShowXstagger,
	SolidPMap bool
}

// Attributes define chart attributes
type Attributes struct {
	BackgroundColor,
	DataColor,
	FrameColor,
	LabelColor,
	RegressionLineColor,
	ValueColor,
	ChartTitle,
	CSVCols,
	DataCondition,
	DataFmt,
	HLine,
	NoteLocation,
	ValuePosition,
	YAxisR string
}

// Measures define chart measures
type Measures struct {
	TextSize,
	Left,
	Right,
	Top,
	Bottom,
	LineSpacing,
	BarWidth,
	LineWidth,
	PSize,
	PWidth,
	UserMin,
	UserMax,
	VolumeOpacity,
	XLabelRotation float64
	XLabelInterval,
	PMapLength int
}

// Settings is a collection of all chart settings
type Settings struct {
	Flags
	Attributes
	Measures
}

const (
	largest      = math.MaxFloat64
	smallest     = -largest
	valuecolor   = "rgb(128,0,0)"
	labelcolor   = "rgb(75,75,75)"
	dotlinecolor = "lightgray"
	wbopacity    = 30.0
	topclock     = math.Pi / 2
	fullcircle   = math.Pi * 2
	transparency = 50.0
)

var blue7 = []string{
	"rgb(8,69,148)",
	"rgb(33,113,181)",
	"rgb(66,146,198)",
	"rgb(107,174,214)",
	"rgb(158,202,225)",
	"rgb(198,219,239)",
	"rgb(239,243,255)",
}

var xmlmap = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;")

// xmlesc escapes XML
func xmlesc(s string) string {
	return xmlmap.Replace(s)
}

// getheader returns the indicies of the comma-separated list of fields
// by default or on error, return 0, 1. For example given this header:
//
// first,second,third,sum
//
// first,sum returns 0,3 and first,third returns 0,2
func getheader(s []string, lv string) (int, int) {
	li := 0
	vi := 1
	cv := strings.Split(lv, ",")
	if len(cv) != 2 {
		return li, vi
	}
	for i, p := range s {
		if p == cv[0] {
			li = i
		}
		if p == cv[1] {
			vi = i
		}
	}
	return li, vi
}

// zerobase uses the correct base for scaling
func zerobase(usez bool, n float64) float64 {
	if usez {
		return 0
	}
	return n
}

// ReadTSV reads tab separated values into a ChartBox
// default values for the top, bottom, left, right (90,50,10,90) are filled in
// as is the default color, black
func ReadTSV(r io.Reader) (ChartBox, error) {
	var d NameValue
	var data []NameValue
	var err error
	maxval := smallest
	minval := largest
	title := ""
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		t := scanner.Text()
		if len(t) == 0 { // skip blank lines
			continue
		}
		if t[0] == '#' && len(t) > 2 { // process titles
			title = strings.TrimSpace(t[1:])
			continue
		}
		fields := strings.Split(t, "\t")
		if len(fields) < 2 {
			continue
		}
		if len(fields) == 3 {
			d.Note = fields[2]
		} else {
			d.Note = ""
		}
		d.Label = fields[0]
		d.Value, err = strconv.ParseFloat(fields[1], 64)
		if err != nil {
			d.Value = 0
		}
		if d.Value > maxval {
			maxval = d.Value
		}
		if d.Value < minval {
			minval = d.Value
		}
		data = append(data, d)
	}
	err = scanner.Err()
	return ChartBox{
		Title:      xmlesc(title),
		Data:       data,
		Minvalue:   minval,
		Maxvalue:   maxval,
		TextSize:   1.2,
		DataFormat: "%.1f",
		DataColor:  "rgb(128,128,128)",
		LabelColor: labelcolor,
		ValueColor: valuecolor,
		Opacity:    100,
		Left:       10,
		Right:      90,
		Top:        90,
		Bottom:     50,
		Zerobased:  true,
	}, err
}

// ReadCSV reads CSV values into a ChartBox
// default values for the top, bottom, left, right (90,50,10,90) are filled in
// as is the default color, black
func ReadCSV(r io.Reader, csvcols string) (ChartBox, error) {
	var (
		data []NameValue
		d    NameValue
		err  error
	)
	input := csv.NewReader(r)
	maxval := smallest
	minval := largest
	title := ""
	n := 0
	li := 0
	vi := 1
	for {
		n++
		fields, csverr := input.Read()
		if csverr == io.EOF {
			break
		}
		if csverr != nil {
			fmt.Fprintf(os.Stderr, "%v %v\n", csverr, fields)
			continue
		}

		if len(fields) < 2 {
			continue
		}
		if fields[0] == "#" {
			title = fields[1]
			continue
		}
		if len(fields) == 3 {
			d.Note = xmlesc(fields[2])
		} else {
			d.Note = ""
		}
		if n == 1 && len(csvcols) > 0 { // column header is assumed to be the first row
			li, vi = getheader(fields, csvcols)
			title = fields[vi]
			continue
		}

		d.Label = xmlesc(fields[li])
		d.Value, err = strconv.ParseFloat(fields[vi], 64)
		if err != nil {
			d.Value = 0
		}
		if d.Value > maxval {
			maxval = d.Value
		}
		if d.Value < minval {
			minval = d.Value
		}
		data = append(data, d)
	}
	return ChartBox{
		Title:      xmlesc(title),
		Data:       data,
		Minvalue:   minval,
		Maxvalue:   maxval,
		TextSize:   1.2,
		DataFormat: "%.1f",
		DataColor:  "rgb(128,128,128)",
		LabelColor: labelcolor,
		ValueColor: valuecolor,
		Opacity:    100,
		Left:       10,
		Right:      90,
		Top:        90,
		Bottom:     50,
		Zerobased:  true,
	}, err
}

// chart types

// Bar makes a (column) bar chart
func (c *ChartBox) Bar(deck *generate.Deck, size float64) {
	dlen := float64(len(c.Data) - 1)
	ymin := zerobase(c.Zerobased, c.Minvalue)
	for i, d := range c.Data {
		x := MapRange(float64(i), 0, dlen, c.Left, c.Right)
		y := MapRange(d.Value, ymin, c.Maxvalue, c.Bottom, c.Top)
		deck.Line(x, c.Bottom, x, y, size, c.DataColor, c.Opacity)
	}
}

// ConditionalBar makes a bar chart with conditional coloring
func (c *ChartBox) ConditionalBar(deck *generate.Deck, size float64, cmin, cmax float64, color string) {
	dlen := float64(len(c.Data) - 1)
	ymin := zerobase(c.Zerobased, c.Minvalue)
	for i, d := range c.Data {
		v := d.Value
		x := MapRange(float64(i), 0, dlen, c.Left, c.Right)
		y := MapRange(v, ymin, c.Maxvalue, c.Bottom, c.Top)
		deck.Line(x, c.Bottom, x, y, size, conditionalcolor(v, cmin, cmax, color, c.DataColor), c.Opacity)
	}
}

// WBar makes a word-based horizontal bar chart
func (c *ChartBox) WBar(deck *generate.Deck, linespacing float64, showval, showpct bool) {
	textsize := c.TextSize
	format := c.DataFormat
	data := c.Data
	var sum float64
	if showpct {
		sum = datasum(data)
	}

	y := c.Top
	left := c.Left
	right := c.Right
	hts := textsize / 2
	mts := textsize
	ymin := zerobase(c.Zerobased, c.Minvalue)
	for _, d := range data {
		deck.Text(left+hts, y, d.Label, "sans", textsize, c.LabelColor)
		bv := MapRange(d.Value, ymin, c.Maxvalue, left, right)
		deck.Line(left+hts, y+hts, bv, y+hts, textsize*1.5, c.DataColor, wbopacity)
		if showval {
			if showpct {
				avgs := fmt.Sprintf(" ("+format+"%%)", 100*(d.Value/sum))
				deck.TextEnd(left, y+(hts/2), fmt.Sprintf(format, d.Value)+avgs, "mono", mts, c.ValueColor)
			} else {
				deck.TextEnd(left, y+(hts/2), fmt.Sprintf(format, d.Value), "mono", mts, c.DataColor)
			}
		}
		y -= linespacing
	}
}

// HBar makes a horizontal bar chart
func (c *ChartBox) HBar(deck *generate.Deck, size, linespacing float64) {
	y := c.Top
	textsize := c.TextSize
	format := c.DataFormat

	xmin := zerobase(c.Zerobased, c.Minvalue)
	for _, d := range c.Data {
		v := d.Value
		deck.TextEnd(c.Left-textsize, y-size/2, d.Label, "sans", textsize, c.LabelColor)
		x2 := MapRange(v, xmin, c.Maxvalue, c.Left, c.Right)
		deck.Line(c.Left, y, x2, y, size, c.DataColor, c.Opacity)
		deck.Text(x2+(textsize/2), y-size/2, fmt.Sprintf(format, v), "mono", textsize*0.75, c.ValueColor)
		y -= linespacing
	}
}

// ConditionalHBar makes a horizontal bar chart with conditional coloring
func (c *ChartBox) ConditionalHBar(deck *generate.Deck, size, linespacing float64, cmin, cmax float64, color string) {

	y := c.Top
	xmin := zerobase(c.Zerobased, c.Minvalue)
	for _, d := range c.Data {
		v := d.Value
		deck.TextEnd(c.Left-2, y-size/2, d.Label, "sans", c.TextSize, c.LabelColor)
		x2 := MapRange(v, xmin, c.Maxvalue, c.Left, c.Right)
		deck.Line(c.Left, y, x2, y, size, conditionalcolor(v, cmin, cmax, color, c.DataColor), c.Opacity)
		y -= linespacing
	}
}

// Line makes a line chart
func (c *ChartBox) Line(deck *generate.Deck, size float64) {
	n := len(c.Data)
	fn := float64(n - 1)
	ymin := zerobase(c.Zerobased, c.Minvalue)
	for i := 0; i < n-1; i++ {
		v1 := c.Data[i].Value
		v2 := c.Data[i+1].Value
		x1 := MapRange(float64(i), 0, fn, c.Left, c.Right)
		y1 := MapRange(v1, ymin, c.Maxvalue, c.Bottom, c.Top)
		x2 := MapRange(float64(i+1), 0, fn, c.Left, c.Right)
		y2 := MapRange(v2, ymin, c.Maxvalue, c.Bottom, c.Top)
		deck.Line(x1, y1, x2, y2, size, c.DataColor, c.Opacity)
	}
}

// ConditionalLine makes a line chart with conditional coloring
func (c *ChartBox) ConditionalLine(deck *generate.Deck, size float64, cmin, cmax float64, color string) {
	n := len(c.Data)
	fn := float64(n - 1)
	ymin := zerobase(c.Zerobased, c.Minvalue)
	for i := 0; i < n-1; i++ {
		v1 := c.Data[i].Value
		v2 := c.Data[i+1].Value
		x1 := MapRange(float64(i), 0, fn, c.Left, c.Right)
		y1 := MapRange(v1, ymin, c.Maxvalue, c.Bottom, c.Top)
		x2 := MapRange(float64(i+1), 0, fn, c.Left, c.Right)
		y2 := MapRange(v2, ymin, c.Maxvalue, c.Bottom, c.Top)
		deck.Line(x1, y1, x2, y2, size, conditionalcolor(v1, cmin, cmax, color, c.DataColor), c.Opacity)
	}
}

// Scatter makes a scatter chart
func (c *ChartBox) Scatter(deck *generate.Deck, size float64) {
	dlen := float64(len(c.Data) - 1)
	ymin := zerobase(c.Zerobased, c.Minvalue)
	for i, d := range c.Data {
		x := MapRange(float64(i), 0, dlen, c.Left, c.Right)
		y := MapRange(d.Value, ymin, c.Maxvalue, c.Bottom, c.Top)
		deck.Circle(x, y, size, c.DataColor, c.Opacity)
	}
}

// ConditionalScatter makes a scatter chart
func (c *ChartBox) ConditionalScatter(deck *generate.Deck, size float64, cmin, cmax float64, color string) {
	dlen := float64(len(c.Data) - 1)
	ymin := zerobase(c.Zerobased, c.Minvalue)
	for i, d := range c.Data {
		v := d.Value
		x := MapRange(float64(i), 0, dlen, c.Left, c.Right)
		y := MapRange(v, ymin, c.Maxvalue, c.Bottom, c.Top)
		deck.Circle(x, y, size, conditionalcolor(v, cmin, cmax, color, c.DataColor), c.Opacity)
	}
}

// Area makes a area chart
func (c *ChartBox) Area(deck *generate.Deck) {
	n := len(c.Data)
	fn := float64(n - 1)
	ymin := zerobase(c.Zerobased, c.Minvalue)
	xvol := make([]float64, n+2)
	yvol := make([]float64, n+2)
	xvol[0] = c.Left
	yvol[0] = c.Bottom
	xvol[n+1] = c.Right
	yvol[n+1] = c.Bottom

	for i := 0; i < n; i++ {
		xvol[i+1] = MapRange(float64(i), 0, fn, c.Left, c.Right)
		yvol[i+1] = MapRange(c.Data[i].Value, ymin, c.Maxvalue, c.Bottom, c.Top)
	}
	deck.Polygon(xvol, yvol, c.DataColor, c.Opacity)
}

// HDot makes a dotted horizontal bar chart
func (c *ChartBox) HDot(deck *generate.Deck, size, linespacing float64) {
	textsize := c.TextSize
	format := c.DataFormat
	y := c.Top
	xmin := zerobase(c.Zerobased, c.Minvalue)
	for _, d := range c.Data {
		deck.TextEnd(c.Left-textsize, y-size/2, d.Label, "sans", textsize, c.LabelColor)
		x2 := MapRange(d.Value, xmin, c.Maxvalue, c.Left, c.Right)
		deck.Text(x2+textsize/2, y-size/2, fmt.Sprintf(format, d.Value), "mono", textsize*0.75, c.ValueColor)
		dottedhline(deck, c.Left, y, x2, size, size*2, c.DataColor)
		y -= linespacing
	}
}

// VDot makes a vertical dotted bar chart
func (c *ChartBox) VDot(deck *generate.Deck, size float64) {
	dlen := float64(len(c.Data) - 1)
	ymin := zerobase(c.Zerobased, c.Minvalue)
	for i, d := range c.Data {
		x := MapRange(float64(i), 0, dlen, c.Left, c.Right)
		y := MapRange(d.Value, ymin, c.Maxvalue, c.Bottom, c.Top)
		dottedvline(deck, x, c.Bottom, y, 0.25, 1, c.DataColor)
		deck.Circle(x, y, size, c.DataColor, c.Opacity)
	}
}

// PMap makes a proportional map
func (c *ChartBox) PMap(deck *generate.Deck, pwidth, pmlen float64, showvalues, solid bool) {
	top := c.Top
	left := c.Left
	right := c.Right
	textsize := c.TextSize
	format := c.DataFormat

	x := left
	pl := (right - left)
	bl := pl / 100.0
	hspace := 0.10
	var ty float64
	var textcolor string

	data := c.Data
	for i, p := range pct(data) {
		bx := (p * bl)
		if p < 3 || float64(len(data[i].Label)) > pmlen {
			ty = top - pwidth*1.2
			deck.Line(x+(bx/2), ty+(textsize*1.5), x+(bx/2), top, 0.1, dotlinecolor)
		} else {
			ty = top
		}
		linecolor, lineop := stdcolor(i, data[i].Note, c.DataColor, p, solid)
		deck.Line(x, top, bx+x, top, pwidth, linecolor, lineop)
		if lineop == 100 {
			textcolor = "white"
		} else {
			textcolor = "black"
		}

		if showvalues {
			deck.TextMid(x+(bx/2), ty+(pwidth), data[i].Label, "sans", textsize*0.75, c.ValueColor)
		}
		deck.TextMid(x+(bx/2), ty-(textsize/2), fmt.Sprintf(format+"%%", p), "sans", textsize, textcolor)

		x += bx - hspace
	}
}

// Slope makes a slope chart
func (c *ChartBox) Slope(deck *generate.Deck, linewidth float64) {
	data := c.Data
	textsize := c.TextSize
	format := c.DataFormat

	if len(data) < 2 {
		fmt.Fprintf(os.Stderr, "slope graphs need at least two data pointextsize")
		return
	}
	ymin := zerobase(c.Zerobased, c.Minvalue)
	top := c.Top
	bottom := c.Bottom
	left := c.Left
	right := c.Right
	datacolor := c.DataColor
	lw := linewidth / 2
	lsize := textsize * 0.75
	tsize := textsize * 1.5
	w := right - left
	h := top - bottom

	// these are magical
	hskip := w * .60
	vskip := h * 1.4

	x1 := left
	x2 := right
	// Process the data in pairs
	for i := 0; i < len(data)-1; i += 2 {
		if len(data[i].Label) > 0 {
			deck.TextMid(x1+(w/2), top+(textsize/2), data[i].Note, "sans", tsize, c.LabelColor)
		}
		v1 := data[i].Value
		v2 := data[i+1].Value
		v1y := MapRange(v1, ymin, c.Maxvalue, bottom, top)
		v2y := MapRange(v2, ymin, c.Maxvalue, bottom, top)
		deck.Line(x1, bottom, x1, top, lw, "black")
		deck.Line(x2, bottom, x2, top, lw, "black")
		deck.Circle(x1, v1y, textsize, datacolor)
		deck.Circle(x2, v2y, textsize, datacolor)
		deck.Line(x1, v1y, x2, v2y, linewidth, datacolor)
		deck.TextMid(x1, bottom-2, data[i].Label, "sans", textsize, c.LabelColor)
		deck.TextMid(x2, bottom-2, data[i+1].Label, "sans", textsize, c.LabelColor)

		// only Show max value id user-specified
		if c.Zerobased {
			deck.TextEnd(x1-1, top, fmt.Sprintf(format, c.Maxvalue), "sans", lsize, c.LabelColor)
		}
		deck.TextEnd(x1-1, v1y, fmt.Sprintf(format, v1), "sans", lsize, c.LabelColor)
		deck.Text(x2+1, v2y, fmt.Sprintf(format, v2), "sans", lsize, c.LabelColor)
		x1 += w + hskip
		x2 += w + hskip
		if x2 > 100 {
			x1 = left
			x2 = right
			top -= vskip
			bottom -= vskip
		}
	}
}

// Donut makes donut and pie charts
func (c *ChartBox) Donut(deck *generate.Deck, psize, pwidth float64, showval, solid bool) {
	top := c.Top
	left := c.Left
	textsize := c.TextSize

	data := c.Data

	if left < 0 {
		left = 50.0
	}
	a1 := 0.0
	dx := left // + (psize / 2)
	dy := top - (psize / 2)

	for i, p := range pct(data) {
		angle := (p / 100) * 360.0
		a2 := a1 + angle
		mid := (a1 + a2) / 2

		bcolor, op := stdcolor(i, data[i].Note, c.DataColor, p, solid)
		deck.Arc(dx, dy, psize, psize, pwidth, a1, a2, bcolor, op)
		tx, ty := polar(dx, dy, psize*.85, mid*(math.Pi/180))
		if showval {
			deck.TextMid(tx, ty, fmt.Sprintf("%s "+c.DataFormat+"%%", data[i].Label, p), "sans", textsize, "")
		}
		a1 = a2
	}
}

// Radial makes a radial chart
func (c *ChartBox) Radial(deck *generate.Deck, psize, pwidth float64, showspokes, showvalues bool) {
	data := c.Data
	top := c.Top
	left := c.Left
	textsize := c.TextSize
	datacolor := c.DataColor
	if left < 0 {
		left = 50.0
	}
	dx := left
	dy := top

	t := topclock
	deck.Circle(dx, dy, pwidth*2, "silver", 10)
	step := fullcircle / float64(len(data))
	var color string
	for _, d := range data {
		cv := MapRange(d.Value, 0, c.Maxvalue, 2, psize)
		px, py := polar(dx, dy, pwidth, t)
		tx, ty := polar(dx, dy, pwidth+(psize/2)+(textsize*2), t)

		if len(d.Note) > 0 {
			color = d.Note
		} else {
			color = datacolor
		}
		deck.TextMid(tx, ty, d.Label, "sans", textsize/2, "black")
		if showvalues {
			deck.TextMid(px, py-textsize/3, fmt.Sprintf(c.DataFormat, d.Value), "mono", textsize, c.LabelColor)
		}
		if showspokes {
			spokes(deck, px, py, psize/2, 0.05, int(d.Value), color)
		} else {
			deck.Circle(px, py, cv, color, transparency)
			deck.Line(tx, ty, px, py, 0.05, "gray", 50)
		}
		t -= step
	}
}

// PGrid makes a proportional grid with the specified rows and columns
func (c *ChartBox) PGrid(deck *generate.Deck, linespacing float64, rows, cols int, showvalues bool) {
	textsize := c.TextSize
	data := c.Data
	top := c.Top
	left := c.Left
	format := c.DataFormat

	// sanity checks
	if left < 0 {
		left = 30.0
	}
	if rows*cols != 100 {
		return
	}
	sum := 0.0
	for _, d := range data {
		sum += d.Value
	}
	pct := make([]float64, len(data))
	for i, d := range data {
		pct[i] = math.Floor((d.Value / sum) * 100)
	}

	// encode the data in a string vector
	chars := make([]string, 100)
	cb := 0
	for k := 0; k < len(data); k++ {
		for l := 0; l < int(pct[k]); l++ {
			chars[cb] = data[k].Note
			cb++
		}
	}

	// make rows and cols
	n := 0
	y := top
	for i := 0; i < rows; i++ {
		x := left
		for j := 0; j < cols; j++ {
			if n >= 100 {
				break
			}
			deck.Circle(x, y, textsize, chars[n])
			n++
			x += linespacing
		}
		y -= linespacing
	}

	cx := (float64(cols-1) * linespacing) + linespacing/2
	for i, d := range data {
		y -= linespacing * 1.2
		deck.Circle(left, y, textsize, d.Note)
		deck.Text(left+textsize, y-(textsize/2), d.Label+" ("+fmt.Sprintf(format, pct[i])+"%)", "sans", textsize, "")
		if showvalues {
			deck.TextEnd(left+cx, y-(textsize/2), fmt.Sprintf(format, d.Value), "sans", textsize, c.ValueColor)
		}
	}
}

// axes

// YAxis makes the Y axis with optional grid lines
func (c *ChartBox) YAxis(deck *generate.Deck, min, max, step float64, gridlines bool) {
	w := c.Right - c.Left
	textsize := c.TextSize
	ymin := zerobase(c.Zerobased, c.Minvalue)
	for v := min; v <= max; v += step {
		y := MapRange(v, ymin, c.Maxvalue, c.Bottom, c.Top)
		if gridlines {
			deck.Line(c.Left, y, c.Left+w, y, 0.05, "gray")
		}
		deck.TextEnd(c.Left-2, y-(textsize/3), fmt.Sprintf(c.DataFormat, v), "sans", textsize, c.LabelColor, c.Opacity)
	}
}

// XLabel makes the x axis labels
func (c *ChartBox) XLabel(deck *generate.Deck, n int) {
	textsize := c.TextSize
	fn := float64(len(c.Data) - 1)
	for i, d := range c.Data {
		x := MapRange(float64(i), 0, fn, c.Left, c.Right)
		if i%n == 0 {
			deck.TextMid(x, c.Bottom-(textsize*2), d.Label, "sans", textsize, c.LabelColor, c.Opacity)
		}
	}
}

// XStaggerLabel makes staggered x axis labels
func (c *ChartBox) XStaggerLabel(deck *generate.Deck, n int) {
	textsize := c.TextSize
	fn := float64(len(c.Data) - 1)
	for i, d := range c.Data {
		x := MapRange(float64(i), 0, fn, c.Left, c.Right)
		if i%n == 0 {
			deck.TextMid(x, c.Bottom-(textsize*2), d.Label, "sans", textsize, c.LabelColor, c.Opacity)
		} else {
			deck.TextMid(x, c.Bottom-(textsize*4), d.Label, "sans", textsize, c.LabelColor, c.Opacity)
		}
	}
}

// XRotateLabel makes rotated x axis labels
func (c *ChartBox) XRotateLabel(deck *generate.Deck, angle float64, n int) {
	textsize := c.TextSize
	fn := float64(len(c.Data) - 1)
	for i, d := range c.Data {
		x := MapRange(float64(i), 0, fn, c.Left, c.Right)
		if i%n == 0 {
			deck.TextRotate(x, c.Bottom-(textsize*2), d.Label, "", "sans", angle, textsize, c.LabelColor, c.Opacity)
		}
	}
}

// chart accessories

// RegressionLine makes a regression line from a data set
func (c *ChartBox) RegressionLine(deck *generate.Deck, size float64) {
	top := c.Top
	left := c.Left
	bottom := c.Bottom
	right := c.Right
	lw := size
	x := make([]float64, len(c.Data))
	y := make([]float64, len(c.Data))
	for i, data := range c.Data {
		x[i] = float64(i)
		y[i] = data.Value
	}
	m, b := dataslope(x, y)
	dl := len(x) - 1
	l := float64(dl)
	x1 := x[0]
	x2 := x[dl]
	y1 := m*x1 + b
	y2 := m*x2 + b
	ymin := zerobase(c.Zerobased, c.Minvalue)
	rx1 := MapRange(x1, 0, l, left, right)
	rx2 := MapRange(x2, 0, l, left, right)
	ry1 := MapRange(y1, ymin, c.Maxvalue, bottom, top)
	ry2 := MapRange(y2, ymin, c.Maxvalue, bottom, top)
	deck.Line(rx1, ry1, rx2, ry2, lw, c.DataColor, c.Opacity)
}

// Values places chart values
func (c *ChartBox) Values(deck *generate.Deck, offset float64) {
	n := len(c.Data)
	fn := float64(n - 1)
	ymin := zerobase(c.Zerobased, c.Minvalue)
	for i := 0; i < n; i++ {
		v := c.Data[i].Value
		x := MapRange(float64(i), 0, fn, c.Left, c.Right)
		y := MapRange(v, ymin, c.Maxvalue, c.Bottom, c.Top)
		deck.TextMid(x, y+offset, fmt.Sprintf(c.DataFormat, v), "mono", c.TextSize, c.ValueColor, c.Opacity)
	}
}

// CTitle makes a centered title
func (c *ChartBox) CTitle(deck *generate.Deck, offset float64) {
	midx := c.Left + ((c.Right - c.Left) / 2)
	deck.TextMid(midx, c.Top+offset, c.Title, "sans", c.TextSize*2, c.DataColor, c.Opacity)
}

// Frame makes a filled frame with the specified opacity (0-100)
func (c *ChartBox) Frame(deck *generate.Deck, opacity float64) {
	w := c.Right - c.Left
	h := c.Top - c.Bottom
	deck.Rect(c.Left+w/2, c.Bottom+h/2, w, h, c.DataColor, opacity)
}

// Notes places notes
func (c *ChartBox) Notes(deck *generate.Deck, position string) {
	textsize := c.TextSize
	fn := float64(len(c.Data) - 1)
	ymin := zerobase(c.Zerobased, c.Minvalue)
	for i, data := range c.Data {
		x := MapRange(float64(i), 0, fn, c.Left, c.Right)
		y := MapRange(data.Value, ymin, c.Maxvalue, c.Bottom, c.Top)
		switch position {
		case "c":
			deck.TextMid(x, y, data.Note, "serif", textsize, c.LabelColor, c.Opacity)
		case "r":
			deck.TextEnd(x, y, data.Note, "serif", textsize, c.LabelColor, c.Opacity)
		case "l":
			deck.Text(x, y, data.Note, "serif", textsize, c.LabelColor, c.Opacity)
		default:
			deck.TextMid(x, y, data.Note, "serif", textsize, c.LabelColor, c.Opacity)
		}
	}
}

// LineNote places a note with a horizontal line set at a value
func (c *ChartBox) LineNote(deck *generate.Deck, v float64, s string, size float64) {
	ymin := zerobase(c.Zerobased, c.Minvalue)
	y := MapRange(v, ymin, c.Maxvalue, c.Bottom, c.Top)
	deck.Line(c.Left, y, c.Right, y, 0.1, c.DataColor, c.Opacity)
	if len(s) > 0 {
		deck.Text(c.Right+(size/2), y-(size/4), s, "serif", c.TextSize*0.75, c.DataColor, c.Opacity)
	}
}

// Grid makes a grid
func (c *ChartBox) Grid(deck *generate.Deck, size, step float64) {
	for x := c.Left; x <= c.Right; x += step {
		deck.Line(x, c.Bottom, x, c.Top, size, c.DataColor, c.Opacity)
	}
	for y := c.Bottom; y <= c.Top; y += step {
		deck.Line(c.Left, y, c.Right, y, size, c.DataColor, c.Opacity)
	}
}

// driver and I/O methods

// NewChart initializes the settings required to make a chart
// chartType may be one of: "line", "slope", "bar", "wbar", "hbar",
// "volume, "scatter", "donut", "pmap", "pgrid","radial"
func NewChart(chartType string, top, bottom, left, right float64) Settings {
	var s Settings

	switch chartType {
	case "bar":
		s.Flags.ShowBar = true
	case "wbar":
		s.Flags.ShowWBar = true
	case "hbar":
		s.Flags.ShowHBar = true
	case "donut":
		s.Flags.ShowDonut = true
	case "pmap":
		s.Flags.ShowPMap = true
	case "pgrid":
		s.Flags.ShowPGrid = true
	case "radial":
		s.Flags.ShowRadial = true
	case "line":
		s.Flags.ShowLine = true
	case "scatter":
		s.Flags.ShowScatter = true
	case "volume", "area":
		s.Flags.ShowVolume = true
	case "slope":
		s.Flags.ShowSlope = true
	}
	if left <= 0 {
		left = 10
	}
	if right <= 0 {
		right = 90
	}
	if top <= 0 {
		top = 90
	}
	if bottom <= 0 {
		bottom = 30
	}
	s.Measures.Left = left
	s.Measures.Right = right
	s.Measures.Top = top
	s.Measures.Bottom = bottom
	s.Measures.XLabelInterval = 1
	s.Measures.TextSize = 1.5
	s.Measures.LineSpacing = 2.4

	s.Attributes.BackgroundColor = "white"
	s.Attributes.DataColor = "lightsteelblue"
	s.Attributes.LabelColor = "rgb(75,75,75)"

	return s
}

// GenerateChart makes charts according to the orientation:
// horizontal bar or line, bar, dot, or donut volume charts
func (s *Settings) GenerateChart(deck *generate.Deck, r io.Reader) {
	f := s.Flags
	m := s.Measures
	a := s.Attributes
	var chart ChartBox
	var err error
	if f.ReadCSV {
		chart, err = ReadCSV(r, a.CSVCols)
	} else {
		chart, err = ReadTSV(r)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	clow, chigh, condcolor, err := parsecondition(a.DataCondition)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	chart.DataColor = a.DataColor
	chart.ValueColor = a.ValueColor
	chart.LabelColor = a.LabelColor
	switch {
	case f.ShowVDot:
		chart.VDot(deck, m.LineWidth)
	case f.ShowHBar:
		chart.ConditionalHBar(deck, m.BarWidth, m.LineSpacing, clow, chigh, condcolor)
	case f.ShowHDot:
		chart.HDot(deck, m.LineWidth, m.LineSpacing)
	case f.ShowWBar:
		chart.WBar(deck, m.LineSpacing, f.ShowValues, f.ShowPercentage)
	case f.ShowDonut:
		chart.Donut(deck, m.PSize, m.PWidth, f.ShowValues, f.SolidPMap)
	case f.ShowPMap:
		chart.PMap(deck, m.PWidth, m.PSize, f.ShowValues, f.SolidPMap)
	case f.ShowPGrid:
		chart.PGrid(deck, m.LineSpacing, 10, 10, f.ShowValues)
	case f.ShowRadial:
		chart.Radial(deck, m.PSize, m.PWidth, f.ShowSpokes, f.ShowValues)
	case f.ShowSlope:
		chart.Slope(deck, m.LineWidth)
	default:
		if m.BarWidth == 0 {
			m.BarWidth = (chart.Right - chart.Left) / float64(len(chart.Data)+1)
		}
		chart.ConditionalBar(deck, m.BarWidth, clow, chigh, condcolor)

		if f.ShowScatter {
			chart.ConditionalScatter(deck, m.LineWidth, clow, chigh, condcolor)
		}
		if f.ShowLine {
			chart.ConditionalLine(deck, m.LineWidth, clow, chigh, condcolor)
		}
		if f.ShowVolume {
			op := chart.Opacity
			chart.Opacity = m.VolumeOpacity
			chart.Area(deck)
			chart.Opacity = op
		}

		if f.ShowTitle {
			chart.DataColor = "black"
			chart.CTitle(deck, 5)
		}
		if f.ShowFrame {
			chart.Frame(deck, 10)
		}
		if m.XLabelInterval != 0 {
			chart.XRotateLabel(deck, m.XLabelRotation, m.XLabelInterval)
		}
		if f.ShowAxis {
			var ymin, ymax, ystep float64
			if a.YAxisR == "" {
				ymin, ymax, ystep = cyrange(zerobase(chart.Zerobased, chart.Minvalue), chart.Maxvalue, 5)
			} else {
				ymin, ymax, ystep = yrange(a.YAxisR)
			}
			chart.YAxis(deck, ymin, ymax, ystep, f.ShowGrid)
		}
	}
}

// helper functions

// MapRange maps the range (low1, high1) to (low2, high2)
func MapRange(value, low1, high1, low2, high2 float64) float64 {
	return low2 + (high2-low2)*(value-low1)/(high1-low1)
}

// mean computes the arithmetic mean of a set of data
func mean(x []float64) float64 {
	sum := 0.0
	n := len(x)
	for i := 0; i < n; i++ {
		sum += x[i]
	}
	return sum / float64(n)
}

// dataslope computes the slope (m, b) of a set of x, y points
func dataslope(x, y []float64) (float64, float64) {
	n := len(x) // assume x and y have the same length
	xy := make([]float64, n)
	for i := 0; i < n; i++ {
		xy[i] = x[i] * y[i]
	}
	sqx := make([]float64, n)
	for i := 0; i < n; i++ {
		sqx[i] = x[i] * x[i]
	}
	meanxy := mean(xy)
	meanx := mean(x)
	meany := mean(y)
	meanxsq := mean(sqx)

	rise := (meanxy - (meanx * meany))
	run := (meanxsq - (meanx * meanx))
	m := rise / run
	b := meany - (m * meanx)
	return m, b
}

// dottedvline makes dotted vertical line, using circles, with specified step
func dottedvline(deck *generate.Deck, x, y1, y2, dotsize, step float64, color string) {

	if y1 < y2 { // positive
		for y := y1; y <= y2; y += step {
			deck.Circle(x, y, dotsize, color)
		}
	} else { // negative
		for y := y2; y <= y1; y += step {
			deck.Circle(x, y, dotsize, color)
		}
	}
}

// dottedhline makes a dotted horizontal line, using circles with specified step and separation
func dottedhline(deck *generate.Deck, x1, y, x2, dotsize, step float64, color string) {
	for x := x1; x < x2; x += step {
		deck.Circle(x, y, dotsize, color)
		x += step
	}
}

// conditionalcolor chooses between two colors when the value falls between min and max
func conditionalcolor(value, min, max float64, trueColor, falseColor string) string {
	if value <= max && value >= min {
		return trueColor
	}
	return falseColor
}

// pct computs the percentage of a range of values
func pct(data []NameValue) []float64 {
	sum := 0.0
	for _, d := range data {
		sum += d.Value
	}

	p := make([]float64, len(data))
	for i, d := range data {
		p[i] = (d.Value / sum) * 100
	}
	return p
}

// stdcolor uses either the standard color (cycling through a list) or specified color and opacity
func stdcolor(i int, dcolor, color string, op float64, solid bool) (string, float64) {
	if color == "std" {
		return blue7[i%len(blue7)], 100
	}
	if len(dcolor) > 0 {
		if solid {
			return dcolor, 100
		}
		return dcolor, 40
	}
	return color, op
}

// polar converts polar to Cartesian coordinates
func polar(x, y, r, t float64) (float64, float64) {
	px := x + r*math.Cos(t)
	py := y + r*math.Sin(t)
	return px, py
}

// datasum computes the sum of the chart data
func datasum(data []NameValue) float64 {
	sum := 0.0
	for _, d := range data {
		sum += d.Value
	}
	return sum
}

// spokes makes the points and lines like spokes on a wheel
func spokes(deck *generate.Deck, cx, cy, r, spokesize float64, n int, color string) {
	t := topclock
	step := fullcircle / float64(n)
	for i := 0; i < n; i++ {
		px, py := polar(cx, cy, r, t)
		deck.Line(cx, cy, px, py, spokesize, "lightgray")
		deck.Circle(px, py, 0.5, color)
		t -= step
	}
}

// parsecondition parses the expression low,high,color. For example "0,10,red"
// means color the data red if the value is between 0 and 10.
func parsecondition(s string) (float64, float64, string, error) {
	if len(s) == 0 {
		return smallest, largest, "", nil
	}
	cs := strings.Split(s, ",")
	if len(cs) != 3 {
		return smallest, largest, "", fmt.Errorf("%s bad condition", s)
	}
	low, err := strconv.ParseFloat(cs[0], 64)
	if err != nil {
		return smallest, largest, "", err
	}
	high, err := strconv.ParseFloat(cs[1], 64)
	if err != nil {
		return smallest, largest, "", err
	}
	return low, high, cs[2], nil
}

// yrange parses the min, max, step for axis labels
func yrange(s string) (float64, float64, float64) {
	var min, max, step float64
	n, err := fmt.Sscanf(s, "%f,%f,%f", &min, &max, &step)
	if n != 3 || err != nil {
		return 0, 0, 0
	}
	return min, max, step
}

// cyrange computes "optimal" min, max, step for axis labels
// rounding the max to the appropriate number, given the number of labels
func cyrange(min, max float64, n int) (float64, float64, float64) {
	l := math.Log10(max)
	p := math.Pow10(int(l))
	pl := math.Ceil(max / p)
	ymax := pl * p
	return min, ymax, ymax / float64(n)
}
