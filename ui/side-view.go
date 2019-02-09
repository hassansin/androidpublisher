package ui

import (
	"fmt"
	"strings"

	"github.com/hassansin/gocui"
	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
)

type Node interface {
	Title() string
	Children() []Node
}

type Dir bool

const (
	pipe       = "│ "
	middle     = "├─"
	last       = "└─"
	Up     Dir = true
	Down   Dir = false
)

type TreeView struct {
	*gocui.View
	nodes []Node
	g     *gocui.Gui
	name  string
}

func NewTreeView(g *gocui.Gui) *TreeView {
	return &TreeView{
		g:    g,
		name: fmt.Sprintf("tree-%v", r.Int()),
	}
}
func (m *TreeView) Name() string {
	return m.name
}

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
		fmt.Fprint(v, strings.TrimSpace(Tree(nodes, "")))
		if _, err := m.g.SetCurrentView(m.name); err != nil {
			return err
		}
		m.View = v
	}
	return nil
}

func (m *TreeView) SetCurrent() error {
	_, err := m.g.SetCurrentView(m.View.Name())
	return errors.Wrap(err, "failed to current view")
}

func (m TreeView) Selected() []int {
	_, y := m.View.Cursor()
	indexes := Selected(m.nodes, &y)
	for i, j := 0, len(indexes)-1; i < j; i, j = i+1, j-1 {
		indexes[i], indexes[j] = indexes[j], indexes[i]
	}
	return indexes
}

func (m *TreeView) MoveCursor(dir Dir) {
	x, y := m.View.Cursor()
	lines := len(m.View.BufferLines())
	if dir == Up {
		if y == 0 {
			m.View.SetCursor(x, lines-1)
			return
		}
		m.View.MoveCursor(0, -1, false)
	} else {
		if y == lines-1 {
			m.View.SetCursor(x, 0)
			return
		}
		m.View.MoveCursor(0, 1, false)
	}
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
