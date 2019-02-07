package main

import (
	"github.com/hassansin/androidpublisher/ui"
	"github.com/hassansin/gocui"
)

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}
	if v.Name() == sideView.Name() {
		sideView.MoveCursor(ui.Down)
	} else {
		v.MoveCursor(0, 1, false)
	}
	status.Reset()
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}
	if v.Name() == sideView.Name() {
		sideView.MoveCursor(ui.Up)
	} else {
		v.MoveCursor(0, -1, false)
	}
	status.Reset()
	return nil
}

func cursorLeft(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		v.MoveCursor(-1, 0, false)
	}
	return nil
}

func cursorRight(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		v.MoveCursor(1, 0, false)
	}
	return nil
}
