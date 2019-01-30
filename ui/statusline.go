package ui

import (
	"fmt"

	"github.com/hassansin/gocui"
	"github.com/pkg/errors"
)

const StatusView = "status-line"

type StatusLine struct {
	g   *gocui.Gui
	v   *gocui.View
	msg string
}

//SetStatusLine returns new StatusLine
func SetStatusLine(g *gocui.Gui, msg string) (*StatusLine, error) {
	maxX, maxY := g.Size()
	height := 2
	if v, err := g.SetView(StatusView, -1, maxY-height, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return nil, errors.Wrap(err, "unable to create status line view")
		}
		v.Frame = false
		//v.BgColor = gocui.ColorDefault | gocui.AttrReverse
		//v.FgColor = gocui.ColorDefault | gocui.AttrReverse
		fmt.Fprintln(v, msg)
		if _, err = g.SetViewOnTop(v.Name()); err != nil {
			return nil, errors.Wrap(err, "unable to set top view")
		}
		return &StatusLine{
			v:   v,
			g:   g,
			msg: msg,
		}, nil
	}
	return nil, nil
}

func (s *StatusLine) Update(msg string) {
	s.g.Update(func(g *gocui.Gui) error {
		s.v.Clear()
		fmt.Fprint(s.v, msg)
		return nil
	})
}

func (s *StatusLine) Reset() {
	s.Update(s.msg)
}
