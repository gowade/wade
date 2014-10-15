package wade

import "fmt"

type (
	handlable struct {
		controllers []PageControllerFunc
	}

	displayScope interface {
		hasPage(id string) bool
		addController(fn PageControllerFunc)
		addParent(parent *pageGroup)
		Controllers() []PageControllerFunc
	}

	Page struct {
		Id    string
		Title string
	}

	page struct {
		handlable
		Page

		route  string
		groups []*pageGroup
	}

	pageGroup struct {
		handlable
		children []displayScope
		parents  []*pageGroup
	}
)

func (h *handlable) addController(fn PageControllerFunc) {
	if h.controllers == nil {
		h.controllers = make([]PageControllerFunc, 0)
	}
	h.controllers = append(h.controllers, fn)
}

func (h *handlable) Controllers() []PageControllerFunc {
	return h.controllers
}

func (p *page) addParent(grp *pageGroup) {
	if p.groups == nil {
		p.groups = make([]*pageGroup, 0)
	}

	p.groups = append(p.groups, grp)
}

func (p *page) hasPage(id string) bool {
	return p.Id == id
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

type globalDisplayScope struct {
	handlable
}

func (s *globalDisplayScope) hasPage(id string) bool {
	return true
}

func (s *globalDisplayScope) addParent(parent *pageGroup) {
	panic("Cannot add parent to global display scope")
}

func (p Page) Register(pm *pageManager, route string) RouteHandler {
	if _, exist := pm.displayScopes[p.Id]; exist {
		panic(fmt.Sprintf(`Page or page group with id "%v" has already been registered.`, p.Id))
	}

	pg := &page{
		Page:   p,
		route:  route,
		groups: []*pageGroup{},
	}

	pm.displayScopes[p.Id] = pg

	return pg
}

func (p *page) UpdatePage(pm *pageManager, pu pageUpdate) (err error, found bool) {
	found = true
	pm.updatePage(p, pu)

	return
}
