package wade

import (
	"fmt"
	gourl "net/url"
	"path"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	urlrouter "github.com/naoina/kocha-urlrouter"

	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
)

const (
	GlobalDisplayScope       = "__global__"
	ProductionPageLinkHandle = false
)

var (
	gFinishChan         = make(chan error, 100)
	gPageProcessingLock *lock
)

type (
	lock struct{}

	bindEngine interface {
		Watcher() *bind.Watcher
		BindModels(root dom.Selection, models []interface{}, once bool)
		RegisterInternalHelpers(pm bind.PageManager)
	}

	History interface {
		ReplaceState(title string, path string)
		PushState(title string, path string)
		OnPopState(fn func())
		CurrentPath() string
		RedirectTo(url string)
	}

	pageManager struct {
		app           *Application
		document      dom.Selection
		sourceElem    dom.Selection
		router        *Router
		currentPage   *page
		basePath      string
		notFoundPage  *page
		rcProto       string
		container     dom.Selection
		tcontainer    dom.Selection
		realContainer dom.Selection

		binding        bindEngine
		scope          *PageScope
		displayScopes  map[string]displayScope
		globalDs       *globalDisplayScope
		formattedTitle string
		history        History
	}
)

func newPageManager(app *Application, history History, document dom.Selection,
	tcontainer dom.Selection, sourceElem dom.Selection, binding bindEngine) *pageManager {
	realContainer := document.Find("[w-app-container]").First()

	if realContainer.Length() == 0 {
		panic("App container doesn't exist.")
	}

	basePath := app.Config.BasePath
	if basePath == "" {
		basePath = "/"
	}

	cl := realContainer.Clone()
	cl.SetHtml("")

	pm := &pageManager{
		sourceElem:    sourceElem,
		app:           app,
		document:      document,
		router:        newRouter(nil),
		currentPage:   nil,
		basePath:      basePath,
		notFoundPage:  nil,
		rcProto:       cl.OuterHtml(),
		tcontainer:    tcontainer,
		realContainer: realContainer,
		binding:       binding,
		displayScopes: make(map[string]displayScope),
		globalDs:      &globalDisplayScope{},
		history:       history,
	}

	pm.router.pm = pm

	pm.displayScopes[GlobalDisplayScope] = pm.globalDs
	return pm
}

func (pm *pageManager) CurrentPageId() string {
	return pm.currentPage.Id
}

func (pm *pageManager) cutPath(spath string) string {
	if strings.HasPrefix(spath, pm.basePath) {
		spath = spath[len(pm.basePath):]
		if !strings.HasPrefix(spath, "/") {
			spath = "/" + spath
		}
	}
	return spath
}

func (pm *pageManager) page(id string) *page {
	if ds, hasDs := pm.displayScopes[id]; hasDs {
		if page, ok := ds.(*page); ok {
			return page
		}
	}

	panic(fmt.Sprintf(`No such page "%v"`, id))
}

func (pm *pageManager) SetNotFoundPage(pageId string) {
	pm.notFoundPage = pm.page(pageId)
}

// Url returns the full path
func (pm *pageManager) FullPath(pa string) string {
	return path.Join(pm.basePath, pa)
}

func (pm *pageManager) GoToPage(page string, params ...interface{}) (found bool) {
	url, err := pm.PageUrl(page, params...)
	if err != nil {
		panic(err.Error())
	}

	_, found = pm.updateUrl(url, true, false)
	return
}

func (pm *pageManager) GoToUrl(url string) (found bool) {
	if strings.HasPrefix(url, pm.BasePath()) {
		_, found = pm.updateUrl(url, true, false)
	} else {
		found = true
		pm.history.RedirectTo(url)
	}

	return
}

func (pm *pageManager) prepare() {
	pm.router.Build()

	pm.history.OnPopState(func() {
		go func() {
			pm.updateUrl(pm.history.CurrentPath(), false, false)
		}()
	})

	p := pm.history.CurrentPath()

	err, _ := pm.updateUrl(p, false, true)

	if err != nil {
		pm.app.ErrChanPut(err)
	}
	return
}

