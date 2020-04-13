package ui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	//"image/color"
)

type Graph struct {
	Container canvas.Rectangle
	Lines     []*canvas.Line

	App    fyne.App
	Window fyne.Window
}

func NewGraph() *Graph {
	a := app.New()
	w := a.NewWindow("eco")
	w.Resize(fyne.Size{800, 500})
	return &Graph{
		App:    a,
		Window: w,
	}
}

func (g *Graph) AddLine(line *canvas.Line) {
	g.Lines = append(g.Lines, line)
	g.Window.SetContent(line)
}

func (g *Graph) Update() {
	for i := range g.Lines {
		g.Lines[i].Refresh()
	}
}
