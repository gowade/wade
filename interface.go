package wade

import "github.com/phaikawl/wade/libs/http"

var (
	ClientSide bool
)

type (
	Map map[string]interface{}

	// PageControllerFunc is the functiong to be run on the load of a page or page scope
	ControllerFunc func(Context) Map

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
		Context() Context // Get the Context for current page
		CurrentPage() Page
		FullPath(string) string
		PageUrl(page string, namedParams ...interface{}) string //Get url of a page with the given namedParams
	}

	Storage interface {
		Get(key string) interface{}
		Set(key string, v interface{})
		Delete(key string)
	}
)
