package wade

import "github.com/phaikawl/wade/libs/http"

var (
	ClientSide bool
)

type (
	// AppFunc is the main application func
	AppFunc func(*Application)

	// PageControllerFunc is the function to be run on the load of a page or page scope
	PageControllerFunc func(*PageScope) error

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
		CurrentPage() *PageScope //Get the current page scope
		FullPath(string) string
		PageUrl(page string, namedParams ...interface{}) (string, error) //Get url of a page with the given namedParams
	}

	// AppConfig is app configurations, used at the start
	AppConfig struct {
		BasePath   string
		ServerBase string
	}
)
