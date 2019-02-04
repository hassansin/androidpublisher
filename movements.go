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

func pageDown(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}
	internalBuffer := len(v.BufferLines())
	_, vh := v.Size()
	cx, cy := v.Cursor()
	ox, oy := v.Origin()

	if cy != 0 {
		if oy+(vh-3) >= (internalBuffer - vh) {
			v.SetOrigin(ox, internalBuffer-vh)
		} else {
			v.SetOrigin(ox, oy+(vh-3))
		}
	}
	if oy+vh == internalBuffer {
		v.SetCursor(cx, vh)
	} else {
		v.SetCursor(cx, vh-2)
	}
	return nil
}

func pageUp(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}
	internalBuffer := len(v.BufferLines())
	_, vh := v.Size()
	cx, _ := v.Cursor()
	ox, oy := v.Origin()

	if oy+vh != internalBuffer {
		v.SetOrigin(ox, oy-(vh-3))
	}
	if oy == 0 {
		v.SetCursor(cx, 0)
	} else {
		v.SetCursor(cx, 2)
	}
	return nil
}
