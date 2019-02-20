package ui

import (
	"fmt"
	"strings"

	"github.com/hassansin/gocui"
	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
)

//Node - represents an item in Tree view
type Node interface {
	Title() string
	Children() []Node
}

const (
	pipe   = "│ "
	middle = "├─"
	last   = "└─"
)

//TreeView - generates a directory tree-like view
type TreeView struct {
	*gocui.View
	nodes []Node
	g     *gocui.Gui
	name  string
}

//NewTreeView - constructor for TreeView
func NewTreeView(g *gocui.Gui) *TreeView {
	return &TreeView{
		g:    g,
		name: fmt.Sprintf("tree-%v", r.Int()),
	}
}

//Name return view name
func (m *TreeView) Name() string {
	return m.name
}

//SetView initializes the view
func (m *TreeView) SetView(title string, nodes []Node) error {
	m.nodes = nodes
	_, maxY := m.g.Size()
	if v, err := m.g.SetView(m.Name(), 0, 0, 29, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = true
		v.Frame = true
		v.Title = title
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		v.HideCursor = true
		fmt.Fprint(v, strings.TrimSpace(Tree(nodes, "")))
		if _, err := m.g.SetCurrentView(m.name); err != nil {
			return err
		}
		m.View = v
	}
	return nil
}
func (m *TreeView) SetKeybinding() error {
	name := m.name
	if err := m.g.SetKeybinding(name, gocui.KeyArrowDown, gocui.ModNone, m.cursorDown); err != nil {
		return err
	}
	if err := m.g.SetKeybinding(name, 'j', gocui.ModNone, m.cursorDown); err != nil {
		return err
	}
	if err := m.g.SetKeybinding(name, gocui.KeyArrowUp, gocui.ModNone, m.cursorUp); err != nil {
		return err
	}
	if err := m.g.SetKeybinding(name, 'k', gocui.ModNone, m.cursorUp); err != nil {
		return err
	}
	return nil
}

//SetCurrent sets this view as current
func (m *TreeView) SetCurrent() error {
	_, err := m.g.SetCurrentView(m.View.Name())
	return errors.Wrap(err, "failed to current view")
}

//Selected returns items that are currently selected/hightlighted in the view
func (m TreeView) Selected() []int {
	_, y := m.View.Cursor()
	indexes := Selected(m.nodes, &y)
	for i, j := 0, len(indexes)-1; i < j; i, j = i+1, j-1 {
		indexes[i], indexes[j] = indexes[j], indexes[i]
	}
	return indexes
}

//moveCursor moves cursor up/down in the tree view
func (m *TreeView) moveCursor(dir bool) error {
	cx, cy := m.View.Cursor()
	ox, oy := m.View.Origin()
	_, vh := m.View.Size()
	lines := len(m.View.BufferLines())
	if dir {
		if cy+oy == 0 {
			oy = lines - vh
			if oy < 0 {
				oy = 0
			}
			if err := m.View.SetOrigin(ox, oy); err != nil {
				return err
			}
			cy = vh - 1
			if cy >= lines {
				cy = lines - 1
			}
			return m.View.SetCursor(cx, cy)
		}
		m.View.MoveCursor(0, -1, false)
	} else {
		if oy+cy == lines-1 {
			if err := m.View.SetCursor(cx, 0); err != nil {
				return err
			}
			return m.View.SetOrigin(ox, 0)
		}
		m.View.MoveCursor(0, 1, false)
	}
	return nil
}
func (m *TreeView) cursorDown(g *gocui.Gui, v *gocui.View) error {
	return m.moveCursor(false)
}

func (m *TreeView) cursorUp(g *gocui.Gui, v *gocui.View) error {
	return m.moveCursor(true)
}

func Tree(nodes []Node, prefix string) string {
	var result string
	for i, node := range nodes {
		isLast := i == len(nodes)-1
		prefixChild := prefix
		if isLast {
			result += prefix + last
			prefixChild += "  "
		} else {
			result += prefix + middle
			prefixChild += pipe
		}
		if len(node.Children()) == 0 {
			result += fmt.Sprintf("%v\n", aurora.Bold(node.Title()))
		} else {
			result += node.Title() + "\n"
		}
		result += Tree(node.Children(), prefixChild)
	}
	return result
}

func Selected(nodes []Node, idx *int) []int {
	var result []int
	for i, node := range nodes {
		if *idx == 0 {
			result = append(result, i)
			return result
		}
		*idx--
		result = append(result, Selected(node.Children(), idx)...)
		if len(result) > 0 {
			return append(result, i)
		}
	}
	return result
}
