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

// RedirectToPage redirects to the given page with the given namedParams values
func (pc *Scope) RedirectToPage(page string, namedParams ...interface{}) {
	pc.pm.RedirectToPage(page, namedParams...)
}

func (pc *Scope) RedirectToUrl(url string) {
	pc.pm.RedirectToUrl(url)
}

// FormatTitle formats the page's title with the given param values
func (pc *Scope) FormatTitle(params ...interface{}) {
	pc.pm.formattedTitle = fmt.Sprintf(pc.pm.currentPage.title, params...)
	pc.PageInfo.Title = pc.pm.formattedTitle
}

// Digest manually triggers the observers for the given object.
// It must be a pointer, normally a pointer to a struct field.
func (pc *Scope) Digest(object interface{}) {
	pc.pm.binding.Watcher().Digest(object)
}

// Observe manually registers an observer for the given model, watching the given field
// and call the given callback when the the field changes
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

// Services returns the global services
func (pc *Scope) Services() GlobalServices {
	return AppServices
}

// Url returns the url for the given page pageId, with the given namedParams values
func (pc *Scope) Url(pageId string, namedParams ...interface{}) (url string, err error) {
	return pc.pm.PageUrl(pageId, namedParams...)
}

// AddValue adds a value to the scope and assigns it a given name
func (pc *Scope) AddValue(name string, value interface{}) {
	pc.valMap[name] = value
}

// AddModel adds a model to the scope, all exported struct fields of the model
// become valid symbols
func (pc *Scope) AddModel(model interface{}) {
	pc.models = append(pc.models, model)
}

func (pc *Scope) bindModels() []interface{} {
	return append(pc.models, pc.valMap, map[string]interface{}{
		"_pageInfo": pc.PageInfo,
	})
}

// Http is a convenient method which returns the default http client
func (pc *Scope) Http() *http.Client {
	return pc.Services().Http
}
