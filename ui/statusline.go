package ui

import (
	"fmt"

	"github.com/hassansin/gocui"
	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
)

const StatusView = "status-line"

type StatusLine struct {
	g      *gocui.Gui
	v      *gocui.View
	height int
	msg    string
}

func NewStatusLine(g *gocui.Gui) *StatusLine {
	return &StatusLine{
		g:      g,
		height: 2,
	}
}

//SetStatusLine returns new StatusLine
func (s *StatusLine) SetView(msg string) error {
	maxX, maxY := s.g.Size()
	if v, err := s.g.SetView(StatusView, -1, maxY-s.height, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Wrap(err, "unable to create status line view")
		}
		v.Frame = false
		//v.BgColor = gocui.ColorDefault | gocui.AttrReverse
		//v.FgColor = gocui.ColorDefault | gocui.AttrReverse
		fmt.Fprintln(v, msg)
		if _, err = s.g.SetViewOnTop(v.Name()); err != nil {
			return errors.Wrap(err, "unable to set top view")
		}
		s.v = v
		s.msg = msg
		return nil
	}
	return nil
}

func (s *StatusLine) Update(msg string) {
	s.g.Update(func(g *gocui.Gui) error {
		s.v.Clear()
		fmt.Fprint(s.v, msg)
		return nil
	})
}

func (s *StatusLine) UpdateSuccess(msg string) {
	s.Update(fmt.Sprint(aurora.Green(msg)))
}
func (s *StatusLine) UpdateError(msg string) {
	s.Update(fmt.Sprint(aurora.Red(msg)))
}

func (s *StatusLine) Reset() {
	s.Update(s.msg)
}
