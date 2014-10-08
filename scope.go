package wade

import (
	"fmt"
	gourl "net/url"

	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
)

// Scope provides access to the page data and operations inside a controller func
type (
	ObserveCallback func(oldVal, newVal interface{}, document dom.Dom)

	Scope struct {
		App         *Application
		PageInfo    *PageInfo
		NamedParams *http.NamedParams
		URL         *gourl.URL

		pm     *pageManager
		p      *page
		valMap map[string]interface{}
		models []interface{}
	}

	PageInfo struct {
		Id    string
		Route string
		Title string
	}
)

func (pm *pageManager) newRootScope(page *page, params *http.NamedParams, url *gourl.URL) *Scope {
	return &Scope{
		PageInfo: &PageInfo{
			Id:    page.id,
			Title: page.title,
			Route: page.path,
		},
		App:         pm.app,
		NamedParams: params,
		URL:         url,
		pm:          pm,
		p:           page,
		valMap:      make(map[string]interface{}),
		models:      make([]interface{}, 0),
	}
}

// Manager returns the page manager
func (pc *Scope) Manager() PageManager {
	return pc.pm
}

// NavigatePage navigates to the given page with the given namedParams values
func (pc *Scope) GoToPage(page string, namedParams ...interface{}) {
	pc.pm.GoToPage(page, namedParams...)
}

func (pc *Scope) GoToUrl(url string) {
	pc.pm.GoToUrl(url)
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

// Services returns the global services
func (pc *Scope) Services() *AppServices {
	return pc.App.Services
}

// Url returns the url for the given page pageId, with the given namedParams values
func (pc *Scope) GetUrl(pageId string, namedParams ...interface{}) (url string, err error) {
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
