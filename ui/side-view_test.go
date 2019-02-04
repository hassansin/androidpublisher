package ui

import (
	"fmt"
	"testing"
)

type Item struct {
	name     string
	children []Item
}

func (n Item) Title() string {
	return n.name
}

func (n Item) Children() []Node {
	nodes := make([]Node, len(n.children))
	for i, item := range n.children {
		nodes[i] = item
	}
	return nodes
}

func TestTree(t *testing.T) {
	nodes := []Node{
		Item{
			name: "item1",
			children: []Item{
				Item{
					name: "item1-child-1",
					children: []Item{
						Item{
							name: "item1-child-1-child-1",
						},
						Item{
							name: "item1-child-1-child-2",
						},
					},
				},
				Item{
					name: "item1-child-2",
				},
			},
		},
		Item{
			name: "item2",
			children: []Item{
				Item{
					name: "item1-child-1",
					children: []Item{
						Item{
							name: "item1-child-1-child-1",
						},
						Item{
							name: "item1-child-1-child-2",
						},
					},
				},
			},
		},
		Item{
			name: "item3",
			children: []Item{
				Item{
					name: "item1-child-1",
					children: []Item{
						Item{
							name: "item1-child-1-child-1",
						},
						Item{
							name: "item1-child-1-child-2",
						},
					},
				},
			},
		},
	}

	fmt.Println(Tree(nodes, ""))
	idx := 11
	fmt.Println(Selected(nodes, &idx))

}
