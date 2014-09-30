package wade

import (
	"fmt"
	"reflect"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
)

// Scope provides access to the page data and operations inside a controller func
type (
	ObserveCallback func(oldVal, newVal interface{}, document dom.Dom)

	Scope struct {
		PageInfo *PageInfo
		pm       *pageManager
		p        *page

		params map[string]interface{}

		valMap map[string]interface{}
		models []interface{}
	}

	PageInfo struct {
		Id    string
		Route string
		Title string
	}
)

func (pm *pageManager) newRootScope(page *page, params map[string]interface{}) *Scope {
	return &Scope{
		PageInfo: &PageInfo{
			Id:    page.id,
			Title: page.title,
			Route: page.path,
		},
		pm:     pm,
		p:      page,
		params: params,
		valMap: make(map[string]interface{}),
		models: make([]interface{}, 0),
	}
}

// Manager returns the page manager
func (pc *Scope) Manager() PageManager {
	return pc.pm
}

func (pc *Scope) RedirectToPage(page string, params ...interface{}) {
	pc.pm.RedirectToPage(page, params...)
}

func (pc *Scope) RedirectToUrl(url string) {
	pc.pm.RedirectToUrl(url)
}

// FormatTitle formats the page's title with the given params
func (pc *Scope) FormatTitle(params ...interface{}) {
	pc.pm.formattedTitle = fmt.Sprintf(pc.pm.currentPage.title, params...)
	pc.PageInfo.Title = pc.pm.formattedTitle
}

func (pc *Scope) Digest(object interface{}) {
	pc.pm.binding.Watcher().Digest(object)
}

func (pc *Scope) Observe(model interface{}, field string, callback ObserveCallback) {
	pc.pm.binding.Watcher().Observe(model, field, func(oldVal, newVal interface{}) {
		callback(oldVal, newVal, pc.pm.document)
	})
}

// GetParam puts the value of a parameter to a dest.
// The dest must be a pointer, typically it would be a pointer to a model's field,
// for example
//	pc.GetParam("postid", &pmodel.PostId)
func (pc *Scope) GetParam(param string, dest interface{}) (err error) {
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

func (pc *Scope) Services() GlobalServices {
	return AppServices
}

func (pc *Scope) Url(pageId string, namedParams ...interface{}) (url string, err error) {
	return pc.pm.PageUrl(pageId, namedParams...)
}

func (pc *Scope) AddValue(name string, value interface{}) {
	pc.valMap[name] = value
}

func (pc *Scope) AddModel(model interface{}) {
	pc.models = append(pc.models, model)
}

func (pc *Scope) bindModels() []interface{} {
	return append(pc.models, pc.valMap, map[string]interface{}{
		"_pageInfo": pc.PageInfo,
	})
}

func (pc *Scope) Http() *http.Client {
	return pc.Services().Http
}
