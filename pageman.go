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

type pageInfo struct {
	path  string
	title string
}

// PageController is the function to be run on the load of a specific page.
// It returns a model to be used in bindings of the elements in the page.
type PageController func(*PageData) interface{}

// PageHandler is an additional function to be run on the load of a specific page,
// does not return anything.
type PageHandler func()

// PageManager is Page Manager
type PageManager struct {
	router          js.Object
	currentPage     string
	pageHandlers    map[string][]PageHandler
	pageControllers map[string]PageController
	startPage       string
	basePath        string
	pages           map[string]pageInfo
	notFoundPage    string
	container       jq.JQuery
	tcontainer      jq.JQuery
	binding         *bind.Binding
	tm              *CustagMan
	pd              *PageData
	//pageModels   []js.Object
}

// PageData provides access to the page-specific data inside a controller func
type PageData struct {
	params map[string]interface{}

	b       *bind.Binding
	helpers []string
}

// ExportParam sets the value of a parameter to a target.
// The target must be a pointer, typically it would be a pointer to a model's field,
// for example
//	pd.ExportParam("postid", &pmodel.PostId)
func (pd *PageData) ExportParam(param string, target interface{}) {
	v, ok := pd.params[param]
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
//
// Helpers registered with this method are deleted when switching page.
func (pd *PageData) RegisterHelper(name string, fn interface{}) {
	pd.helpers = append(pd.helpers, name)
	pd.b.RegisterHelper(name, fn)
}

func (pd *PageData) deleteHelpers() {
	for _, name := range pd.helpers {
		pd.b.DeleteHelper(name)
	}
}

func newPageManager(startPage, basePath string,
	tcontainer jq.JQuery, binding *bind.Binding, tm *CustagMan) *PageManager {

	container := gJQ("<div class='wade-wrapper'></div>")
	container.AppendTo(gJQ("body"))
	return &PageManager{
		router:          js.Global.Get("RouteRecognizer").New(),
		currentPage:     "",
		pageHandlers:    make(map[string][]PageHandler),
		pageControllers: make(map[string]PageController),
		basePath:        basePath,
		startPage:       startPage,
		pages:           make(map[string]pageInfo),
		notFoundPage:    "",
		container:       container,
		tcontainer:      tcontainer,
		binding:         binding,
		tm:              tm,
		//pageModels:   make([]js.Object, 0),
	}
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

func (pm *PageManager) page(pageId string) pageInfo {
	if page, ok := pm.pages[pageId]; ok {
		return page
	}
	panic(fmt.Sprintf("no such page #%v found.", pageId))
}

func (pm *PageManager) SetNotFoundPage(pageId string) {
	_ = pm.page(pageId)
	pm.notFoundPage = pageId
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

func elemForPage(parent jq.JQuery, pageid string) jq.JQuery {
	elem := parent.Find(fmt.Sprintf("wpage[pid='%v']", pageid))
	if elem.Length == 0 {
		panic(fmt.Sprintf(`Cannot find wpage element for page "%v".`, pageid))
	}
	return elem
}

func (pm *PageManager) setupPageOnLoad() {
	path := pm.cutPath(documentUrl())
	if path == "/" {
		path = pm.page(pm.startPage).path
		gHistory.Call("replaceState", nil, pm.pages[pm.startPage].title, pm.Url(path))
	}
	pm.updatePage(path, false)
}

func (pm *PageManager) getReady() {
	if pm.container.Length == 0 {
		panic(fmt.Sprintf("Cannot find the page container #%v.", pm.container))
	}

	//pm.tcontainer.Find("a").Each(func(idx int, a jq.JQuery) {
	//	href := a.Attr("href")
	//	if strings.HasPrefix(href, ":") {
	//		pageId := string([]rune(href)[1:])
	//		a.SetAttr("href", pm.Url(pm.page(pageId).path))
	//		a.SetAttr("data-wade-page", pageId)
	//	}
	//})

	gJQ(js.Global.Get("window")).On("popstate", func() {
		pm.updatePage(documentUrl(), false)
	})

	pm.setupPageOnLoad()
}

func (pm *PageManager) updatePage(url string, pushState bool) {
	url = pm.cutPath(url)
	matches := pm.router.Call("recognize", url)
	println("path: " + url)
	if matches.IsUndefined() || matches.Length() == 0 {
		if pm.notFoundPage != "" {
			pm.updatePage(pm.page(pm.notFoundPage).path, false)
		} else {
			panic("Page not found. No 404 handler declared.")
		}
	}

	match := matches.Index(0)
	pageId := match.Get("handler").Invoke().Str()
	if pushState {
		gHistory.Call("pushState", nil, pm.page(pageId).title, pm.Url(url))
	}
	params := make(map[string]interface{})
	prs := match.Get("params")
	if !prs.IsUndefined() {
		params = prs.Interface().(map[string]interface{})
	}

	pageElem := elemForPage(pm.tcontainer, pageId)
	gJQ("title").SetText(pm.page(pageId).title)
	if pm.currentPage != pageId {
		jqparents := pageElem.Parents("wpage")
		leng := jqparents.Length
		parents := make([]jq.JQuery, leng+1)
		resultElems := make([]jq.JQuery, leng+1)
		for i := 0; i < leng; i++ {
			parents[i] = jqparents.Eq(leng - i - 1)
			clone := gJQ(parents[i].Get(0).Call("cloneNode"))
			resultElems[i] = clone

			if i == 0 {
				c := pm.container.Children("*")
				if c.Length > 1 {
					panic("Page container should only have 1 child element. Something is wrong?")
				}
				if c.Length == 0 {
					pm.container.Append(resultElems[0])
				} else {
					c.First().ReplaceWith(resultElems[0])
				}
			}
		}

		resultElems[leng] = gJQ(pageElem.Get(0).Call("cloneNode"))
		parents[leng] = pageElem
		for i := leng; i >= 0; i-- {
			p := parents[i]
			p.Contents().Each(func(_ int, e jq.JQuery) {
				if !e.Is("wpage") {
					re := e
					if i < leng {
						e.Find("wpage").Each(func(_ int, ewpage jq.JQuery) {
							if pageElem.Closest(ewpage).Length == 0 {
								ewpage.SetAttr(WadeExcludeAttr, "t")
							}
						})

						re = e.Clone()
						re.Find("wpage").Each(func(_ int, wpage jq.JQuery) {
							if wpage.Attr(WadeExcludeAttr) != "" {
								wpage.Remove()
							}
						})

						e.Find("wpage").Each(func(_ int, ewpage jq.JQuery) {
							ewpage.RemoveAttr(WadeExcludeAttr)
						})
					}

					if re.Is("wrep") {
						re.Hide()
					}

					nodeType := e.Get(0).Get("nodeType").Int()
					if nodeType == 1 {
						resultElems[i].Append(re.Get(0).Get("outerHTML"))
						//println(re.Get(0).Get("outerHTML"))
					} else if nodeType == 3 {
						resultElems[i].Append(re.Text())
					}
				} else if i < leng && e.Is(parents[i+1].Get(0)) {
					resultElems[i].Append(resultElems[i+1])
				}
			})
		}

		pm.container.Find("wrep").Each(func(_ int, e jq.JQuery) {
			e.Remove()
			pm.container.Find("#" + WadeReservedPrefix + e.Attr("target")).
				SetHtml(e.Html())
		})

		pm.container.Find("wsection").Each(func(_ int, e jq.JQuery) {
			e.ReplaceWith(e.Html())
		})

		pm.currentPage = pageId

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

// PageUrl returns the url and route parameters for the specified pageid
func (pm *PageManager) PageUrl(pageid string, params []interface{}) (u string, err error) {
	err = nil
	route, ok := pm.pages[pageid]
	if !ok {
		err = fmt.Errorf(`No such page with id "%v".`, pageid)
	}

	n := len(params)
	if n == 0 {
		u = route.path
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

	u = gRouteParamRegexp.ReplaceAllStringFunc(route.path, repl)
	if i != n {
		err = fmt.Errorf("Too many parameters supplied for the route")
		return
	}
	return
}

func (pm *PageManager) bind(params map[string]interface{}) {
	if pm.pd != nil {
		pm.pd.deleteHelpers()
	}

	pageElem := elemForPage(pm.container, pm.currentPage)
	if handlers, ok := pm.pageHandlers[pm.currentPage]; ok {
		for _, handler := range handlers {
			handler()
		}
	}

	pm.pd = &PageData{params, pm.binding, make([]string, 0)}
	if controller, exist := pm.pageControllers[pm.currentPage]; exist {
		model := controller(pm.pd)
		pm.binding.Bind(pm.container, model, false)
	} else {
		stop := false
		pageElem.Parents("wpage").Each(func(_ int, p jq.JQuery) {
			if stop {
				return
			}

			if controller, exist = pm.pageControllers[p.Attr("pid")]; exist {
				pm.binding.Bind(pm.container, controller(pm.pd), false)
				stop = true
				return
			}
		})

		if !stop {
			pm.binding.Bind(pm.container, nil, true)
		}
	}

	for _, tag := range pm.tm.custags {
		tagElem := tag.elem
		elems := pm.container.Find(tag.name)
		elems.Each(func(i int, elem jq.JQuery) {
			elem.Append(tagElem.Html())
			pm.binding.Bind(elem, pm.tm.ModelForElem(elem), false)
		})
	}
}

// RegisterController assigns a PageController function to handle the specified
// page.
func (pm *PageManager) RegisterController(pageId string, fn PageController) {
	if _, exist := pm.pageControllers[pageId]; exist {
		panic(fmt.Sprintf("That page #%v already has a controller.", pageId))
	}
	pm.pageControllers[pageId] = fn
}

// RegisterHandler hooks a PageHandler to the specified page
func (pm *PageManager) RegisterHandler(pageId string, fn PageHandler) {
	if _, exist := pm.pageHandlers[pageId]; !exist {
		pm.pageHandlers[pageId] = make([]PageHandler, 0)
	}
	pm.pageHandlers[pageId] = append(pm.pageHandlers[pageId], fn)
}

// RegisterPages registers pages from the hierarchy of <wpage> inside a root wpage element
// with the given rootId.
//
// Each child <wpage> may have a "pid" (page id), a "route" and a "title" attribute.
//
// "pid" is the page's unique id. "title" is the page title.
//
// "route" is the page's route pattern which may contain
// route parameters like ":param1", ":postid". "route" is absolute path,
// it doesn't' use the parent page's route as base.
func (pm *PageManager) RegisterPages(rootId string) {
	root := pm.tcontainer.Find("#" + rootId)
	if root.Length == 0 {
		panic(fmt.Sprintf("Unable to find #%v, no such element.", rootId))
	}

	if !root.Is("wpage") {
		panic(fmt.Sprintf(`Root element #%v for RegisterPages must be a "wpage".`, rootId))
	}

	root.Find("wpage").Each(func(_ int, elem jq.JQuery) {
		pageId := elem.Attr("pid")
		if pageId != "" {
			route := elem.Attr("route")
			if route == "" {
				panic(fmt.Sprintf(`Page #%v does not have an associated route, please set its "route" attribute.`, pageId))
			}

			if _, exist := pm.pages[pageId]; exist {
				panic(fmt.Sprintf("Duplicate page id #%v.", pageId))
			}

			pm.router.Call("add", []map[string]interface{}{
				map[string]interface{}{
					"path": route,
					"handler": func() string {
						return pageId
					},
				},
			})

			pm.pages[pageId] = pageInfo{path: route, title: elem.Attr("title")}

			elem.SetAttr("id", WadeReservedPrefix+pageId)
		}
	})

	// preprocess wsection elements
	root.Find("wsection").Each(func(_ int, e jq.JQuery) {
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
}
