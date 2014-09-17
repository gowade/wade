package wade

import "fmt"

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
