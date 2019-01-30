package main

import (
	"github.com/hassansin/gocui"
)

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		v.MoveCursor(0, 1, false)
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		v.MoveCursor(0, -1, false)
	}
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

//@TODO
func cursorEnd(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		lines := len(v.BufferLines())
		x, y0 := v.Cursor()
		_, y := v.Size()
		if err := v.SetCursor(x, y-y0); err != nil {
			ox, _ := v.Origin()
			if err := v.SetOrigin(ox, lines-y); err != nil {
				return err
			}
		}
	}
	return nil
}
