package page

import (
	"fmt"
	gourl "net/url"
	"path"
	"strings"

	urlrouter "github.com/naoina/kocha-urlrouter"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
	//"github.com/phaikawl/wade/vquery"
	"github.com/phaikawl/wade/libs/http"
)

const (
	appViewAttr = "!appview"
)

type (
	// Context provides access to the page data and operations inside a controller func
	Context struct {
		*PageManager
		NamedParams *http.NamedParams
		URL         *gourl.URL
		redirected bool
	}

	History interface {
		ReplaceState(title string, path string)
		PushState(title string, path string)
		OnPopState(fn func())
		CurrentPath() string
		RedirectTo(url string)
	}

	PageManager struct {
		basePath       string
		router         *router
		currentPage    *page
		ctx            *Context
		formattedTitle string

		displayScopes map[string]DisplayScope
		history       History
		document      dom.Selection
		container     dom.Selection
		titleElem     dom.Selection
		markup        *core.VNode
	}
)

func (ctx *Context) GoToPage(page string, params ...interface{}) (dummy *core.VNode) {
	ctx.redirected = true
	ctx.GoToPage(page, params...)
	return nil
}

func NewPageManager(basePath string, history History,
	document dom.Selection) *PageManager {
	c := document.Find("[\\" + appViewAttr + "]")
	if c.Length() == 0 {
		panic(fmt.Errorf(`No view container (element with "%v" attribute) found.`, appViewAttr))
	}

	headElem := document.Find("head").First()
	titleElem := headElem.Find("title")
	if titleElem.Length() == 0 {
		titleElem = document.NewFragment("<title></title>")
		headElem.Append(titleElem)
	}

	pm := &PageManager{
		basePath:      basePath,
		history:       history,
		router:        newRouter(),
		displayScopes: map[string]DisplayScope{},
		document:      document,
		container:     c.First(),
		titleElem:     titleElem,
	}

	return pm
}

func (pm *PageManager) Document() dom.Selection {
	return pm.document
}

func (pm *PageManager) Render() {
	pm.markup.Update()
	//for _, n := range vq.New(pm.markup).Find(vq.Selector{}) {
	//	println(n.DebugInfo())
	//}
	//js.Global.Get("console").Call("profile")
	pm.container.Render(pm.markup)
	//js.Global.Get("console").Call("profileEnd")
}

func (pm *PageManager) Router() Router {
	return Router{
		router: pm.router,
		pm:     pm,
	}
}

func (pm *PageManager) Context() *Context {
	return pm.ctx
}

// FormatTitle formats the page's title with the given param values
func (pm *PageManager) FormatTitle(params ...interface{}) string {
	pm.formattedTitle = fmt.Sprintf(pm.currentPage.Title, params...)
	return pm.formattedTitle
}

func (pm *PageManager) CurrentPage() Page {
	if pm.currentPage == nil {
		panic("the page manager has not been started")
	}

	return pm.currentPage.Page
}

func (pm *PageManager) cutPath(spath string) string {
	if strings.HasPrefix(spath, pm.basePath) {
		spath = spath[len(pm.basePath):]
		if !strings.HasPrefix(spath, "/") {
			spath = "/" + spath
		}
	}
	return spath
}

func (pm *PageManager) page(id string) *page {
	if ds, hasDs := pm.displayScopes[id]; hasDs {
		if page, ok := ds.(*page); ok {
			return page
		}
	}

	panic(fmt.Sprintf(`No such page "%v"`, id))
}

// Url returns the full path
func (pm *PageManager) FullPath(pa string) string {
	return path.Join(pm.basePath, pa)
}

func (pm *PageManager) GoToPage(page string, params ...interface{}) {
	url := pm.PageUrl(page, params...)
	pm.updateUrl(url, true, false)
}

func (pm *PageManager) GoToUrl(url string) {
	if strings.HasPrefix(url, pm.BasePath()) {
		pm.updateUrl(url, true, false)
	} else {
		pm.history.RedirectTo(url)
	}

	return
}

func (pm *PageManager) Start() {
	pm.router.build()

	pm.history.OnPopState(func() {
		go func() {
			pm.updateUrl(pm.history.CurrentPath(), false, false)
		}()
	})

	p := pm.history.CurrentPath()

	pm.updateUrl(p, false, true)

	return
}

type pageUpdate struct {
	url         *gourl.URL
	routeParams []urlrouter.Param
	pushState   bool
	firstLoad   bool
}

func (pm *PageManager) updateUrl(url string, pushState bool, firstLoad bool) bool {
	u, err := gourl.Parse(pm.cutPath(url))
	if err != nil {
		panic(err)
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
			panic(fmt.Errorf("404 page not found."))
			return false
		}

		//gopherjs:blocking
		pm.router.notFoundHandler.UpdatePage(pm, pu)
		return false
	}

	//gopherjs:blocking
	match.UpdatePage(pm, pu)

	return true
}

func (pm *PageManager) updatePage(page *page, pu pageUpdate) {
	if pu.pushState {
		pm.history.PushState(page.Title, pm.FullPath(pu.url.Path))
	}

	pm.currentPage = page
	pm.formattedTitle = page.Title

	//gopherjs:blocking
	pm.titleElem.SetHtml(pm.formattedTitle)
	tmpl, redirected := pm.runControllers(http.NewNamedParams(pu.routeParams), pu.url)
	
	if tmpl != nil && !redirected {
		pm.markup = tmpl
		pm.Render()
	}	
}

// PageUrl returns the url for the page with the given parameters
func (pm *PageManager) PageUrl(pageId string, params ...interface{}) string {
	u, err := pm.pageUrl(pageId, params)
	if err != nil {
		panic(err)
	}

	return pm.FullPath(u)
}

func (pm *PageManager) pageUrl(pageId string, params []interface{}) (u string, err error) {
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
	}

	return
}

func (pm *PageManager) BasePath() string {
	return pm.basePath
}

func (pm *PageManager) runControllers(namedParams *http.NamedParams, url *gourl.URL) (
tmpl *core.VNode, redirected bool) {
	ctx := Context{
		PageManager: pm,
		NamedParams: namedParams,
		URL:         url,
		redirected: false,
	}
	
	pm.ctx = &ctx

	if ctrl := GlobalDisplayScope.Controller; ctrl != nil {
		ctrl(&ctx)
	}

	for _, grp := range pm.currentPage.groups {
		if ctrl := grp.Controller; ctrl != nil {
			ctrl(&ctx)
		}
	}

	if ctrl := pm.CurrentPage().Controller; ctrl != nil {
		tmpl = ctrl(&ctx)
		redirected = ctx.redirected
		return
	}

	return nil, false
}

func (pm *PageManager) AddPageGroup(pg PageGroup) {
	err := pg.AddTo(pm.displayScopes)
	if err != nil {
		panic(err)
	}
}
