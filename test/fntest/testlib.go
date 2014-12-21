package fntest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"

	//"golang.org/x/net/html"

	"github.com/phaikawl/wade/app"
	//"github.com/phaikawl/wade/core"

	//"github.com/phaikawl/wade/dom"

	"github.com/phaikawl/wade/libs/http"
	"github.com/phaikawl/wade/platform/serverside"
)

type (
	TestApp struct {
		View *TestView
		*app.Application
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

	found := app.PageMgr.GoToUrl(path)
	if !found {
		panic(fmt.Errorf(`Page not found for "%v"`, path))
	}
}

func (app *TestApp) Start(appMain app.Main) error {
	app.started = true
	return app.Application.Start(appMain)
}

func (app *TestApp) Render() {
	app.Application.Render()
	app.View.rsessId++
}

func NewDummyTestApp(initialPath string, httpMock http.Backend) *TestApp {
	return NewTestApp(app.Config{}, initialPath, "", httpMock)
}

func NewTestApp(conf app.Config,
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
	app.SetApp(wapp)

	return &TestApp{
		Application: wapp,
		View:        &TestView{wapp.Document(), 1},
	}
}

type dummyMain struct{}

func (dm dummyMain) Main(app *app.Application) {
}

func StartDummyApp(httpMock http.Backend) (*TestApp, error) {
	app := NewTestApp(app.Config{}, "/", "", httpMock)
	err := app.Start(dummyMain{})
	return app, err
}
