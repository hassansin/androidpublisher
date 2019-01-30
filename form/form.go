package form

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/hassansin/gocui"
	"github.com/pkg/errors"
)

/*
form := NewForm(title, x, y)
form.Input(name, default, required, mask, size, focus)
form.OnSubmit(func(){
})
form.OnCancel(func(){
})
*/

//Form represents an input form
type Form struct {
	g              *gocui.Gui
	title          string
	x0, y0, x1, y1 int
	formView       *gocui.View
	inputs         []*Input
	onCancel       func() error
	onSubmit       func(map[string]string) error
}

//New returns a new Form
func New(g *gocui.Gui, title string, x0, y0 int) (*Form, error) {
	x1 := x0 + len(title) + 2
	y1 := y0 + 1
	v, err := g.SetView(fmt.Sprintf("%v", rand.Int()), x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return nil, errors.Wrap(err, "unable to create form view")
	}
	v.Frame = true
	v.Title = title
	return &Form{
		g:        g,
		title:    title,
		x0:       x0,
		y0:       y0,
		x1:       x1,
		y1:       y1,
		formView: v,
	}, nil
}

//NewInput return new Input
func NewInput(name string, size int, focused bool) *Input {
	return &Input{
		Name:    name,
		Cols:    size,
		Rows:    1,
		focused: focused,
	}
}

//Input represents one line input
type Input struct {
	Rows, Cols        int
	Required, focused bool
	Mask              rune
	Name, value       string
	view              *gocui.View
}

//Input adds a new input line to the form
func (f *Form) Input(input *Input) error {

	if v, err := f.g.SetView(fmt.Sprintf("%v", rand.Int()), f.x0+1, f.y1, f.x0+1+input.Cols, f.y1+input.Rows+1); err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Wrap(err, "unable to create input view")
		}
		v.Title = input.Name
		v.Wrap = false
		v.Editable = true
		v.Mask = input.Mask
		if err := f.g.SetKeybinding(v.Name(), gocui.KeyEsc, gocui.ModNone, f.cancel); err != nil {
			return err
		}
		if err := f.g.SetKeybinding(v.Name(), gocui.KeyTab, gocui.ModNone, f.next); err != nil {
			return err
		}
		if err := f.g.SetKeybinding(v.Name(), gocui.KeyEnter, gocui.ModNone, f.submit); err != nil {
			return err
		}
		if input.focused {
			if _, err = f.g.SetCurrentView(v.Name()); err != nil {
				return err
			}
		}
		f.y1 = f.y1 + 3
		if input.Cols >= f.x1-f.x0 {
			f.x1 = f.x0 + input.Cols + 2
		}

		input.view = v
		f.inputs = append(f.inputs, input)
		if _, err := f.g.SetView(f.formView.Name(), f.x0, f.y0, f.x1, f.y1); err != nil {
			return err
		}
	}
	return nil
}

func (f *Form) close() error {
	for _, input := range f.inputs {
		if err := f.deleteView(input.view.Name()); err != nil {
			return err
		}
	}
	return f.deleteView(f.formView.Name())
}
func (f *Form) cancel(g *gocui.Gui, v *gocui.View) error {
	if err := f.close(); err != nil {
		return errors.Wrap(err, "error closing form")
	}
	if f.onCancel != nil {
		return f.onCancel()
	}
	return nil
}

//next cycles through the input fields
func (f *Form) next(g *gocui.Gui, v *gocui.View) error {
	for i, input := range f.inputs {
		if input.focused {
			next := i + 1
			if i >= len(f.inputs)-1 {
				next = 0
			}
			if _, err := f.g.SetCurrentView(f.inputs[next].view.Name()); err != nil {
				return err
			}
			f.inputs[next].focused = true
			input.focused = false
			break
		}
	}
	return nil
}

func (f *Form) submit(g *gocui.Gui, v *gocui.View) error {
	values := make(map[string]string)
	for _, input := range f.inputs {
		values[input.Name] = strings.TrimSpace(input.view.Buffer())
	}
	if f.onSubmit != nil {
		if err := f.onSubmit(values); err != nil {
			return err
		}
	}
	return f.close()
}

func (f *Form) deleteView(name string) error {
	f.g.DeleteKeybindings(name)
	if err := f.g.DeleteView(name); err != nil && err != gocui.ErrUnknownView {
		return err
	}
	return nil
}

//OnCancel binds function to be called when form is cancelled
func (f *Form) OnCancel(fn func() error) *Form {
	f.onCancel = fn
	return f
}

//OnSubmit binds function to be called when form is submitted
func (f *Form) OnSubmit(fn func(map[string]string) error) *Form {
	f.onSubmit = fn
	return f
}

/*

func showDialog(g *gocui.Gui, y int, name string) (*gocui.View, error) {
	maxX, _ := g.Size()
	v, err := g.SetView("input-"+name, maxX/2-30, y-1, maxX/2+30, y+1)
	if err != nil && err != gocui.ErrUnknownView {
		return nil, err
	}
	if err := g.SetKeybinding("input-"+name, gocui.KeyEnter, gocui.ModNone, inputSubmit); err != nil {
		return nil, err
	}
	if err := g.SetKeybinding("input-"+name, gocui.KeyTab, gocui.ModNone, inputTab); err != nil {
		return nil, err
	}
	if err := g.SetKeybinding("input-"+name, gocui.KeyEsc, gocui.ModNone, inputCancel); err != nil {
		return nil, err
	}
	v.Title = name
	v.Editable = true
	return v, nil
}
*/
