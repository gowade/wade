package wade

import (
	"fmt"
	"reflect"
)

type handlable struct {
	controllers []PageControllerFunc
}

func (h *handlable) addController(fn PageControllerFunc) {
	if h.controllers == nil {
		h.controllers = make([]PageControllerFunc, 0)
	}
	h.controllers = append(h.controllers, fn)
}

func (h *handlable) Controllers() []PageControllerFunc {
	return h.controllers
}

type displayScope interface {
	hasPage(id string) bool
	addController(fn PageControllerFunc)
	Controllers() []PageControllerFunc
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

type globalDisplayScope struct {
	handlable
}

func (s *globalDisplayScope) hasPage(id string) bool {
	return true
}

type DisplayScope interface {
	Register(id string, pm *pageManager) displayScope
}

type PageDesc struct {
	route string
	title string
}

func (p PageDesc) Register(pageId string, pm *pageManager) displayScope {
	route := p.route

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

	page := newPage(pageId, route, p.title)
	pm.displayScopes[pageId] = page

	return page
}

func MakePage(route string, title string) PageDesc {
	return PageDesc{
		route: route,
		title: title,
	}
}

type PageGroupDesc struct {
	pageids []string
}

func MakePageGroup(pageids ...string) PageGroupDesc {
	return PageGroupDesc{
		pageids: pageids,
	}
}

func (pg PageGroupDesc) Register(id string, pm *pageManager) displayScope {
	grp := newPageGroup(make([]*page, len(pg.pageids)))
	for i, pid := range pg.pageids {
		page := pm.page(pid)
		page.addGroup(grp)
		grp.pages[i] = page
	}
	return grp
}

// PageCtrl provides access to the page data and operations inside a controller func
type PageCtrl struct {
	app AppEnv
	pm  *pageManager
	p   *page

	params  map[string]interface{}
	helpers map[string]interface{}
}

type PageInfo struct {
	Id    string
	Route string
	Title string
}

// PageInfo returns information about the page
func (pc *PageCtrl) Info() PageInfo {
	return PageInfo{
		Id:    pc.p.id,
		Route: pc.p.path,
		Title: pc.p.title,
	}
}

// Manager returns the page manager
func (pc *PageCtrl) Manager() PageManager {
	return pc.pm
}

// Services returns AppEnv Services
func (pc *PageCtrl) Services() AppServices {
	return pc.app.Services
}

func (pc *PageCtrl) RedirectToPage(page string, params ...interface{}) {
	pc.pm.RedirectToPage(page, params...)
}

func (pc *PageCtrl) RedirectToUrl(url string) {
	pc.pm.RedirectToUrl(url)
}

// FormatTitle formats the page's title with the given params
func (pc *PageCtrl) FormatTitle(params ...interface{}) {
	pc.pm.formattedTitle = fmt.Sprintf(pc.pm.currentPage.title, params...)
}

// GetParam puts the value of a parameter to a dest.
// The dest must be a pointer, typically it would be a pointer to a model's field,
// for example
//	pc.GetParam("postid", &pmodel.PostId)
func (pc *PageCtrl) GetParam(param string, dest interface{}) (err error) {
	v, ok := pc.params[param]
	if !ok {
		err = fmt.Errorf("No such parameter %v.", param)
		return
	}

	if reflect.TypeOf(dest).Kind() != reflect.Ptr {
		return fmt.Errorf("The dest for saving the parameter value must be a pointer so that it could be modified.")
	}
	_, err = fmt.Sscan(v.(string), dest)
	return
}

// RegisterHelper registers fn as a local helper with the given name.
func (pc *PageCtrl) RegisterHelper(name string, fn interface{}) {
	pc.helpers[name] = fn
}
