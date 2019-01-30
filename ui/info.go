package ui

import (
	"fmt"
	"time"

	"github.com/hassansin/gocui"
)

func ShowInfo(g *gocui.Gui, msg string, timeout time.Duration) func() {
	maxX, maxY := g.Size()
	width := 50
	if len(msg) > 50 {
		msg = msg[:45] + "..."
	}

	close := func() {
		g.Update(func(g *gocui.Gui) error {
			return deleteView(g, "infobox")
		})
	}

	g.Update(func(g *gocui.Gui) error {
		v, err := g.SetView("infobox", maxX/2-width/2, maxY/2, maxX/2+width/2, maxY/2+2)
		if err != nil && err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, msg)
		return nil
	})
	if timeout > 0 {
		go func() {
			time.Sleep(timeout)
			close()
		}()
	}
	return close
}
