// Copyright 2014 The gocui Authors. All rights reserved.

// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"androidpublisher-cli/form"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/hassansin/gocui"
	"github.com/nwidger/jsoncolor"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/androidpublisher/v3"
)

var (
	service *androidpublisher.Service
	groups  Groups
	pkgName string
	idx     int
)

type App struct {
	service *androidpublisher.Service
	groups  Groups
	pkgName string
	idx     int
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == "side" {
		_, err := g.SetCurrentView("main")
		return err
	}
	_, err := g.SetCurrentView("side")
	return err
}

func processOp(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	grp, op := groups.FromIdx(cy)
	if grp == nil && op == nil {
		return nil
	}
	if grp != nil && op == nil {
		//@TODO
		//expand/collapse group
		return nil
	}
	maxX, maxY := g.Size()
	if len(op.Params) == 0 {
		go makeRequest(g, grp, op)
		return nil
	}
	f, err := form.New(g, "Parameters", maxX/2-30, maxY/2-(len(op.Params)-1)/2)
	if err != nil {
		return err
	}
	f.OnCancel(func() error {
		_, err := g.SetCurrentView("side")
		return err
	})
	f.OnSubmit(func(values map[string]string) error {
		for _, param := range op.Params {
			param.Value = values[param.Name]
		}
		go makeRequest(g, grp, op)
		return nil
	})

	for i, param := range op.Params {
		focused := false
		if i == 0 {
			focused = true
			//g.SetCurrentView(v.Name())
		}
		if err := f.Input(form.NewInput(param.Name, 60, focused)); err != nil {
			return err
		}
	}
	return nil
}

func makeRequest(g *gocui.Gui, grp *Group, op *Operation) {
	g.Update(func(g *gocui.Gui) error {
		_, err := showInfo(g, "Loading...")
		return errors.Wrap(err, "faliled to show info")
	})
	result := op.Do(op.Params)
	g.Update(func(g *gocui.Gui) error {
		if err := deleteView(g, "infobox"); err != nil {
			return errors.Wrap(err, "failed to close info")
		}
		v, err := g.SetCurrentView("main")
		if err != nil {
			return errors.Wrap(err, "failed to current view to main")
		}
		v.Title = fmt.Sprintf("Response(%v %v)", op.Name, grp.Name)
		v.Clear()
		v.SetCursor(0, 0)
		v.SetOrigin(0, 0)
		fmt.Fprintf(v, result)
		g.SetCurrentView("side")
		return nil
	})
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("side", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyArrowLeft, gocui.ModNone, cursorLeft); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyArrowRight, gocui.ModNone, cursorRight); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyPgdn, gocui.ModNone, cursorEnd); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyEnter, gocui.ModNone, processOp); err != nil {
		return err
	}

	return nil
}

func init() {
	pflag.String("package", "", "android package name")
	pflag.String("credentials", "credentials.json", "path to google service account JSON credentials file")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	grp := &Group{Name: "Inappproducts"}
	groups = append(groups, grp)
	grp.Operations = append(grp.Operations, &Operation{
		Name: "List",
		Do: func(params []*Param) string {
			s := androidpublisher.NewInappproductsService(service)
			call := s.List(pkgName)
			res, err := call.Do()
			if err != nil {
				return err.Error()
			}
			body, err := jsoncolor.MarshalIndent(res, "", " ")
			if err != nil {
				return err.Error()
			}
			return string(body)
		},
	})
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "Get",
		Params: []*Param{{Name: "SKU"}},
		Do: func(params []*Param) string {
			s := androidpublisher.NewInappproductsService(service)
			call := s.Get(pkgName, params[0].Value)
			res, err := call.Do()
			if err != nil {
				return err.Error()
			}
			body, err := jsoncolor.MarshalIndent(res, "", " ")
			if err != nil {
				return err.Error()
			}
			return string(body)
		},
	})
	grp = &Group{Name: "Orders"}
	groups = append(groups, grp)

	grp = &Group{Name: "Purchases.subscriptions"}
	groups = append(groups, grp)
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "Get",
		Params: []*Param{{Name: "SubscriptionId"}, {Name: "Token"}},
		Do: func(params []*Param) string {
			s := androidpublisher.NewPurchasesSubscriptionsService(service)
			call := s.Get(pkgName, params[0].Value, params[1].Value)
			res, err := call.Do()
			if err != nil {
				return err.Error()
			}
			body, err := jsoncolor.MarshalIndent(res, "", " ")
			if err != nil {
				return err.Error()
			}
			return string(body)
		},
	})
}

func layout(g *gocui.Gui) error {

	maxX, maxY := g.Size()
	if v, err := g.SetView("side", -1, 0, 30, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = true
		v.Frame = true
		v.Title = "Operations"
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		for i, grp := range groups {
			if i == len(groups)-1 {
				fmt.Fprint(v, "└─")
			} else {
				fmt.Fprint(v, "│─")
			}
			fmt.Fprintln(v, grp.Name)

			for j, op := range grp.Operations {
				if i != len(groups)-1 {
					fmt.Fprint(v, "│ ")
				} else {
					fmt.Fprint(v, " ")
				}
				if j == len(grp.Operations)-1 {
					fmt.Fprint(v, "└─")
				} else {
					fmt.Fprint(v, "│─")
				}
				fmt.Fprintln(v, op.Name)
			}
		}
		if _, err := g.SetCurrentView("side"); err != nil {
			return err
		}
	}
	if v, err := g.SetView("main", 30, 0, maxX, maxY); err != nil {
		v.Frame = true
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Title = "Response"
	}
	return nil
}

func main() {
	if err := do(); err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}
func do() error {
	pkgName = viper.GetString("package")
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
	service, err = androidpublisher.New(client)
	if err != nil {
		return err
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer g.Close()
	g.InputEsc = true
	g.Cursor = true

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		return err
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}
func showInfo(g *gocui.Gui, msg string) (*gocui.View, error) {
	maxX, maxY := g.Size()
	if v, err := g.SetView("infobox", maxX/2-20, maxY/2, maxX/2+20, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return nil, err
		}
		v.Title = "Info"
		fmt.Fprintln(v, msg)
		return v, nil
	}
	return nil, nil
}

func deleteView(g *gocui.Gui, name string) error {
	g.DeleteKeybindings(name)
	if err := g.DeleteView(name); err != nil && err != gocui.ErrUnknownView {
		return err
	}
	return nil
}
