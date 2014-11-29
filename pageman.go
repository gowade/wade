package wade

import (
	"fmt"
	gourl "net/url"
	"path"
	"strings"

	urlrouter "github.com/naoina/kocha-urlrouter"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/libs/http"
)

type (
	bindEngine interface {
		Bind(root *core.VNode, models ...interface{})
	}

	// Context provides access to the page data and operations inside a controller func
	Context struct {
		*PageManager
		NamedParams *http.NamedParams
		URL         *gourl.URL
	}

	History interface {
		ReplaceState(title string, path string)
		PushState(title string, path string)
		OnPopState(fn func())
		CurrentPath() string
		RedirectTo(url string)
	}

	OutputManager interface {
		RenderPage(title string, condFn core.CondFn)
		VirtualDOM() *core.VNode
	}

	PageManager struct {
		output         OutputManager
		binding        bindEngine
		basePath       string
		router         *router
		currentPage    *page
		ctx            Context
		formattedTitle string

		displayScopes map[string]displayScope
		history       History
	}
)

func NewPageManager(basePath string, history History,
	output OutputManager, bindEngine bindEngine) *PageManager {
	pm := &PageManager{
		output:        output,
		basePath:      basePath,
		history:       history,
		binding:       bindEngine,
		router:        newRouter(),
		displayScopes: map[string]displayScope{},
	}

	return pm
}

func (pm *PageManager) RouteMgr() Router {
	return Router{
		router: pm.router,
		pm:     pm,
	}
}

func (pm *PageManager) Context() Context {
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

func (pm *PageManager) GoToPage(page string, params ...interface{}) (found bool) {
	url := pm.PageUrl(page, params...)
	found = pm.updateUrl(url, true, false)
	return
}

func (pm *PageManager) GoToUrl(url string) (found bool) {
	if strings.HasPrefix(url, pm.BasePath()) {
		found = pm.updateUrl(url, true, false)
	} else {
		found = true
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
			err = fmt.Errorf("404 page not found. No handler for page not found has been set.")
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

	pm.binding.Bind(pm.output.VirtualDOM(),
		pm.runControllers(http.NewNamedParams(pu.routeParams), pu.url)...)

	pm.output.RenderPage(pm.formattedTitle,
		func(vnode core.VNode) bool {
			if belongstr, ok := vnode.Attr(core.BelongAttrName); ok {
				belongs := strings.Split(belongstr.(string), " ")
				for _, belong := range belongs {
					if ds, ok := pm.displayScopes[belong]; ok {
						if ds.hasPage(pm.currentPage.Id) {
							return true
						}
					} else {
						panic(fmt.Errorf(`In !belong specification %v:
					no such page or page group with id "%v"`, belongstr, belong))
					}
				}

				return false
			}

			return true
		})
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

func (pm *PageManager) runControllers(namedParams *http.NamedParams, url *gourl.URL) []interface{} {
	pm.ctx = Context{
		PageManager: pm,
		NamedParams: namedParams,
		URL:         url,
	}

	controllers := make([]ControllerFunc, 0)

	add := func(ds displayScope) {
		if ctrls := ds.Controllers(); ctrls != nil {
			for _, controller := range ctrls {
				controllers = append(controllers, controller)
			}
		}
	}

	add(pm.currentPage)

	for _, grp := range pm.currentPage.groups {
		add(grp)
	}

	add(GlobalDisplayScope)

	models := []interface{}{}

	if len(controllers) > 0 {
		for _, controller := range controllers {
			//gopherjs:blocking
			models = append(models, controller(pm.ctx))
		}
	}

	return models
}

func (pm *PageManager) AddPageGroup(pg PageGroup) {
	if _, exist := pm.displayScopes[pg.Id]; exist {
		panic(fmt.Sprintf(`Page or page group with id "%v" has already been registered.`, pg.Id))
	}

	grp := newPageGroup(make([]displayScope, len(pg.Children)))
	for i, id := range pg.Children {
		ds, ok := pm.displayScopes[id]
		if !ok {
			panic(fmt.Errorf(`Wrong children for page group "%v",
			there's no page or page group with id "%v".`, pg.Id, id))
		}

		ds.addParent(grp)
		grp.children[i] = ds
	}

	grp.AddController(pg.Controller)
	pm.displayScopes[pg.Id] = grp
}
