package main

type Groups []*Group

func (g Groups) FromIdx(idx int) (*Group, *Operation) {
	i := 0
	for _, grp := range g {
		if i == idx {
			return grp, nil
		}
		for _, op := range grp.Operations {
			i++
			if idx == i {
				return grp, op
			}
		}
		i++
	}
	return nil, nil
}

type Group struct {
	Name       string
	Operations []*Operation
}

type Operation struct {
	Name   string
	Params []*Param
	Do     func([]*Param) string
}

type Param struct {
	Name  string
	Value interface{}
}