func (pm *pageManager) walk(elem dom.Selection, remove bool) (err error) {
	pendingIncludes := 0
	for _, e := range elem.Children().Elements() {
		belong, hasbelong := e.Attr("w-belong")
		excluded := false
		if hasbelong {
			list := strings.Split(belong, " ")

			excluded = true
			for _, belong := range list {
				belong = strings.TrimSpace(belong)
				if ds, ok := pm.displayScopes[belong]; ok {
					if ds.hasPage(pm.currentPage.Id) {
						excluded = false
						break
					}
				} else {
					panic(fmt.Sprintf(`Invalid value "%v" for w-belong, no such page or page group is registered.`, belong))
				}
			}
		}

		if !excluded {
			if e.Is("winclude") {
				pendingIncludes++
				go func(e dom.Selection) {
					html, err := htmlInclude(pm.app.Http(),
						e, pm.app.Config.ServerBase)
					if err == nil {
						ne := elem.NewFragment("<ww>" + html + "</ww>")
						if hasbelong {
							ne.SetAttr("w-belong", belong)
						}
						e.ReplaceWith(ne)
						err = pm.walk(ne, remove)
						if err != nil {
							return
						}
					}
					gFinishChan <- err
				}(e)
			} else {
				pm.walk(e, remove)
			}
		} else {
			if remove {
				e.Remove()
			}
		}
	}

	for i := 0; i < pendingIncludes; i++ {
		err = <-gFinishChan
		if err != nil {
			return
		}
	}

	return
}

func newHiddenContainer(rcProto string, document dom.Dom) dom.Selection {
	hiddenRoot := document.NewRootFragment()
	container := document.NewFragment(rcProto)
	hiddenRoot.Append(container)
	return container
}

type scrollItem struct {
	elem dom.Selection
	posx int
	posy int
}

type scrollPreserver struct {
	scrolls []scrollItem
}

func (sp *scrollPreserver) getScroll(oldElem dom.Selection) {
	if posy, posx := oldElem.Underlying().Call("scrollTop").Int(),
		oldElem.Underlying().Call("scrollLeft").Int(); posx != 0 || posy != 0 {
		sp.scrolls = append(sp.scrolls, scrollItem{oldElem, posx, posy})
	}
}

func (sp *scrollPreserver) applyScrolls(newCtn dom.Selection) {
	for _, item := range sp.scrolls {
		ne := dom.GetElemCounterpart(item.elem, newCtn)
		if item.posy != 0 {
			ne.Underlying().Call("scrollTop", item.posy)
		}
		if item.posx != 0 {
			ne.Underlying().Call("scrollLeft", item.posx)
		}
	}
}

type pageUpdate struct {
	url         *gourl.URL
	routeParams []urlrouter.Param
	pushState   bool
	firstLoad   bool
}

func (pm *pageManager) updateUrl(url string, pushState bool, firstLoad bool) (err error, found bool) {
	u, err := gourl.Parse(pm.cutPath(url))
	if err != nil {
		return
	}

	match, routeparams := pm.router.Lookup(u.Path)
	pu := pageUpdate{
		url:         u,
		routeParams: routeparams,
		pushState:   pushState,
		firstLoad:   firstLoad,
	}

	if match == nil {
		if pm.router.notFoundHandler == nil {
			err = fmt.Errorf("404 page not found. No handler for page not found has been set.")
			return
		}

		//gopherjs:blocking
		pm.router.notFoundHandler.UpdatePage(pm, pu)
		return
	}

	found = true

	//gopherjs:blocking
	err, found = match.UpdatePage(pm, pu)

	return
}

func (pm *pageManager) updatePage(page *page, pu pageUpdate) {
	lck := &lock{}
	gPageProcessingLock = lck

	if pu.pushState {
		pm.history.PushState(page.Title, pm.FullPath(pu.url.Path))
	}

	namedParams := http.NewNamedParams(pu.routeParams)

	pm.formattedTitle = page.Title

	pm.currentPage = page

	if !ClientSide {
		err := pm.walk(pm.tcontainer, false)
		if err != nil {
			panic(err)
		}

		pm.sourceElem.SetHtml(pm.tcontainer.Html())
		pm.app.tm.ResolveTemplates(pm.tcontainer, false)
	}

	pm.container = newHiddenContainer(pm.rcProto, pm.document)
	pm.container.SetHtml(pm.tcontainer.Html())

	err := pm.walk(pm.container, true)

	if !ClientSide {
		pm.container.Find("template").Remove()
	} else {
		pm.app.tm.ResolveTemplates(pm.container, true)
	}

	if err != nil {
		panic(err)
	}

	pm.binding.Watcher().ResetWatchers()
	pm.bind(namedParams, pu.url)

	if gPageProcessingLock != lck {
		return
	}

	pm.setTitle(pm.formattedTitle)

	if ClientSide {
		if gPageProcessingLock == lck {
			jqwindow := js.Global.Call("jQuery", js.Global.Get("window"))
			if pu.firstLoad {
				scrollpos := jqwindow.Call("scrollTop")
				sp := &scrollPreserver{[]scrollItem{}}
				for _, c := range pm.realContainer.Find(".w-scrolled").Elements() {
					sp.getScroll(c)
				}

				pm.realContainer.ReplaceWith(pm.container)
				sp.applyScrolls(pm.container)
				jqwindow.Call("scrollTop", scrollpos)
			} else {
				pm.realContainer.ReplaceWith(pm.container)
				jqwindow.Call("scrollTop", 0)
			}
		}
	} else {
		pm.realContainer.ReplaceWith(pm.container)
	}

	pm.realContainer = pm.container

	//Handle link events on Dev mode
	if ProductionPageLinkHandle || DevMode {
		pm.realContainer.Listen("click", "a", func(e dom.Event) {
			href, ok := e.Target().Attr("href")
			if !ok {
				return
			}

			if !strings.HasPrefix(href, pm.BasePath()) { //not a wade page link, let the browser do its job
				return
			}

			e.PreventDefault()

			go func() {
				pm.updateUrl(href, true, false)
			}()
		})
	}

	return
}

