package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

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
	defaultStatus = fmt.Sprintf("%v:Switch Panel %v:Request %v:Save Response %v:Quit %v:Navigate %v:Refresh", aurora.Cyan("TAB"), aurora.Cyan("ENTER"), aurora.Cyan("CTRL+S"), aurora.Cyan("CTRL+C"), aurora.Cyan("↑↓"), aurora.Cyan("F5"))
	activeGr      *Group
	activeOp      *Operation
)

func nextView(g *gocui.Gui, v *gocui.View) error {
	status.Reset()
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
	f, err := ui.NewForm(g, "Parameters", maxX/2-30, maxY/2-int(float64(len(op.Params))*4/2)-1)
	if err != nil {
		return err
	}
	f.OnCancel(func() error {
		return sideView.SetCurrent()
	}).OnError(func(err error) {
		status.UpdateError(err.Error())
	}).OnSubmit(func() error {
		go makeRequest(g, grp, op)
		return sideView.SetCurrent()
	})
	for i, param := range op.Params {
		focused := false
		if i == 0 {
			focused = true
		}
		input := ui.NewInput(param.Name, &param.Value, 60, focused)
		if param.Multiline {
			input.Rows = 6
		}
		input.Required = param.Required
		if err := f.Input(input); err != nil {
			return err
		}
	}
	return nil
}

func reRequest(g *gocui.Gui, v *gocui.View) error {
	if activeGr != nil && activeOp != nil {
		go makeRequest(g, activeGr, activeOp)
	}
	return nil
}

