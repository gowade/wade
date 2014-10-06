package wade

import (
	"github.com/phaikawl/wade/custom"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
)

var (
	AppServices GlobalServices
	AppConf     AppConfig
	ClientSide  bool
)

type (
	// AppFunc is the main application func
	AppFunc func(Registration)

	// PageControllerFunc is the function to be run on the load of a page or page scope
	PageControllerFunc func(*Scope) error

	Redirecter interface {
		RedirectToPage(page string, params ...interface{})
		RedirectToUrl(string)
	}

	//GlobalServices is the struct to contain global services
	GlobalServices struct {
		Http           *http.Client
		PageManager    PageManager
		LocalStorage   Storage
		SessionStorage Storage
	}

	// PageManager manages the web page and switching between pages
	PageManager interface {
		Redirecter
		BasePath() string
		CurrentPage() *Scope //Get the current page scope
		FullPath(string) string
		PageUrl(page string, namedParams ...interface{}) (string, error) //Get url of a page with the given namedParams
	}

	// Registration is used in the appFunc before the app is actually started.
	// Register things like pages, custom tags...
	Registration interface {
		RegisterDisplayScopes(pages []PageDesc, pageGroups []PageGroupDesc)
		RegisterCustomTags(...custom.HtmlTag)
		RegisterController(displayScope string, controller PageControllerFunc)
		RegisterNotFoundPage(pageid string)
	}

	// AppConfig is app configurations, used at the start
	AppConfig struct {
		BasePath string
		// The application container, if not specified, it's an element added into <body>
		Container  dom.Selection
		ServerBase string
	}
)
