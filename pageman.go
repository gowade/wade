package wade

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
	"github.com/phaikawl/wade/bind"
)

const (
	WadeReservedPrefix = "wade-rsvd-"
	WadeExcludeAttr    = WadeReservedPrefix + "exclude"
)

var (
	gRouteParamRegexp = regexp.MustCompile(`\:\w+`)
)

type handlable struct {
	controller PageControllerFunc
	handlers   []PageHandler
}

func (h *handlable) addHandler(fn PageHandler) {
	if h.handlers == nil {
		h.handlers = make([]PageHandler, 0)
	}
	h.handlers = append(h.handlers, fn)
}

func (h *handlable) setController(fn PageControllerFunc) {
	h.controller = fn
}

type displayScope interface {
	hasPage(id string) bool
	addHandler(fn PageHandler)
	setController(fn PageControllerFunc)
}

type page struct {
	handlable
	id    string
	path  string
	title string

	groups []*pageGroup
}

func (p *page) addGroup(grp *pageGroup) {
	if p.groups == nil {
		p.groups = make([]*pageGroup, 0)
	}

	p.groups = append(p.groups, grp)
}

func (p *page) hasPage(id string) bool {
	return p.id == id
}

type pageGroup struct {
	handlable
	pages []*page
}

func newPageGroup(pages []*page) *pageGroup {
	return &pageGroup{
		pages: pages,
	}
}

func (pg *pageGroup) hasPage(id string) bool {
	for _, page := range pg.pages {
		if page.id == id {
			return true
		}
	}

	return false
}

func newPage(id, path, title string) *page {
	return &page{
		id:    id,
		path:  path,
		title: title,
	}
}

// PageControllerFunc is the function to be run on the load of a specific page.
// It returns a model to be used in bindings of the elements in the page.
type PageControllerFunc func(*PageCtrl) interface{}

// PageHandler is an additional function to be run on the load of a specific page,
// does not return anything.
type PageHandler func()

// PageManager is Page Manager
type PageManager struct {
	router       js.Object
	currentPage  *page
	startPageId  string
	basePath     string
	notFoundPage *page
	container    jq.JQuery
	tcontainer   jq.JQuery

	binding       *bind.Binding
	tm            *CustagMan
	pc            *PageCtrl
	displayScopes map[string]displayScope
}

// PageView provides access to the page-specific data inside a controller func
type PageCtrl struct {
	params map[string]interface{}

	pm      *PageManager
	helpers []string
}

type PageInfo struct {
	Id    string
	Route string
	Title string
}

// PageInfo returns information about the page
func (pc *PageCtrl) Info() PageInfo {
	pg := pc.pm.currentPage
	return PageInfo{
		Id:    pg.id,
		Route: pg.path,
		Title: pg.title,
	}
}

// SetTitle formats the page's title with the given params
func (pc *PageCtrl) FormatTitle(params ...interface{}) {
	title := fmt.Sprintf(pc.pm.currentPage.title, params...)
	tElem := gJQ("<title>").SetHtml(title)
	oElem := gJQ("head").Find("title")
	if oElem.Length == 0 {
		gJQ("head").Append(tElem)
	} else {
		oElem.ReplaceWith(tElem)
	}
}

// ExportParam sets the value of a parameter to a target.
// The target must be a pointer, typically it would be a pointer to a model's field,
// for example
//	pc.ExportParam("postid", &pmodel.PostId)
func (pc *PageCtrl) ExportParam(param string, target interface{}) {
	v, ok := pc.params[param]
	if !ok {
		panic(fmt.Errorf("Request invalid parameter %v.", param))
	}
	if reflect.TypeOf(target).Kind() != reflect.Ptr {
		panic("The target for saving the parameter value must be a pointer so that it could be modified.")
	}
	_, err := fmt.Sscan(v.(string), target)
	if err != nil {
		panic(err.Error())
	}
	return
}

// RegisterHelper registers fn as a local helper with the given name.
func (pc *PageCtrl) RegisterHelper(name string, fn interface{}) {
	pc.helpers = append(pc.helpers, name)
}

func newPageManager(startPage, basePath string,
	tcontainer jq.JQuery, binding *bind.Binding, tm *CustagMan) *PageManager {

	container := gJQ("<div class='wade-wrapper'></div>")
	container.AppendTo(gJQ("body"))
	return &PageManager{
		router:        js.Global.Get("RouteRecognizer").New(),
		currentPage:   nil,
		basePath:      basePath,
		startPageId:   startPage,
		notFoundPage:  nil,
		container:     container,
		tcontainer:    tcontainer,
		binding:       binding,
		tm:            tm,
		displayScopes: make(map[string]displayScope),
	}
}

func (pm *PageManager) CurrentPageId() string {
	return pm.currentPage.id
}

// Set the target element that receives Wade's real HTML output,
// by default the container is <body>
func (pm *PageManager) SetOutputContainer(elementId string) {
	parent := gJQ("#" + elementId)
	if parent.Length == 0 {
		panic(fmt.Sprintf("No such element #%v.", elementId))
	}

	parent.Append(pm.container)
}

func (pm *PageManager) cutPath(path string) string {
	if strings.HasPrefix(path, pm.basePath) {
		path = path[len(pm.basePath):]
	}
	return path
}

func (pm *PageManager) page(id string) *page {
	if ds, hasDs := pm.displayScopes[id]; hasDs {
		if page, ok := ds.(*page); ok {
			return page
		}
	}

	panic(fmt.Sprintf(`No such page "%v" found.`, id))
}

func (pm *PageManager) displayScope(id string) displayScope {
	if ds, ok := pm.displayScopes[id]; ok {
		return ds
	}
	panic(fmt.Sprintf(`No such page or page group "%v" found.`, id))
}

