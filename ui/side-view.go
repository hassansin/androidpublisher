package ui

import (
	"fmt"
	"strings"

	"github.com/hassansin/gocui"
	"github.com/pkg/errors"
)

type Node interface {
	Title() string
	Children() []Node
}

const (
	pipe   = "│ "
	middle = "├─"
	last   = "└─"
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
	if v, err := m.g.SetView(m.Name(), -1, 0, 30, maxY-2); err != nil {
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
		result += node.Title() + "\n"
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