func makeRequest(g *gocui.Gui, grp *Group, op *Operation) {
	status.Update("Loading...")
	activeGr = grp
	activeOp = op
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
	if err := mainView.SetKeybinding(); err != nil {
		return err
	}
	if err := sideView.SetKeybinding(); err != nil {
		return err
	}
	if err := g.SetKeybinding(sideView.Name(), gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding(mainView.Name(), gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding(sideView.Name(), gocui.KeyEnter, gocui.ModNone, processOp); err != nil {
		return err
	}
	if err := g.SetKeybinding(sideView.Name(), gocui.KeyF5, gocui.ModNone, reRequest); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, func(_ *gocui.Gui, _ *gocui.View) error {
		status.Reset()
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, func(_ *gocui.Gui, _ *gocui.View) error {
		status.Reset()
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
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

	mainView.OnSave(func(filename string, err error) {
		if err != nil {
			status.UpdateError(fmt.Sprintf("Unable to save response: %v", err.Error()))
			return
		}
		status.UpdateSuccess(fmt.Sprintf("File saved to %v", filename))
	})
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
		Name:   "List",
		Params: []*Param{{Name: "PageToken"}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewInappproductsService(service)
			call := s.List(pkgName)
			if params[0].Value != "" {
				call.Token(params[0].Value)
			}
			return call.Do()
		},
	})
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "Delete",
		Params: []*Param{{Name: "SKU", Required: true}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewInappproductsService(service)
			call := s.Delete(pkgName, params[0].Value)
			return nil, call.Do()
		},
	})
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "Get",
		Params: []*Param{{Name: "SKU", Required: true}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewInappproductsService(service)
			call := s.Get(pkgName, params[0].Value)
			return call.Do()
		},
	})
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "Patch",
		Params: []*Param{{Name: "SKU", Required: true}, {Name: "Body", Multiline: true}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewInappproductsService(service)
			call := s.Get(pkgName, params[0].Value)
			return call.Do()
		},
	})
	grp = &Group{Name: "Orders"}
	groups = append(groups, grp)
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "Refund",
		Params: []*Param{{Name: "OrderID", Required: true}, {Name: "Revoke (true/false)"}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewOrdersService(service)
			call := s.Refund(pkgName, params[0].Value)
			if strings.ToLower(params[1].Value) == "true" {
				call.Revoke(true)
			}
			return nil, call.Do()
		},
	})

	grp = &Group{Name: "Purchases.products"}
	groups = append(groups, grp)
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "Get",
		Params: []*Param{{Name: "ProductID", Required: true}, {Name: "Token", Required: true}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewPurchasesProductsService(service)
			call := s.Get(pkgName, params[0].Value, params[1].Value)
			return call.Do()
		},
	})
	grp = &Group{Name: "Purchases.subscriptions"}
	groups = append(groups, grp)
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "Cancel",
		Params: []*Param{{Name: "SubscriptionId", Required: true}, {Name: "Token", Required: true}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewPurchasesSubscriptionsService(service)
			call := s.Cancel(pkgName, params[0].Value, params[1].Value)
			return nil, call.Do()
		},
	})
	grp.Operations = append(grp.Operations, &Operation{
		Name: "Defer",
		Params: []*Param{
			{Name: "SubscriptionId", Required: true}, {Name: "Token", Required: true},
			{Name: "DesiredExpiryTimeMillis", Required: true}, {Name: "ExpectedExpiryTimeMillis", Required: true},
		},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewPurchasesSubscriptionsService(service)
			desired, err := strconv.ParseInt(params[2].Value, 10, 64)
			if err != nil {
				return nil, err
			}
			expected, err := strconv.ParseInt(params[3].Value, 10, 64)
			if err != nil {
				return nil, err
			}
			call := s.Defer(pkgName, params[0].Value, params[1].Value, &androidpublisher.SubscriptionPurchasesDeferRequest{
				DeferralInfo: &androidpublisher.SubscriptionDeferralInfo{
					DesiredExpiryTimeMillis:  desired,
					ExpectedExpiryTimeMillis: expected,
				},
			})
			return call.Do()
		},
	})
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "Get",
		Params: []*Param{{Name: "SubscriptionId", Required: true}, {Name: "Token", Required: true}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewPurchasesSubscriptionsService(service)
			call := s.Get(pkgName, params[0].Value, params[1].Value)
			return call.Do()
		},
	})
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "Refund",
		Params: []*Param{{Name: "SubscriptionId", Required: true}, {Name: "Token", Required: true}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewPurchasesSubscriptionsService(service)
			call := s.Refund(pkgName, params[0].Value, params[1].Value)
			return nil, call.Do()
		},
	})
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "Revoke",
		Params: []*Param{{Name: "SubscriptionId", Required: true}, {Name: "Token", Required: true}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewPurchasesSubscriptionsService(service)
			call := s.Revoke(pkgName, params[0].Value, params[1].Value)
			return nil, call.Do()
		},
	})

	grp = &Group{Name: "Purchases.voidedpurchases"}
	groups = append(groups, grp)
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "List",
		Params: []*Param{{Name: "StartTime(milliseconds)"}, {Name: "EndTime(milliseconds)"}, {Name: "MaxResults"}, {Name: "PageToken"}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewPurchasesVoidedpurchasesService(service)
			call := s.List(pkgName)
			if params[0].Value != "" {
				ms, _ := strconv.ParseInt(params[0].Value, 10, 64)
				call.StartTime(ms)
			}
			if params[1].Value != "" {
				ms, _ := strconv.ParseInt(params[1].Value, 10, 64)
				call.EndTime(ms)
			}
			if params[2].Value != "" {
				v, _ := strconv.ParseInt(params[2].Value, 10, 64)
				call.MaxResults(v)
			}
			if params[3].Value != "" {
				call.Token(params[3].Value)
			}
			return call.Do()
		},
	})

	grp = &Group{Name: "Reviews"}
	groups = append(groups, grp)
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "List",
		Params: []*Param{{Name: "MaxResults"}, {Name: "PageToken"}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewReviewsService(service)
			call := s.List(pkgName)
			if params[0].Value != "" {
				v, _ := strconv.ParseInt(params[0].Value, 10, 64)
				call.MaxResults(v)
			}
			if params[1].Value != "" {
				call.Token(params[1].Value)
			}
			return call.Do()
		},
	})
	grp.Operations = append(grp.Operations, &Operation{
		Name:   "Get",
		Params: []*Param{{Name: "ReviewID", Required: true}},
		Do: func(params []*Param) (interface{}, error) {
			s := androidpublisher.NewReviewsService(service)
			call := s.Get(pkgName, params[0].Value)
			return call.Do()
		},
	})
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
	g.Highlight = true
	g.FgColor = gocui.ColorDefault
	g.SelFgColor = gocui.ColorWhite | gocui.AttrBold

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