func (pm *PageManager) SetNotFoundPage(pageId string) {
	pm.notFoundPage = pm.page(pageId)
}

// Url returns the full url for a path
func (pm *PageManager) Url(path string) string {
	return pm.basePath + path
}

func documentUrl() string {
	location := gHistory.Get("location")
	if location.IsNull() || location.IsUndefined() {
		location = js.Global.Get("document").Get("location")
	}
	return location.Get("pathname").Str()
}

func (pm *PageManager) setupPageOnLoad() {
	path := pm.cutPath(documentUrl())
	if path == "/" {
		startPage := pm.page(pm.startPageId)
		path = startPage.path
		gHistory.Call("replaceState", nil, startPage.title, pm.Url(path))
	}
	pm.updatePage(path, false)
}

func (pm *PageManager) prepare() {
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
		pm.updatePage(documentUrl(), false)
	})

	pm.setupPageOnLoad()
}

func walk(elem jq.JQuery, pm *PageManager) {
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

func (pm *PageManager) updatePage(url string, pushState bool) {
	url = pm.cutPath(url)
	matches := pm.router.Call("recognize", url)
	println("path: " + url)
	if matches.IsUndefined() || matches.Length() == 0 {
		if pm.notFoundPage != nil {
			pm.updatePage(pm.notFoundPage.path, false)
		} else {
			panic("Page not found. No 404 handler declared.")
		}
	}

	match := matches.Index(0)
	pageId := match.Get("handler").Invoke().Str()
	page := pm.page(pageId)
	if pushState {
		gHistory.Call("pushState", nil, page.title, pm.Url(url))
	}
	params := make(map[string]interface{})
	prs := match.Get("params")
	if !prs.IsUndefined() {
		params = prs.Interface().(map[string]interface{})
	}

	gJQ("head title").SetText(page.title)
	if pm.currentPage != page {
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
			e.ReplaceWith(e.Html())
		})

		pm.bind(params)

		//Rebind link events
		pm.container.Find("a").On(jq.CLICK, func(e jq.Event) {
			a := gJQ(e.Target)

			pagepath := a.Attr(bind.WadePageAttr)
			if pagepath == "" { //not a wade page link, let the browser do its job
				return
			}

			e.PreventDefault()

			pm.updatePage(pagepath, true)
		})
	}
}

// PageUrl returns the url and route parameters for the specified pageId
func (pm *PageManager) PageUrl(pageId string, params []interface{}) (u string, err error) {
	err = nil
	page := pm.page(pageId)

	n := len(params)
	if n == 0 {
		u = page.path
		return
	}

	i := 0
	repl := func(src string) (out string) {
		out = src
		if i >= n {
			err = fmt.Errorf("Not enough parameters supplied for the route.")
			return
		}
		out = fmt.Sprintf("%v", params[i])
		i += 1
		return
	}

	u = gRouteParamRegexp.ReplaceAllStringFunc(page.path, repl)
	if i != n {
		err = fmt.Errorf("Too many parameters supplied for the route")
		return
	}
	return
}

func (pm *PageManager) bind(params map[string]interface{}) {
	models := make([]interface{}, 0)

	pc := &PageCtrl{params, pm, make([]string, 0)}

	if controller := pm.currentPage.handlable.controller; controller != nil {
		models = append(models, controller(pc))
	}

	for _, handler := range pm.currentPage.handlers {
		handler()
	}

	for _, grp := range pm.currentPage.groups {
		if controller := grp.handlable.controller; controller != nil {
			models = append(models, controller(pc))
		}

		for _, handler := range grp.handlable.handlers {
			handler()
		}
	}

	if len(models) == 0 {
		pm.binding.Bind(pm.container, nil, true, false)
	} else {
		pm.binding.BindModels(pm.container, models, false, false)
	}

	pm.pc = pc
}

// RegisterController sets the controller function for the specified
// page / page group.
func (pm *PageManager) RegisterController(displayScope string, fn PageControllerFunc) {
	ds := pm.displayScope(displayScope)
	ds.setController(fn)
}

// RegisterHandler hooks a PageHandler to the specified page / page group
func (pm *PageManager) RegisterHandler(displayScope string, fn PageHandler) {
	ds := pm.displayScope(displayScope)
	ds.addHandler(fn)
}

type DisplayScope interface {
	Register(id string, pm *PageManager) displayScope
}

type Page struct {
	Route string
	Title string
}

func (p Page) Register(pageId string, pm *PageManager) displayScope {
	route := p.Route

	if _, exist := pm.displayScopes[pageId]; exist {
		panic(fmt.Sprintf(`Page or page group with id "%v" already registered.`, pageId))
	}

	pm.router.Call("add", []map[string]interface{}{
		map[string]interface{}{
			"path": route,
			"handler": func() string {
				return pageId
			},
		},
	})

	page := newPage(pageId, route, p.Title)
	pm.displayScopes[pageId] = page

	return page
}

func MakePage(route string, title string) Page {
	return Page{
		Route: route,
		Title: title,
	}
}

type PageGroup struct {
	pageids []string
}

func MakePageGroup(pageids ...string) PageGroup {
	return PageGroup{
		pageids: pageids,
	}
}

func (pg PageGroup) Register(id string, pm *PageManager) displayScope {
	grp := newPageGroup(make([]*page, len(pg.pageids)))
	for i, pid := range pg.pageids {
		page := pm.page(pid)
		page.addGroup(grp)
		grp.pages[i] = page
	}
	return grp
}

// RegisterDisplayScopes registers the given map of pages and page groups
func (pm *PageManager) RegisterDisplayScopes(m map[string]DisplayScope) {
	for id, item := range m {
		if id == "" {
			panic("id of page/page group cannot be empty.")
		}

		pm.displayScopes[id] = item.Register(id, pm)
	}
}
