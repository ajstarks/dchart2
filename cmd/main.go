package main

import (
	"fmt"
	"os"

	"github.com/ajstarks/dchart2"
	"github.com/ajstarks/deck/generate"
)

func openchart(filename string, cols ...string) (dchart2.ChartBox, error) {
	var chart dchart2.ChartBox
	r, err := os.Open(filename)
	if err != nil {
		return chart, err
	}
	defer r.Close()
	if len(cols) > 0 {
		return dchart2.ReadCSV(r, cols[0])
	}
	return dchart2.ReadTSV(r)
}

func main() {
	deck := generate.NewSlides(os.Stdout, 0, 0)
	appld, err := openchart("AAPL.d")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	applcsv, err := openchart("AAPL.csv", "Date,Close")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	browser, err := openchart("browser.d")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	browser2, err := openchart("browser2.d")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	incar, err := openchart("incar.d")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	slope, err := openchart("slope1.d")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	count, err := openchart("count.d")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	deck.StartDeck()

	// initial
	deck.StartSlide()
	appld.CTitle(deck, 4)
	appld.Bar(deck, 3)
	appld.DataColor = "blue"
	appld.Line(deck, 0.05)
	appld.RegressionLine(deck, 0.2)
	appld.DataColor = "red"
	appld.Scatter(deck, 2)
	appld.DataColor = "orange"
	appld.Opacity = 30
	appld.Area(deck)
	appld.Opacity = 100
	appld.DataFormat = "%.2f"
	appld.Values(deck, 2)
	appld.XStaggerLabel(deck, 2)
	appld.Notes(deck, "c")
	appld.DataFormat = "$ %0.f"
	appld.YAxis(deck, 0, 300, 50, true)
	appld.DataColor = "green"
	appld.LineNote(deck, 200, "More Money", 2)
	appld.DataColor = "red"
	appld.LineNote(deck, 50, "I'm Poor", 2)

	appld.Top = 40
	appld.Bottom = 20
	appld.DataColor = "purple"
	appld.Frame(deck, 10)
	appld.Opacity = 20
	appld.ConditionalLine(deck, 0.2, 200, 300, "green")
	appld.ConditionalScatter(deck, 2, 250, 295, "blue")

	appld.ConditionalBar(deck, 1, 50, 200, "orange")
	appld.VDot(deck, 1)
	appld.Opacity = 100
	appld.DataColor = "gray"
	appld.XRotateLabel(deck, 270, 1)
	appld.YAxis(deck, 0, 300, 50, false)
	deck.EndSlide()

	// repeat with CSV
	deck.StartSlide()
	applcsv.CTitle(deck, 4)
	applcsv.Bar(deck, 3)
	applcsv.DataColor = "blue"
	applcsv.Line(deck, 0.05)
	applcsv.RegressionLine(deck, 0.2)
	applcsv.DataColor = "red"
	applcsv.Scatter(deck, 2)
	applcsv.DataColor = "orange"
	applcsv.Opacity = 30
	applcsv.Area(deck)
	applcsv.Opacity = 100
	applcsv.DataColor = "black"
	applcsv.DataFormat = "%.2f"
	applcsv.Values(deck, 2)
	applcsv.XStaggerLabel(deck, 2)
	applcsv.Notes(deck, "c")
	applcsv.DataFormat = "$ %.0f"
	applcsv.YAxis(deck, 0, 300, 50, true)
	applcsv.DataColor = "green"
	applcsv.LineNote(deck, 200, "More Money", 2)
	applcsv.DataColor = "red"
	applcsv.LineNote(deck, 50, "I'm Poor", 2)
	applcsv.DataColor = "red"
	applcsv.Opacity = 50
	applcsv.Grid(deck, 0.1, 2)
	deck.EndSlide()

	// Horizontals
	deck.StartSlide()
	browser.Top = 90
	browser.Left = 15
	browser.Right = 80
	browser.DataColor = "steelblue"
	browser.WBar(deck, 4, true, true)

	applcsv.Top = 50
	applcsv.Bottom = 5
	applcsv.Left = 15
	applcsv.Right = applcsv.Left + 25
	applcsv.DataColor = "purple"
	applcsv.HBar(deck, 1, 3)

	applcsv.Left += 45
	applcsv.Right += 45
	applcsv.HDot(deck, 0.3, 3)
	deck.EndSlide()

	// PMaps
	pmh := 5.0
	deck.StartSlide()
	browser.Left = 5
	browser.Top = 80
	browser.Right = 90
	browser.PMap(deck, pmh, 60, true, true)
	browser.Top -= 20
	browser.PMap(deck, pmh, 60, true, false)
	browser2.Top = browser.Top - 20
	browser2.Left = browser.Left
	browser2.DataColor = "steelblue"
	browser2.PMap(deck, pmh, 60, true, false)
	deck.EndSlide()

	// pgrid and slope
	deck.StartSlide()
	incar.Top = 80
	incar.PGrid(deck, 3.0, 10, 10, false)
	slope.Top = 80
	slope.Bottom = 30
	slope.Left = 50
	slope.Right = 80
	slope.DataColor = "steelblue"
	slope.Slope(deck, 0.2)
	deck.EndSlide()

	// donut and radial
	deck.StartSlide()
	browser.Left = 20
	browser.Top = 60
	browser.Donut(deck, 20, 2, true, true)

	count.Top = 50
	count.Left = browser.Left + 50
	count.DataFormat = "%.0f"
	count.LabelColor = "black"
	count.Radial(deck, 4, 18, false, true)
	deck.EndSlide()

	// composition
	deck.StartSlide()
	t := 90.0
	l := 5.0
	w := 40.0
	h := 20.0
	vs := h + 10
	hs := w + 10

	applcsv.Left = l
	applcsv.Right = applcsv.Left + w
	applcsv.Top = t
	applcsv.Bottom = applcsv.Top - h

	center1 := applcsv.Left + (w / 2)
	applcsv.DataColor = "lightsteelblue"
	deck.TextMid(center1, applcsv.Top-(h/2), "Bar", "sans", 3, "Black", 50)
	applcsv.Bar(deck, 0.75)
	applcsv.Top -= vs
	applcsv.Bottom -= vs

	deck.TextMid(center1, applcsv.Top-(h/2), "Line", "sans", 3, "Black", 50)
	applcsv.Line(deck, 0.2)
	applcsv.Top -= vs
	applcsv.Bottom -= vs

	deck.TextMid(center1, applcsv.Top-(h/2), "Scatter", "sans", 3, "Black", 50)
	applcsv.Scatter(deck, 1)

	applcsv.Top = t
	applcsv.Bottom = applcsv.Top - h
	applcsv.Left += hs
	applcsv.Right += hs
	applcsv.DataColor = "gray"

	center2 := applcsv.Left + (w / 2)
	deck.TextMid(center2, applcsv.Top-(h/2), "XLabel", "sans", 3, "Black", 50)
	applcsv.XLabel(deck, 2)
	applcsv.Top -= vs
	applcsv.Bottom -= vs

	deck.TextMid(center2, applcsv.Top-(h/2), "YAxis", "sans", 3, "Black", 50)
	applcsv.YAxis(deck, 0, 290, 50, true)
	applcsv.Top -= vs
	applcsv.Bottom -= vs

	deck.TextMid(center2, applcsv.Top-(h/2), "Frame", "sans", 3, "Black", 50)
	applcsv.Frame(deck, 10)
	deck.EndSlide()

	// composite
	deck.StartSlide()
	applcsv.Top = 80
	applcsv.Bottom = 40
	applcsv.Left = 20
	applcsv.Right = 80
	applcsv.Opacity = 100
	applcsv.DataColor = "lightsteelblue"
	applcsv.Bar(deck, 0.75)
	applcsv.Line(deck, 0.2)
	applcsv.Scatter(deck, 1)
	applcsv.DataColor = "gray"
	applcsv.XLabel(deck, 2)
	applcsv.DataFormat = "$ %0.f"
	applcsv.YAxis(deck, 0, 290, 50, true)
	applcsv.Frame(deck, 10)
	deck.EndSlide()

	deck.EndDeck()
}
