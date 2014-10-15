package wade

import (
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
)

var (
	ClientSide bool
)

type (
	// AppFunc is the main application func
	AppFunc func(*Application)

	// PageControllerFunc is the function to be run on the load of a page or page scope
	PageControllerFunc func(*Scope) error

	//AppServices is the struct to contain basic services
	AppServices struct {
		Http           *http.Client
		PageManager    PageManager
		LocalStorage   Storage
		SessionStorage Storage
	}

	// PageManager manages the web page and switching between pages
	PageManager interface {
		GoToPage(page string, params ...interface{}) (found bool)
		GoToUrl(string) (found bool)
		BasePath() string
		CurrentPage() *Scope //Get the current page scope
		FullPath(string) string
		PageUrl(page string, namedParams ...interface{}) (string, error) //Get url of a page with the given namedParams
	}

	// AppConfig is app configurations, used at the start
	AppConfig struct {
		BasePath string
		// The application container, if not specified, it's an element added into <body>
		Container  dom.Selection
		ServerBase string
	}
)
