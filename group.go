package main

import "github.com/hassansin/androidpublisher/ui"

type Groups []*Group

func (g Groups) ToNodes() []ui.Node {
	nodes := make([]ui.Node, len(g))
	for i, op := range g {
		nodes[i] = op
	}
	return nodes
}

type Group struct {
	Name       string
	Operations []*Operation
}

func (g Group) Title() string {
	return g.Name
}

func (g Group) Children() []ui.Node {
	nodes := make([]ui.Node, len(g.Operations))
	for i, op := range g.Operations {
		nodes[i] = op
	}
	return nodes
}

type Operation struct {
	Name   string
	Params []*Param
	Do     func([]*Param) (interface{}, error)
}

func (op Operation) Title() string {
	return op.Name
}

func (op Operation) Children() []ui.Node {
	return nil
}

type Param struct {
	Name, Value         string
	Required, Multiline bool
}
