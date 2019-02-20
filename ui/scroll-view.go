package ui

import (
	"fmt"
	"math"

	"github.com/hassansin/gocui"
)

type Scrollbar struct {
	*gocui.View
	parentView *gocui.View
	g          *gocui.Gui
	name       string
}

func NewScrollbar(g *gocui.Gui, parent *gocui.View) *Scrollbar {
	v := &Scrollbar{
		g:          g,
		parentView: parent,
		name:       fmt.Sprintf("scrollbar-%v", r.Int()),
	}
	return v
}
func (m *Scrollbar) Name() string {
	return m.name
}

func (m *Scrollbar) height(bufferLines, viewHight int) int {
	multipler := 1
	for {
		if multipler*viewHight > bufferLines {
			break
		}
		multipler++
	}

	//scrollHeight := int(math.Ceil(float64(viewHight) * float64(viewHight) / float64(bufferLines)))
	scrollHeight := int(math.Ceil(float64(viewHight) / float64(multipler)))
	if scrollHeight >= viewHight {
		return 0
	}
	if scrollHeight == 1 {
		return 2
	}
	return scrollHeight
}

func (m *Scrollbar) pos(bufferOrigin, bufferLines, viewHight int) int {
	return bufferOrigin * viewHight / (bufferLines - viewHight)
}

func (m *Scrollbar) Redraw() error {
	l := len(m.parentView.ViewBufferLines())
	_, h := m.parentView.Size()
	_, oy := m.parentView.Origin()

	height := m.height(l, h)
	if height == 0 {
		return deleteView(m.g, m.name)
	}
	offset := m.pos(oy, l, h)

	_, y0, maxX, maxY := m.parentView.Coordinates()

	if offset+height > maxY {
		offset = maxY - height
	}

	if v, err := m.g.SetView(m.name, maxX-2, y0+offset, maxX, y0+offset+height); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.BgColor = gocui.ColorDefault | gocui.AttrReverse
		v.FgColor = gocui.ColorDefault | gocui.AttrReverse
		v.Frame = false
		m.View = v
		return nil
	}
	return nil
}
