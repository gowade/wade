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
	addParent(parent *pageGroup)
	Controllers() []PageControllerFunc
}

type page struct {
	handlable
	id    string
	path  string
	title string

	groups []*pageGroup
}

func (p *page) addParent(grp *pageGroup) {
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
	children []displayScope
	parents  []*pageGroup
}

func newPageGroup(children []displayScope) *pageGroup {
	return &pageGroup{
		children: children,
	}
}

func (pg *pageGroup) addParent(parent *pageGroup) {
	pg.parents = append(pg.parents, parent)
}

func (pg *pageGroup) hasPage(id string) bool {
	for _, c := range pg.children {
		if c.hasPage(id) {
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

func (s *globalDisplayScope) addParent(parent *pageGroup) {
	panic("Cannot add parent to global display scope")
}

type PageDesc struct {
	id    string
	route string
	title string
}

func (p PageDesc) Register(pm *pageManager) displayScope {
	route := p.route

	if _, exist := pm.displayScopes[p.id]; exist {
		panic(fmt.Sprintf(`Page or page group with id "%v" already registered.`, p.id))
	}

	page := newPage(p.id, route, p.title)
	pm.displayScopes[p.id] = page

	pm.addRoute(page)

	return page
}

func MakePage(id string, route string, title string) PageDesc {
	return PageDesc{
		id:    id,
		route: route,
		title: title,
	}
}

type PageGroupDesc struct {
	id       string
	children []string
}

func MakePageGroup(id string, children []string) PageGroupDesc {
	return PageGroupDesc{
		id:       id,
		children: children,
	}
}

func (pg PageGroupDesc) Register(pm *pageManager) displayScope {
	grp := newPageGroup(make([]displayScope, len(pg.children)))
	for i, id := range pg.children {
		ds := pm.displayScope(id)
		ds.addParent(grp)
		grp.children[i] = ds
	}
	return grp
}

// BaseScope provides access to the page data and operations inside a controller func
type BaseScope struct {
	PageInfo *PageInfo
	pm       *pageManager
	p        *page

	params  map[string]interface{}
	helpers map[string]interface{}
}

type PageInfo struct {
	Id    string
	Route string
	Title string
}

// Manager returns the page manager
func (pc *BaseScope) Manager() PageManager {
	return pc.pm
}

func (pc *BaseScope) needsToEmbedBaseScope() {
}

func (pc *BaseScope) RedirectToPage(page string, params ...interface{}) {
	pc.pm.RedirectToPage(page, params...)
}

func (pc *BaseScope) RedirectToUrl(url string) {
	pc.pm.RedirectToUrl(url)
}

// FormatTitle formats the page's title with the given params
func (pc *BaseScope) FormatTitle(params ...interface{}) {
	pc.pm.formattedTitle = fmt.Sprintf(pc.pm.currentPage.title, params...)
	pc.PageInfo.Title = pc.pm.formattedTitle
}

func (pc *BaseScope) ApplyChanges(object interface{}) {
	pc.pm.binding.Watcher().ApplyChanges(object)
}

func (pc *BaseScope) Apply() {
	pc.pm.binding.Watcher().Apply()
}

// GetParam puts the value of a parameter to a dest.
// The dest must be a pointer, typically it would be a pointer to a model's field,
// for example
//	pc.GetParam("postid", &pmodel.PostId)
func (pc *BaseScope) GetParam(param string, dest interface{}) (err error) {
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

func (pc *BaseScope) Services() GlobalServices {
	return AppServices
}

// RegisterHelper registers fn as a local helper with the given name.
func (pc *BaseScope) RegisterHelper(name string, fn interface{}) {
	pc.helpers[name] = fn
}

func (pc *BaseScope) Url(pageId string, params ...interface{}) (url string, err error) {
	return pc.pm.pageUrl(pageId, params)
}
