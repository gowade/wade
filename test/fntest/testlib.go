package fntest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"

	//"golang.org/x/net/html"

	"github.com/phaikawl/wade/rt"
	//"github.com/phaikawl/wade/core"

	"github.com/phaikawl/wade/page"

	"github.com/phaikawl/wade/libs/http"
	"github.com/phaikawl/wade/platform/serverside"
)

type (
	TestApp struct {
		View *TestView
		*rt.Application
		started bool
	}
)

func trim(str string) string {
	re := regexp.MustCompile(" +")
	return re.ReplaceAllString(str, " ")
}

// GoTo navigates the app to the given path
func (app *TestApp) GoTo(path string) {
	if !app.started {
		panic(fmt.Errorf("Application has not been started."))
	}

	app.PageMgr.GoToUrl(path)
}

func (app *TestApp) Start(appMain rt.Main) error {
	app.started = true
	return app.Application.Start(appMain)
}

func NewDummyTestApp(initialPath string, httpMock http.Backend) *TestApp {
	return NewTestApp(rt.Config{}, initialPath, "", httpMock)
}

func NewTestApp(conf rt.Config,
	initialPath string,
	indexFile string,
	httpMock http.Backend) *TestApp {

	var document io.Reader
	if indexFile != "" {
		var err error
		document, err = os.Open(indexFile)
		if err != nil {
			panic(err)
		}
	} else {
		document = bytes.NewReader([]byte(`<html>
			<head></head>
			<body !appview>
			</body>
		</html>`))
	}

	wapp := serverside.NewApp(conf, document, initialPath, httpMock)
	rt.SetApp(wapp)

	return &TestApp{
		Application: wapp,
		View:        &TestView{wapp.Document()},
	}
}

type dummyMain struct{}

func (dm dummyMain) Main() {}

func (dm dummyMain) Setup(r page.Router) {}

func StartDummyApp(httpMock http.Backend) (*TestApp, error) {
	app := NewTestApp(rt.Config{}, "/", "", httpMock)
	err := app.Start(dummyMain{})
	return app, err
}