func (pm *pageManager) setTitle(title string) {
	tElem := pm.document.NewFragment("<title>" + title + "</title>")
	head := pm.document.Find("head").First()
	oElem := head.Find("title")
	if oElem.Length() == 0 {
		head.Append(tElem)
	} else {
		oElem.ReplaceWith(tElem)
	}
}

// PageUrl returns the url for the page with the given parameters
func (pm *pageManager) PageUrl(pageId string, params ...interface{}) (u string, err error) {
	u, err = pm.pageUrl(pageId, params)
	u = pm.FullPath(u)
	return
}

func (pm *pageManager) pageUrl(pageId string, params []interface{}) (u string, err error) {
	err = nil
	page := pm.page(pageId)

	k, i := 0, 0
	route := page.route
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
			pageId, len(routeparams), len(params))
		return
	}

	return
}

func (pm *pageManager) BasePath() string {
	return pm.basePath
}

func (pm *pageManager) CurrentPage() *PageScope {
	return pm.scope
}

func (pm *pageManager) bind(namedParams *http.NamedParams, url *gourl.URL) {
	s := pm.newPageScope(pm.currentPage, namedParams, url)
	controllers := make([]PageControllerFunc, 0)

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
		for i, controller := range controllers {
			go func(controller PageControllerFunc) {
				s.ModelHolder.inMainCtrl = (i == len(controllers)-1)
				//gopherjs:blocking
				err := controller(s)
				if err != nil {
					pm.app.ErrChanPut(err)
				}

				queueChan <- true
				if len(queueChan) == len(controllers) {
					completeChan <- true
				}
			}(controller)
		}
		<-completeChan
	}

	//gopherjs:blocking
	pm.binding.BindModels(pm.container, s.bindModels(), false)

	pm.scope = s

	pm.binding.Watcher().Checkpoint()
}

// RegisterController adds a new controller function for the specified
// page / page group.
func (pm *pageManager) registerController(displayScope string, fn PageControllerFunc) {
	ds, ok := pm.displayScopes[displayScope]
	if !ok {
		panic(fmt.Errorf(`Registering controller for "%v", there's no such page or page group.`, displayScope))
	}

	ds.addController(fn)
}

//// RegisterDisplayScopes registers the given maps of pages and pageGroups
//func (pm *pageManager) registerDisplayScopes(pages []PageDesc, pageGroups []PageGroupDesc) {
//	if pages != nil {
//		for _, pg := range pages {
//			pm.displayScopes[pg.id] = pg.Register(pm)
//		}
//	}

//	if pageGroups != nil {
//		for _, pg := range pageGroups {
//			pm.displayScopes[pg.id] = pg.Register(pm)
//		}
//	}
//}

func (pm *pageManager) registerPageGroup(pgid string, children []string) {
	if _, exist := pm.displayScopes[pgid]; exist {
		panic(fmt.Sprintf(`Page or page group with id "%v" has already been registered.`, pgid))
	}

	grp := newPageGroup(make([]displayScope, len(children)))
	for i, id := range children {
		ds, ok := pm.displayScopes[id]
		if !ok {
			panic(fmt.Errorf(`Wrong children for page group "%v", there's no page or page group with id "%v".`, pgid, id))
		}
		ds.addParent(grp)
		grp.children[i] = ds
	}

	pm.displayScopes[pgid] = grp
}
