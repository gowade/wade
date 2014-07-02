package wade

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
)

const (
	WadePageAttr = "data-wade-page"
)

var (
	gRouteParamRegexp = regexp.MustCompile(`\:\w+`)
)

type pageInfo struct {
	path  string
	title string
}

type PageController func(*PageData) interface{}
type PageHandler func()

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
	binding         *Binding
	tm              *CustagMan
	//pageModels   []js.Object
}

type PageData struct {
	params map[string]interface{}
}

//ExportParam sets the value of a parameter to a target.
//The target must be a pointer, typically it would be a pointer to a model's field,
//for example
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

func newPageManager(startPage, basePath string, container string,
	tcontainer jq.JQuery, binding *Binding, tm *CustagMan) *PageManager {
	return &PageManager{
		router:          js.Global.Get("RouteRecognizer").New(),
		currentPage:     "",
		pageHandlers:    make(map[string][]PageHandler),
		pageControllers: make(map[string]PageController),
		basePath:        basePath,
		startPage:       startPage,
		pages:           make(map[string]pageInfo),
		notFoundPage:    "",
		container:       gJQ("#" + container),
		tcontainer:      tcontainer,
		binding:         binding,
		tm:              tm,
		//pageModels:   make([]js.Object, 0),
	}
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
	elem := parent.Find("#" + pageid)
	if elem.Length == 0 {
		panic(fmt.Sprintf("Cannot find element for page #%v.", pageid))
	}
	if WadeDevMode {
		elem = parent.Find("[id='" + pageid + "']")
		if elem.Length > 1 {
			panic(fmt.Sprintf("Unexpected: duplicated element id #%v when getting the page element.", pageid))
		}
		if !elem.Is("wpage") {
			panic(fmt.Sprintf("page element #%v must be a wpage.", pageid))
		}
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
			} else {
				resultElems[i-1].Append(clone)
			}
		}

		resultElems[leng] = gJQ(pageElem.Get(0).Call("cloneNode"))
		parents[leng] = pageElem
		for i := leng; i >= 0; i-- {
			p := parents[i]
			p.Contents().Each(func(_ int, e jq.JQuery) {
				if !e.Is("wpage") {
					nodeType := e.Get(0).Get("nodeType").Int()
					if nodeType == 1 {
						resultElems[i].Append(e.Get(0).Get("outerHTML"))
					} else if nodeType == 3 {
						resultElems[i].Append(e.Text())
					}
				} else if e.Is(pageElem.Get(0)) {
					resultElems[leng-1].Append(resultElems[leng])
				}
			})
		}

		pm.currentPage = pageId

		pm.bind(params)

		//Rebind link events
		pm.container.Find("a").On(jq.CLICK, func(e jq.Event) {
			a := gJQ(e.Target)

			pagepath := a.Attr(WadePageAttr)
			if pagepath == "" { //not a wade page link, let the browser do its job
				return
			}

			e.PreventDefault()

			pm.updatePage(pagepath, true)
		})
	}
}

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
	pageElem := elemForPage(pm.container, pm.currentPage)
	if handlers, ok := pm.pageHandlers[pm.currentPage]; ok {
		for _, handler := range handlers {
			handler()
		}
	}

	pdata := &PageData{params}
	if controller, exist := pm.pageControllers[pm.currentPage]; exist {
		model := controller(pdata)
		pm.binding.Bind(pageElem, model, false)
	}

	pm.binding.Bind(pm.container, nil, true)

	for tagName, tag := range pm.tm.custags {
		tagElem := pm.tcontainer.Find("#" + tag.meid)
		elems := pageElem.Find(tagName)
		elems.Each(func(i int, elem jq.JQuery) {
			elem.Append(tagElem.Html())
			pm.binding.Bind(elem, pm.tm.modelForElem(elem), false)
		})
	}
}

func (pm *PageManager) RegisterController(pageId string, fn PageController) {
	if _, exist := pm.pageControllers[pageId]; exist {
		panic(fmt.Sprintf("That page #%v already has a controller.", pageId))
	}
	pm.pageControllers[pageId] = fn
}

func (pm *PageManager) RegisterHandler(pageId string, fn PageHandler) {
	if _, exist := pm.pageHandlers[pageId]; !exist {
		pm.pageHandlers[pageId] = make([]PageHandler, 0)
	}
	pm.pageHandlers[pageId] = append(pm.pageHandlers[pageId], fn)
}

func (pm *PageManager) RegisterPages(pages map[string]string) {
	for path, pageId := range pages {
		if _, exist := pm.pages[pageId]; exist {
			panic(fmt.Sprintf("Page #%v has already been registered.", pageId))
		}
		pageElem := pm.tcontainer.Find("#" + pageId)
		if pageElem.Length == 0 {
			panic(fmt.Sprintf("There is no such page element #%v.", pageId))
		}

		(func(path, pageId string) {
			pm.router.Call("add", []map[string]interface{}{
				map[string]interface{}{
					"path": path,
					"handler": func() string {
						return pageId
					},
				},
			})
		})(path, pageId)

		pm.pages[pageId] = pageInfo{path: path, title: pageElem.Attr("title")}
	}
}
