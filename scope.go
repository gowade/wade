package wade

import (
	"fmt"
	gourl "net/url"

	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
)

// Scope provides access to the page data and operations inside a controller func
type (
	ObserveCallback func(oldVal, newVal interface{}, document dom.Dom)

	Scope struct {
		*bind.Watcher
		*ModelHolder

		App         *Application
		NamedParams *http.NamedParams
		URL         *gourl.URL

		pm     *pageManager
		p      *page
		valMap map[string]interface{}
	}

	ModelHolder struct {
		inMainCtrl bool
		mainIndex  int
		mainName   string

		models      []interface{}
		namedModels map[string]interface{}
	}
)

func (pm *pageManager) newRootScope(page *page, params *http.NamedParams, url *gourl.URL) *Scope {
	return &Scope{
		ModelHolder: &ModelHolder{
			models:      []interface{}{},
			namedModels: map[string]interface{}{},
		},
		Watcher:     pm.binding.Watcher(),
		App:         pm.app,
		NamedParams: params,
		URL:         url,
		pm:          pm,
		p:           page,
		valMap:      make(map[string]interface{}),
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

func (pc *Scope) Page() Page {
	return pc.p.Page
}

// FormatTitle formats the page's title with the given param values
func (pc *Scope) FormatTitle(params ...interface{}) {
	pc.pm.formattedTitle = fmt.Sprintf(pc.pm.currentPage.Title, params...)
}

// Observe manually registers an observer for the given model, watching the given field
// and call the given callback when the the field changes
func (pc *Scope) Observe(model interface{}, field string, callback ObserveCallback) {
	pc.Watcher.Observe(model, field, func(oldVal, newVal interface{}) {
		callback(oldVal, newVal, pc.pm.document)
	})
}

// Services returns the global services
func (pc *Scope) Services() AppServices {
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

// SetModel sets the main model struct of the scope if called in the direct controller,
// otherwise (in page group controllers for example) it adds a model to the scope.
//
// This method makes the exported fields of the struct available in the page's HTML code.
func (mh *ModelHolder) SetModel(model interface{}) {
	mh.models = append(mh.models, model)

	if mh.inMainCtrl {
		mh.mainIndex = len(mh.models) - 1
	}
}

// SetModelNamed is like SetModel, here we associate the model with a name
func (mh *ModelHolder) SetModelNamed(name string, model interface{}) {
	mh.namedModels[name] = model

	if mh.inMainCtrl {
		mh.mainName = name
	}
}

// Model returns the list of UNNAMED models added to the scope
func (mh *ModelHolder) Models() []interface{} {
	return mh.models
}

// Model returns the main model (the one that is set by the direct page controller)
func (mh *ModelHolder) Model() interface{} {
	if mh.mainName != "" {
		return mh.namedModels[mh.mainName]
	}

	return mh.models[mh.mainIndex]
}

func (mh *ModelHolder) NamedModel(name string) interface{} {
	return mh.namedModels[name]
}

func (pc *Scope) bindModels() (ret []interface{}) {
	ret = []interface{}{}
	ret = append(ret, pc.ModelHolder.Models()...)

	ret = append(ret,
		pc.ModelHolder.namedModels,
		pc.valMap,
		map[string]interface{}{
			"_pageInfo": pc.Page(),
		})

	return
}

// Http is a convenient method which returns the default http client
func (pc *Scope) Http() *http.Client {
	return pc.Services().Http
}
