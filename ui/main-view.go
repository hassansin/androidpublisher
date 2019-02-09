package ui

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/hassansin/gocui"
	"github.com/nwidger/jsoncolor"
	"github.com/pkg/errors"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type MainView struct {
	*gocui.View
	g    *gocui.Gui
	name string
	body interface{}
}

func NewMainView(g *gocui.Gui) *MainView {
	return &MainView{
		g:    g,
		name: fmt.Sprintf("main-%v", r.Int()),
	}
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
		return nil
	}
	return nil
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
}

func (m *MainView) SaveContent(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, m.View)
	return err
}
