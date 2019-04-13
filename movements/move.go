package movements

import (
	"fmt"
	"strings"

	"github.com/hassansin/gocui"
)

func PgDn(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}
	internalBuffer := len(v.ViewBufferLines())
	_, vh := v.Size()
	//only move cursor for multi-line view
	if vh == 1 {
		return nil
	}
	cx, cy := v.Cursor()
	ox, oy := v.Origin()

	if cy != 0 {
		if oy+(vh) >= (internalBuffer - vh) {
			v.SetOrigin(ox, internalBuffer-vh)
		} else {
			v.SetOrigin(ox, oy+vh)
		}
	}
	v.SetCursor(cx, vh-1)
	return nil
}

func PgUp(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}
	_, vh := v.Size()
	//only move cursor for multi-line view
	if vh == 1 {
		return nil
	}
	cx, cy := v.Cursor()
	ox, oy := v.Origin()

	if cy == 0 {
		if oy < vh {
			v.SetOrigin(ox, 0)
		} else {
			v.SetOrigin(ox, oy-vh)
		}
	}
	v.SetCursor(cx, 0)
	return nil
}

func Home(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}
	v.SetCursor(0, 0)
	v.SetOrigin(0, 0)
	return nil
}

func End(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}
	_, vh := v.Size()
	buffer := v.ViewBufferLines()
	v.SetCursor(len(buffer[len(buffer)-1])-1, vh-1)
	v.SetOrigin(0, len(buffer)-vh)
	return nil
}

func CursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		v.MoveCursor(0, 1, false)
	}
	return nil
}

func CursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		v.MoveCursor(0, -1, false)
	}
	return nil
}

func CursorLeft(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		v.MoveCursor(-1, 0, false)
	}
	return nil
}

func CursorRight(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		v.MoveCursor(1, 0, false)
	}
	return nil
}

func DeleteTheRestOfTheLine(g *gocui.Gui, v *gocui.View) error {
	x, y := v.Cursor()
	lines := v.BufferLines()
	if len(lines) < y {
		return nil
	}
	lines[y] = lines[y][:x]

	v.Clear()
	fmt.Fprint(v, strings.Join(lines, "\n"))
	return nil
}
