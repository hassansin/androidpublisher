package ui

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/hassansin/gocui"
	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
)

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

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

//NewForm returns a new Form
func NewForm(g *gocui.Gui, title string, x0, y0 int) (*Form, error) {
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

	if v, err := f.g.SetView(fmt.Sprintf("%v", rand.Int()), f.x0+2, f.y1, f.x0+2+input.Cols, f.y1+input.Rows+1); err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Wrap(err, "unable to create input view")
		}
		v.Title = input.Name
		v.Wrap = true
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
		f.y1 = f.y1 + 4
		if input.Cols >= f.x1-f.x0 {
			f.x1 = f.x0 + input.Cols + 4
		}

		input.view = v
		f.inputs = append(f.inputs, input)
		if _, err := f.g.SetView(f.formView.Name(), f.x0, f.y0, f.x1, f.y1); err != nil {
			return err
		}
		f.setFooter()
	}
	return nil
}

func (f *Form) close() error {
	for _, input := range f.inputs {
		if err := deleteView(f.g, input.view.Name()); err != nil {
			return err
		}
	}
	return deleteView(f.g, f.formView.Name())
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

func deleteView(g *gocui.Gui, name string) error {
	g.DeleteKeybindings(name)
	if err := g.DeleteView(name); err != nil && err != gocui.ErrUnknownView {
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

func (f *Form) setFooter() {
	v := f.formView
	v.Clear()
	x, y := v.Size()
	var msg string
	if len(f.inputs) > 1 {
		msg = msg + fmt.Sprintf("%v:Switch Input ", aurora.Cyan("TAB"))
	}
	msg = msg + fmt.Sprintf("%v:Submit %v:Cancel", aurora.Cyan("ENTER"), aurora.Cyan("ESC"))

	for i := 0; i < y; i++ {
		fmt.Fprintln(v)
	}
	if x < len(strip(msg)) {
		fmt.Fprintln(v, msg)
	}
	for i := 0; i < (x-len(strip(msg)))/2; i++ {
		fmt.Fprint(v, " ")
	}
	fmt.Fprint(v, msg)
}

func strip(str string) string {
	return re.ReplaceAllString(str, "")
}
