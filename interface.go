package wade

import (
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
)

var (
	AppServices GlobalServices
)

type AppFunc func(Registration)

// PageControllerFunc is the function to be run on the load of a specific page.
// It returns a model to be used in bindings of the elements in the page.
type PageControllerFunc func(ThisPage) interface{}

type Redirecter interface {
	RedirectToPage(page string, params ...interface{})
	RedirectToUrl(string)
}

type ThisPage interface {
	Services() GlobalServices
	Manager() PageManager
	Info() PageInfo
	FormatTitle(params ...interface{})
	GetParam(param string, dest interface{}) error
	RegisterHelper(name string, fn interface{})
	Redirecter
}

type GlobalServices struct {
	Http           *http.Client
	PageManager    PageManager
	LocalStorage   Storage
	SessionStorage Storage
}

type PageManager interface {
	Redirecter
	BasePath() string
	CurrentPage() ThisPage
	Fullpath(string) string
	PageUrl(page string, params ...interface{}) (string, error)
}

type NeedsInit interface {
	Init(services GlobalServices)
}

type Registration interface {
	RegisterDisplayScopes(pages []PageDesc, pageGroups []PageGroupDesc)
	RegisterCustomTags(...CustomTag)
	RegisterController(displayScope string, controller PageControllerFunc)
	ModuleInit(...NeedsInit)
}

type AppConfig struct {
	StartPage  string
	BasePath   string
	Container  dom.Selection
	ServerBase string
	ServerMode bool
}
