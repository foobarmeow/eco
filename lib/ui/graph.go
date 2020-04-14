package ui

import (
	"eco/lib"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	//"image/color"
)

type Graph struct {
	price      canvas.Line
	priceFloat float64

	tick int

	canvas fyne.CanvasObject

	app    fyne.App
	Window fyne.Window
}

func NewGraph() *Graph {
	a := app.New()
	w := a.NewWindow("eco")
	w.Resize(fyne.Size{800, 500})

	g := &Graph{
		app:    a,
		Window: w,
	}

	content := g.render()
	w.SetContent(content)
	return g
}

func (g *Graph) Start() {
	g.Window.ShowAndRun()
}

func (g *Graph) render() *fyne.Container {
	container := fyne.NewContainer(&g.price)
	container.Layout = g
	g.canvas = container
	return container

}

func (g *Graph) Update(report lib.MarketReport) {
	g.tick++
	g.priceFloat = report.AveragePrice
	g.Layout(nil, g.canvas.Size())
}

func (g *Graph) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	bottomLeft := fyne.Position{0, size.Height}
	g.price.Position1 = bottomLeft
	g.price.Position2 = fyne.Position{g.tick, size.Height - int(g.priceFloat)}
}

func (g *Graph) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(200, 200)
}
