package wade

import (
	"github.com/phaikawl/wade/libs/http"
	"github.com/phaikawl/wade/libs/pdata"
)

type EventHandler func()

// PageControllerFunc is the function to be run on the load of a specific page.
// It returns a model to be used in bindings of the elements in the page.
type PageControllerFunc func(ThisPage) interface{}

type Redirecter interface {
	RedirectToPage(page string, params ...interface{})
	RedirectToUrl(string)
}

type ThisPage interface {
	Services() AppServices
	Manager() PageManager
	Info() PageInfo
	FormatTitle(params ...interface{})
	GetParam(param string, dest interface{}) error
	RegisterHelper(name string, fn interface{})
	Redirecter
}

type AppServices struct {
	Http           *http.Client
	LocalStorage   pdata.Storage
	SessionStorage pdata.Storage
}

type PageManager interface {
	Redirecter
	BasePath() string
	CurrentPage() ThisPage
	Url(string) string
	PageUrl(page string, params ...interface{}) (string, error)
	SetOutputContainer(elementId string)
}

type AppEnv struct {
	Services    AppServices
	PageManager PageManager
}

type NeedsInit interface {
	Init(AppEnv)
}

type Registration interface {
	RegisterDisplayScopes(map[string]DisplayScope)
	RegisterCustomTags(src string, models map[string]interface{})
	RegisterController(displayScope string, controller PageControllerFunc)
	ModuleInit(...NeedsInit)
}
