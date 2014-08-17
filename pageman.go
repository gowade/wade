package wade

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
	urlrouter "github.com/naoina/kocha-urlrouter"
	_ "github.com/naoina/kocha-urlrouter/regexp"
	"github.com/phaikawl/wade/bind"
)

const (
	WadeReservedPrefix = "wade-rsvd-"
	WadeExcludeAttr    = WadeReservedPrefix + "exclude"

	GlobalDisplayScope = "__global__"
)

type (
	History interface {
		ReplaceState(title string, path string)
		PushState(title string, path string)
		CurrentPath() string
	}

	pageManager struct {
		routes       []urlrouter.Record
		router       urlrouter.URLRouter
		currentPage  *page
		startPageId  string
		basePath     string
		notFoundPage *page
		container    jq.JQuery
		tcontainer   jq.JQuery

		binding        *bind.Binding
		tm             *custagMan
		pc             *PageCtrl
		displayScopes  map[string]displayScope
		globalDs       *globalDisplayScope
		formattedTitle string
		history        History
	}
)

func newPageManager(history History, config AppConfig,
	tcontainer jq.JQuery, binding *bind.Binding, tm *custagMan) *pageManager {

	container := gJQ("<div class='wade-wrapper'></div>")
	container.AppendTo(gJQ("body"))
	pm := &pageManager{
		routes:        make([]urlrouter.Record, 0),
		router:        urlrouter.NewURLRouter("regexp"),
		currentPage:   nil,
		basePath:      config.BasePath,
		startPageId:   config.StartPage,
		notFoundPage:  nil,
		container:     container,
		tcontainer:    tcontainer,
		binding:       binding,
		tm:            tm,
		displayScopes: make(map[string]displayScope),
		globalDs:      &globalDisplayScope{},
		history:       history,
	}

	pm.displayScopes[GlobalDisplayScope] = pm.globalDs
	return pm
}

// Set the target element that receives Wade's real HTML output,
// by default the container is <body>
func (pm *pageManager) SetOutputContainer(elementId string) {
	parent := gJQ("#" + elementId)
	if parent.Length == 0 {
		panic(fmt.Sprintf("No such element #%v.", elementId))
	}

	parent.Append(pm.container)
}

func (pm *pageManager) addRoute(p *page) {
	pm.routes = append(pm.routes, urlrouter.NewRecord(p.path, p))
}

func (pm *pageManager) CurrentPageId() string {
	return pm.currentPage.id
}

func (pm *pageManager) cutPath(path string) string {
	if strings.HasPrefix(path, pm.basePath) {
		path = path[len(pm.basePath):]
	}
	return path
}

func (pm *pageManager) page(id string) *page {
	if ds, hasDs := pm.displayScopes[id]; hasDs {
		if page, ok := ds.(*page); ok {
			return page
		}
	}

	panic(fmt.Sprintf(`No such page "%v" found.`, id))
}

func (pm *pageManager) displayScope(id string) displayScope {
	if ds, ok := pm.displayScopes[id]; ok {
		return ds
	}
	panic(fmt.Sprintf(`No such page or page group "%v" found.`, id))
}

func (pm *pageManager) SetNotFoundPage(pageId string) {
	pm.notFoundPage = pm.page(pageId)
}

// Url returns the full path
func (pm *pageManager) Fullpath(path string) string {
	return pm.basePath + path
}

func (pm *pageManager) setupPageOnLoad() {
	path := pm.cutPath(pm.history.CurrentPath())
	if path == "/" {
		startPage := pm.page(pm.startPageId)
		path = startPage.path
		pm.history.ReplaceState(startPage.title, pm.Fullpath(path))
	}
	pm.updatePage(path, false)
}

func (pm *pageManager) RedirectToPage(page string, params ...interface{}) {
	url, err := pm.PageUrl(page, params...)
	if err != nil {
		panic(err.Error())
	}
	pm.updatePage(url, true)
}

func (pm *pageManager) RedirectToUrl(url string) {
	js.Global.Get("window").Set("location", url)
}

func (pm *pageManager) prepare() {
	//build the router
	pm.router.Build(pm.routes)

	// preprocess wsection elements
	pm.tcontainer.Find("wsection").Each(func(_ int, e jq.JQuery) {
		name := strings.TrimSpace(e.Attr("name"))
		if name == "" {
			panic(`Error: a <wsection> doesn't have or have empty name`)
		}
		for _, c := range name {
			if !unicode.IsDigit(c) && !unicode.IsLetter(c) && c != '-' {
				panic(fmt.Sprintf("Invalid character '%q' in wsection name.", c))
			}
		}
		e.SetAttr("id", WadeReservedPrefix+name)
	})

	if pm.container.Length == 0 {
		panic(fmt.Sprintf("Cannot find the page container #%v.", pm.container))
	}

	gJQ(js.Global.Get("window")).On("popstate", func() {
		pm.updatePage(pm.history.CurrentPath(), false)
	})

	//Handle link events
	pm.container.On(jq.CLICK, "a", func(e jq.Event) {
		a := gJQ(e.Target)

		pagepath := a.Attr(bind.WadePageAttr)
		if pagepath == "" { //not a wade page link, let the browser do its job
			return
		}

		e.PreventDefault()

		pm.updatePage(pagepath, true)
	})

	pm.setupPageOnLoad()
}

