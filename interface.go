package wade

import (
	"github.com/phaikawl/wade/custom"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
)

var (
	AppServices GlobalServices
)

type (
	AppFunc func(Registration)

	// PageControllerFunc is the function to be run on the load of a specific page.
	PageControllerFunc func(*Scope) error

	Redirecter interface {
		RedirectToPage(page string, params ...interface{})
		RedirectToUrl(string)
	}

	GlobalServices struct {
		Http           *http.Client
		PageManager    PageManager
		LocalStorage   Storage
		SessionStorage Storage
	}

	PageManager interface {
		Redirecter
		BasePath() string
		CurrentPage() *Scope
		Fullpath(string) string
		PageUrl(page string, params ...interface{}) (string, error)
	}

	NeedsInit interface {
		Init(services GlobalServices)
	}

	Registration interface {
		RegisterDisplayScopes(pages []PageDesc, pageGroups []PageGroupDesc)
		RegisterCustomTags(...custom.HtmlTag)
		RegisterController(displayScope string, controller PageControllerFunc)
		RegisterNotFoundPage(pageid string)
		ModuleInit(...NeedsInit)
	}

	AppConfig struct {
		StartPage  string
		BasePath   string
		Container  dom.Selection
		ServerBase string
		ServerMode bool
	}
)
