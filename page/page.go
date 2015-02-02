package page

import (
	"fmt"

	"github.com/phaikawl/wade/core"
)

var (
	GlobalDisplayScope = &globalDisplayScope{}
)

type (
	// PageControllerFunc is the functiong to be run on the load of a page or page scope
	PageControllerFunc      func(*Context) *core.VNode
	PageGroupControllerFunc func(*Context)
)

type (
	Page struct {
		Id         string
		Title      string
		Controller PageControllerFunc
	}

	PageGroup struct {
		Id         string
		Children   []string
		Controller PageGroupControllerFunc
	}
)

type (
	DisplayScope interface {
		HasPage(id string) bool
		addParent(parent *pageGroup)
	}

	page struct {
		Page

		route  string
		groups []*pageGroup
	}

	pageGroup struct {
		PageGroup
		children []DisplayScope
		parents  []*pageGroup
	}
)

func (p *page) addParent(grp *pageGroup) {
	if p.groups == nil {
		p.groups = make([]*pageGroup, 0)
	}

	p.groups = append(p.groups, grp)
}

func (p *page) HasPage(id string) bool {
	return p.Id == id
}

func newPageGroup(children []DisplayScope) *pageGroup {
	return &pageGroup{
		children: children,
	}
}

func (pg *pageGroup) addParent(parent *pageGroup) {
	pg.parents = append(pg.parents, parent)
}

func (pg *pageGroup) HasPage(id string) bool {
	for _, c := range pg.children {
		if c.HasPage(id) {
			return true
		}
	}

	return false
}

func (pg *pageGroup) setPGController(controller PageGroupControllerFunc) {
	pg.Controller = controller
}

func (pg PageGroup) AddTo(dscopes map[string]DisplayScope) error {
	if _, exist := dscopes[pg.Id]; exist {
		return fmt.Errorf(`Page or page group with id "%v" has already been registered.`, pg.Id)
	}

	grp := newPageGroup(make([]DisplayScope, len(pg.Children)))
	for i, id := range pg.Children {
		ds, ok := dscopes[id]
		if !ok {
			return fmt.Errorf(`Wrong children for page group "%v",
			there's no page or page group with id "%v".`, pg.Id, id)
		}

		ds.addParent(grp)
		grp.children[i] = ds
	}

	dscopes[pg.Id] = grp

	return nil
}

type globalDisplayScope struct {
	Controller PageGroupControllerFunc
}

func (s *globalDisplayScope) setPGController(controller PageGroupControllerFunc) {
	s.Controller = controller
}

func (s *globalDisplayScope) HasPage(id string) bool {
	return true
}

func (s *globalDisplayScope) addParent(parent *pageGroup) {
	panic("Cannot add parent to global display scope")
}

func (p Page) AddTo(dscopes map[string]DisplayScope) error {
	if _, exist := dscopes[p.Id]; exist {
		return fmt.Errorf(`Page or page group with id "%v" has already been registered.`, p.Id)
	}

	dscopes[p.Id] = &page{Page: p}
	return nil
}

func (p Page) Register(pm *PageManager, route string) RouteHandler {
	if _, exist := pm.displayScopes[p.Id]; exist {
		panic(fmt.Sprintf(`Page or page group with id "%v" has already been registered.`, p.Id))
	}

	pp := &page{
		Page:   p,
		route:  route,
		groups: []*pageGroup{},
	}

	pm.displayScopes[p.Id] = pp

	return pp
}

func (p *page) UpdatePage(pm *PageManager, pu pageUpdate) (found bool) {
	pm.updatePage(p, pu)

	return true
}
