package ui

import (
	"bytes"
	"eco/lib"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"github.com/wcharczuk/go-chart"
	"image/color"
)

type res struct {
	buffer *bytes.Buffer
}

func (r *res) Name() string {
	return "chart"
}

func (r *res) Content() []byte {
	return r.buffer.Bytes()
}

var productColor = color.RGBA{63, 195, 128, 1}

type GraphLine struct {
	Line *canvas.Line
	Next *canvas.Line
}

type Graph struct {
	product [][]int
	chart   chart.Chart

	report     lib.MarketReport
	lastReport *lib.MarketReport

	tick int

	app    fyne.App
	window fyne.Window
	canvas *fyne.Container
}

func NewGraph() *Graph {
	return &Graph{}
	a := app.New()
	w := a.NewWindow("eco")
	w.Resize(fyne.Size{800, 500})

	g := &Graph{
		chart:  chart.Chart{},
		app:    a,
		window: w,
	}
	g.canvas = fyne.NewContainerWithLayout(g)

	content := g.render(nil)
	w.SetContent(content)
	w.SetFixedSize(true)
	return g
}

func (g *Graph) Start() {
	c := make(chan bool)
	<-c
	//g.window.ShowAndRun()
}

func (g *Graph) render(new *lib.MarketReport) *fyne.Container {
	if new == nil {
		g.chart.Series = []chart.Series{
			chart.ContinuousSeries{
				XValues: []float64{0},
				YValues: []float64{0},
			},
		}
		g.product = [][]int{{0, 0}}
		return g.canvas
	}

	newValue := []int{g.tick, new.ProductSold}
	g.product = append(g.product, newValue)

	switch v := g.chart.Series[0].(type) {
	case chart.ContinuousSeries:
		v.XValues = append(v.XValues, float64(newValue[0]))
		v.YValues = append(v.YValues, float64(newValue[1]))
		g.chart.Series[0] = v
	}

	buffer := bytes.NewBuffer([]byte{})
	err := g.chart.Render(chart.PNG, buffer)
	if err != nil {
		panic(err)

	}

	r := res{buffer}
	img := canvas.NewImageFromResource(&r)
	g.window.SetContent(img)

	return g.canvas
}

func (g *Graph) Update(report lib.MarketReport) {
	return
	g.tick++

	g.lastReport = &g.report
	g.report = report

	lib.Log("NEW REPORT", report)
	g.render(&report)
	g.window.Canvas().Refresh(g.canvas)
}

func (g *Graph) Layout(_ []fyne.CanvasObject, size fyne.Size) {
}

func (g *Graph) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(600, 400)
}
