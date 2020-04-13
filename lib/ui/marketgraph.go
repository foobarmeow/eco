package ui

import (
	"eco/lib"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"image/color"
)

const ()

type MarketGraph struct {
	Graph    *Graph
	SoldLine *canvas.Line
	Tick     int
}

func NewMarketGraph() *MarketGraph {
	g := NewGraph()
	return &MarketGraph{
		Graph: g,
	}
}

func (m *MarketGraph) Start() {
	m.Graph.Window.ShowAndRun()
}

func (m *MarketGraph) Update(report lib.MarketReport) {
	m.Tick++
	if m.SoldLine == nil {
		m.SoldLine = canvas.NewLine(color.CMYK{0, 0, 0, 0})
		m.SoldLine.Position1 = fyne.Position{0, 0}
		m.Graph.AddLine(m.SoldLine)
	}

	m.SoldLine.Position2 = fyne.Position{m.Tick, report.ProductSold}
	m.Graph.Update()
}
