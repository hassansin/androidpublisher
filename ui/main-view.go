package ui

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/atotto/clipboard"
	"github.com/hassansin/androidpublisher/movements"
	"github.com/hassansin/gocui"
	"github.com/nwidger/jsoncolor"
	"github.com/pkg/errors"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type MainView struct {
	*gocui.View
	scrollbarView *Scrollbar
	g             *gocui.Gui
	name          string
	body          interface{}
	onSave        func(string, error)
}

func NewMainView(g *gocui.Gui) *MainView {
	v := &MainView{
		g:    g,
		name: fmt.Sprintf("main-%v", r.Int()),
	}
	return v
}
func (m *MainView) Name() string {
	return m.name
}
func (m *MainView) SetView() error {
	maxX, maxY := m.g.Size()
	if v, err := m.g.SetView(m.Name(), 30, 0, maxX-1, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Wrap = true
		v.Title = "Response"
		m.View = v
		m.scrollbarView = NewScrollbar(m.g, v)
		return nil
	}
	return nil
}

func (m *MainView) SetKeybinding() error {
	name := m.name
	if err := m.g.SetKeybinding(name, gocui.KeyArrowDown, gocui.ModNone, m.moveWithScroll(movements.CursorDown)); err != nil {
		return err
	}
	if err := m.g.SetKeybinding(name, gocui.KeyArrowUp, gocui.ModNone, m.moveWithScroll(movements.CursorUp)); err != nil {
		return err
	}
	if err := m.g.SetKeybinding(name, gocui.KeyArrowLeft, gocui.ModNone, m.moveWithScroll(movements.CursorLeft)); err != nil {
		return err
	}
	if err := m.g.SetKeybinding(name, gocui.KeyArrowRight, gocui.ModNone, m.moveWithScroll(movements.CursorRight)); err != nil {
		return err
	}
	if err := m.g.SetKeybinding(name, gocui.KeyPgdn, gocui.ModNone, m.moveWithScroll(movements.PgDn)); err != nil {
		return err
	}
	if err := m.g.SetKeybinding(name, gocui.KeyPgup, gocui.ModNone, m.moveWithScroll(movements.PgUp)); err != nil {
		return err
	}
	if err := m.g.SetKeybinding(name, gocui.KeyCtrlH, gocui.ModNone, m.moveWithScroll(movements.Home)); err != nil {
		return err
	}
	if err := m.g.SetKeybinding(name, gocui.KeyCtrlE, gocui.ModNone, m.moveWithScroll(movements.End)); err != nil {
		return err
	}
	if err := m.g.SetKeybinding("", gocui.KeyCtrlS, gocui.ModNone, m.saveDialog); err != nil {
		return err
	}
	if err := m.g.SetKeybinding("", gocui.KeyCtrlX, gocui.ModNone, m.copyToClipboard); err != nil {
		return err
	}
	return nil
}

func (m *MainView) moveWithScroll(fn func(*gocui.Gui, *gocui.View) error) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if err := fn(g, v); err != nil {
			return err
		}
		return m.scrollbarView.Redraw()
	}
}

func (m *MainView) SetCurrent() error {
	_, err := m.g.SetCurrentView(m.View.Name())
	return errors.Wrap(err, "failed to current view to main")
}

func (m *MainView) LoadContent(name string, res interface{}) {
	var result string
	if err, ok := res.(error); ok {
		result = err.Error()
	} else {
		body, err := jsoncolor.MarshalIndent(res, "", " ")
		if err != nil {
			result = err.Error()
		} else {
			result = string(body)
		}
	}
	m.View.Title = fmt.Sprintf("Response(%v)", name)
	m.View.Clear()
	m.View.SetCursor(0, 0)
	m.View.SetOrigin(0, 0)
	fmt.Fprintf(m.View, result)
	m.g.Update(func(g *gocui.Gui) error {
		return m.scrollbarView.Redraw()
	})
}

func (m *MainView) SaveContent(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, m.View)
	return err
}

//OnSave binds function to be called when save requested
func (m *MainView) OnSave(fn func(string, error)) *MainView {
	m.onSave = fn
	return m
}

func (m *MainView) saveDialog(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()
	currentView := g.CurrentView()
	f, err := NewForm(g, "Save Response", maxX/2-20, maxY/2)
	if err != nil {
		return err
	}
	f.OnCancel(func() error {
		_, err := g.SetCurrentView(currentView.Name())
		return err
	})
	var filename string
	f.OnSubmit(func() error {
		if filename == "" {
			return nil
		}
		if err := m.SaveContent(filename); err != nil {
			m.onSave("", err)
			return nil
		}
		fullPath, err := filepath.Abs(filename)
		if err != nil {
			return err
		}
		m.onSave(fullPath, nil)
		_, err = g.SetCurrentView(currentView.Name())
		return err
	})
	return f.Input(NewInput("File Name", &filename, 40, true))
}

func (m *MainView) copyToClipboard(g *gocui.Gui, v *gocui.View) error {
	//@TODO update status line
	clipboard.WriteAll(m.View.Buffer())
	return nil
}