func walk(elem jq.JQuery, pm *pageManager) {
	elem.Children("").Each(func(_ int, e jq.JQuery) {
		belong := e.Attr("w-belong")
		if belong == "" {
			walk(e, pm)
		} else {
			if ds, ok := pm.displayScopes[belong]; ok {
				if ds.hasPage(pm.currentPage.id) {
					walk(e, pm)
				} else {
					e.Remove()
				}
			} else {
				panic(fmt.Sprintf(`Invalid value "%v" for w-belong, no such page or page group is registered.`, belong))
			}
		}
	})
}

func (pm *pageManager) updatePage(url string, pushState bool) {
	url = pm.cutPath(url)
	println("path: " + url)
	match, routeparams := pm.router.Lookup(url)
	if match == nil {
		if pm.notFoundPage != nil {
			pm.updatePage(pm.notFoundPage.path, false)
		} else {
			panic("Page not found. No 404 handler declared.")
		}
		return
	}

	page := match.(*page)

	if pushState {
		pm.history.PushState(page.title, pm.Fullpath(url))
	}

	params := make(map[string]interface{})
	for _, param := range routeparams {
		params[param.Name] = param.Value
	}

	if pm.currentPage != page {
		pm.formattedTitle = page.title
		pm.container.Hide()
		pm.currentPage = page
		pcontents := pm.tcontainer.Clone()
		walk(pcontents, pm)
		pm.container.SetHtml(pcontents.Html())

		pm.container.Find("wrep").Each(func(_ int, e jq.JQuery) {
			e.Remove()
			pm.container.Find("#" + WadeReservedPrefix + e.Attr("target")).
				SetHtml(e.Html())
		})

		pm.container.Find("wsection").Each(func(_ int, e jq.JQuery) {
			e.Children("").First().Unwrap()
		})

		go func() {
			pm.bind(params)
			wrapperElemsUnwrap(pm.container)
			pm.container.Show()
		}()

		tElem := gJQ("<title>").SetHtml(pm.formattedTitle)
		oElem := gJQ("head").Find("title")
		if oElem.Length == 0 {
			gJQ("head").Append(tElem)
		} else {
			oElem.ReplaceWith(tElem)
		}
	}
}

// PageUrl returns the url for the page with the given parameters
func (pm *pageManager) PageUrl(pageId string, params ...interface{}) (u string, err error) {
	err = nil
	page := pm.page(pageId)

	n := len(params)
	if n == 0 {
		u = page.path
		return
	}

	k, i := 0, 0
	route := page.path
	routeparams := urlrouter.ParamNames(route)
	for {
		if i >= len(route) {
			break
		}

		if urlrouter.IsMetaChar(route[i]) && route[i:i+len(routeparams[k])] == routeparams[k] {
			if k < len(params) && params[k] != nil {
				u += fmt.Sprintf("%v", params[k])
			}
			i += len(routeparams[k])
			k++
		} else {
			u += string(route[i])
			i++
		}
	}

	if k != len(params) || k != len(routeparams) {
		err = fmt.Errorf(`Wrong number of parameters for the route of %v. Expected %v, got %v.`,
			pageId, len(params), len(routeparams))
		return
	}

	return
}

func (pm *pageManager) BasePath() string {
	return pm.basePath
}

func (pm *pageManager) newPageCtrl(page *page, params map[string]interface{}) *PageCtrl {
	return &PageCtrl{
		pm:      pm,
		p:       page,
		params:  params,
		helpers: make(map[string]interface{}),
	}
}

func (pm *pageManager) CurrentPage() ThisPage {
	return pm.pc
}

func (pm *pageManager) bind(params map[string]interface{}) {
	models := make([]interface{}, 0)
	controllers := make([]PageControllerFunc, 0)

	pc := pm.newPageCtrl(pm.currentPage, params)

	add := func(ds displayScope) {
		if ctrls := ds.Controllers(); ctrls != nil {
			for _, controller := range ctrls {
				controllers = append(controllers, controller)
			}
		}
	}

	add(pm.globalDs)
	for _, grp := range pm.currentPage.groups {
		add(grp)
	}
	add(pm.currentPage)

	if len(controllers) > 0 {
		completeChan := make(chan bool, 1)
		queueChan := make(chan bool, len(controllers))
		for _, controller := range controllers {
			go func(controller PageControllerFunc) {
				//gopherjs:blocking
				models = append(models, controller(pc))
				queueChan <- true
				if len(queueChan) == len(controllers) {
					completeChan <- true
				}
			}(controller)
		}
		<-completeChan
	}

	if len(models) == 0 {
		pm.binding.Bind(pm.container, nil, true, false)
	} else {
		pm.binding.BindModels(pm.container, models, false, false)
	}

	pm.pc = pc
}

// RegisterController adds a new controller function for the specified
// page / page group.
func (pm *pageManager) registerController(displayScope string, fn PageControllerFunc) {
	ds := pm.displayScope(displayScope)
	ds.addController(fn)
}

// RegisterDisplayScopes registers the given map of pages and page groups
func (pm *pageManager) registerDisplayScopes(m map[string]DisplayScope) {
	for id, item := range m {
		if id == "" {
			panic("id of page/page group cannot be empty.")
		}

		pm.displayScopes[id] = item.Register(id, pm)
	}
}
