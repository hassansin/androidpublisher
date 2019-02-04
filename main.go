package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hassansin/androidpublisher/ui"

	"github.com/hassansin/gocui"
	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/androidpublisher/v3"
)

var (
	status        *ui.StatusLine
	mainView      *ui.MainView
	sideView      *ui.TreeView
	groups        Groups
	defaultStatus = fmt.Sprintf("%v:Switch Panel %v:Request %v:Save Response %v:Quit %v:Navigate", aurora.Cyan("TAB"), aurora.Cyan("ENTER"), aurora.Cyan("CTRL+S"), aurora.Cyan("CTRL+C"), aurora.Cyan("↑↓"))
)

func nextView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == sideView.Name() {
		return mainView.SetCurrent()
	}
	return sideView.SetCurrent()
}

func processOp(g *gocui.Gui, v *gocui.View) error {
	status.Reset()
	idx := sideView.Selected()
	if len(idx) < 2 {
		return nil
	}
	grp := groups[idx[0]]
	op := grp.Operations[idx[1]]
	maxX, maxY := g.Size()
	if len(op.Params) == 0 {
		go makeRequest(g, grp, op)
		return nil
	}
	f, err := ui.NewForm(g, "Parameters", maxX/2-30, maxY/3-(len(op.Params)-1)/2)
	if err != nil {
		return err
	}
	f.OnCancel(func() error {
		return sideView.SetCurrent()
	})
	f.OnSubmit(func(values map[string]string) error {
		for _, param := range op.Params {
			param.Value = values[param.Name]
		}
		go makeRequest(g, grp, op)
		return sideView.SetCurrent()
	})

	for i, param := range op.Params {
		focused := false
		if i == 0 {
			focused = true
			//g.SetCurrentView(v.Name())
		}
		if err := f.Input(ui.NewInput(param.Name, 60, focused)); err != nil {
			return err
		}
	}
	return nil
}

func makeRequest(g *gocui.Gui, grp *Group, op *Operation) {
	status.Update("Loading...")
	result, err := op.Do(op.Params)
	g.Update(func(g *gocui.Gui) error {
		if err != nil {
			status.UpdateError("Request failed")
			result = err
		} else {
			status.UpdateSuccess("Request successful")
		}
		mainView.LoadContent(op.Name+" "+grp.Name, result)
		return nil
	})
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding(sideView.Name(), gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding(mainView.Name(), gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding(sideView.Name(), gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(sideView.Name(), 'j', gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(mainView.Name(), gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(sideView.Name(), gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding(sideView.Name(), 'k', gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding(mainView.Name(), gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding(mainView.Name(), gocui.KeyArrowLeft, gocui.ModNone, cursorLeft); err != nil {
		return err
	}
	if err := g.SetKeybinding(mainView.Name(), gocui.KeyArrowRight, gocui.ModNone, cursorRight); err != nil {
		return err
	}
	if err := g.SetKeybinding(mainView.Name(), gocui.KeyPgdn, gocui.ModNone, pageDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(mainView.Name(), gocui.KeyPgup, gocui.ModNone, pageUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding(sideView.Name(), gocui.KeyEnter, gocui.ModNone, processOp); err != nil {
		return err
	}
	if err := g.SetKeybinding(sideView.Name(), gocui.KeyCtrlS, gocui.ModNone, saveDialog); err != nil {
		return err
	}
	if err := g.SetKeybinding(mainView.Name(), gocui.KeyCtrlS, gocui.ModNone, saveDialog); err != nil {
		return err
	}
	return nil
}

func init() {
	pflag.String("package", "", "android package name")
	pflag.String("credentials", "credentials.json", "path to google service account JSON credentials file")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

}

func createLayout(g *gocui.Gui) func(*gocui.Gui) error {
	status = ui.NewStatusLine(g)
	mainView = ui.NewMainView(g)
	sideView = ui.NewTreeView(g)
	return func(g *gocui.Gui) error {
		if err := sideView.SetView("Operations", groups.ToNodes()); err != nil {
			return err
		}
		if err := mainView.SetView(); err != nil {
			return err
		}

		if err := status.SetView(defaultStatus); err != nil {
			return err
		}
		return nil
	}
}

func initOperations(service *androidpublisher.Service, pkgName string) {
	grp := &Group{Name: "Inappproducts"}
	groups = append(groups, grp)
	grp.Operations = append(grp.Operations, &Operation{
		Name: "List",
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewInappproductsService(service)
			call := s.List(pkgName)
			return call.Do()
		},
	})
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "Get",
		Params: []*Param{{Name: "SKU"}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewInappproductsService(service)
			call := s.Get(pkgName, params[0].Value)
			return call.Do()
		},
	})
	grp = &Group{Name: "Orders"}
	groups = append(groups, grp)

	grp = &Group{Name: "Purchases.subscriptions"}
	groups = append(groups, grp)
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "Get",
		Params: []*Param{{Name: "SubscriptionId"}, {Name: "Token"}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewPurchasesSubscriptionsService(service)
			call := s.Get(pkgName, params[0].Value, params[1].Value)
			return call.Do()
		},
	})
}

func saveDialog(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()
	currentView := g.CurrentView()
	f, err := ui.NewForm(g, "Save Response", maxX/2-20, maxY/2)
	if err != nil {
		return err
	}
	f.OnCancel(func() error {
		_, err := g.SetCurrentView(currentView.Name())
		return err
	})
	f.OnSubmit(func(values map[string]string) error {
		filename, _ := values["File Name"]
		if filename == "" {
			return nil
		}
		if err := mainView.SaveContent(filename); err != nil {
			status.UpdateError(fmt.Sprintf("Unable to save response: %v", err.Error()))
			return nil
		}
		fullPath, err := filepath.Abs(filename)
		if err != nil {
			return err
		}
		status.UpdateSuccess(fmt.Sprintf("File saved to %v", fullPath))
		_, err = g.SetCurrentView(currentView.Name())
		return err
	})
	return f.Input(ui.NewInput("File Name", 40, true))
}

func do() error {
	pkgName := viper.GetString("package")
	if pkgName == "" {
		return errors.New("missing android package name")
	}

	data, err := ioutil.ReadFile(viper.GetString("credentials"))
	if err != nil {
		return errors.Wrapf(err, "unable to read credentials (%v)", viper.GetString("credentials"))
	}

	conf, err := google.JWTConfigFromJSON(data, androidpublisher.AndroidpublisherScope)
	if err != nil {
		return err
	}
	client := conf.Client(oauth2.NoContext)
	service, err := androidpublisher.New(client)
	if err != nil {
		return err
	}
	initOperations(service, pkgName)

	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		return err
	}
	defer g.Close()
	g.InputEsc = true
	g.Cursor = true

	g.SetManagerFunc(createLayout(g))

	if err := keybindings(g); err != nil {
		return err
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

func main() {
	if err := do(); err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}
