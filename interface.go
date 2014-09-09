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
	//interface to enforce that every page model must embed *BaseScope
	ScopeModel interface {
		needsToEmbedBaseScope()
	}

	AppFunc func(Registration)

	// PageControllerFunc is the function to be run on the load of a specific page.
	// It returns a model to be used in bindings of the elements in the page.
	PageControllerFunc func(*BaseScope) ScopeModel

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
		CurrentPage() *BaseScope
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
