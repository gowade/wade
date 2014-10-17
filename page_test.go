package wade

import (
	"testing"

	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/dom/goquery"
	"github.com/phaikawl/wade/icommon"
	hm "github.com/phaikawl/wade/test/httpmock"
	"github.com/stretchr/testify/require"
)

type (
	NoopBindEngine struct {
		models []interface{}
	}

	NoopJsBackend struct {
		bind.BasicWatchBackend
		JsHistory History
	}

	Struct1 struct {
		A int
	}

	Struct2 struct {
		B int
	}

	Struct3 struct {
		C int
	}
)

func (b *NoopBindEngine) Watcher() *bind.Watcher {
	return bind.NewWatcher(bind.BasicWatchBackend{})
}

func (b *NoopBindEngine) BindModels(root dom.Selection, models []interface{}, once bool) {
	b.models = models
}

func (b *NoopBindEngine) RegisterInternalHelpers(pm bind.PageManager) {
}

func (b *NoopJsBackend) CheckJsDep(symbol string) bool {
	return true
}

func (b *NoopJsBackend) History() History {
	return b.JsHistory
}

func (b *NoopJsBackend) WebStorages() (Storage, Storage) {
	s := Storage{}
	return s, s
}

func TestPageUrl(t *testing.T) {
	pm := &pageManager{}
	pm.displayScopes = make(map[string]displayScope)
	pm.router = newRouter(pm)
	route := "/:testparam/:testparam2/*testparam3"
	pm.router.Handle(route,
		Page{
			Id: "test",
		})

	var u string
	var err error
	u, err = pm.PageUrl("test", 12, "abc", "some.go")
	expected := "/12/abc/some.go"
	if err != nil {
		t.Fatalf(err.Error())
	}

	require.Equal(t, u, expected)

	u, err = pm.PageUrl("test", 12, "abc")
	if err == nil {
		t.Fatalf("It should have raised an error for not having enough parameters.")
	}

	u, err = pm.PageUrl("test", 12, "abc", "zz", 22)
	if err == nil {
		t.Fatalf("It should have raised an error for having too many parameters.")
	}
}

func TestPageManager(t *testing.T) {
	doc := goquery.GetDom().NewDocument(`
	<html>
		<head>
			<script type="text/wadin">
				<winclude src="/pages.html"></winclude>
			</script>
		</head>
		<body>
			<div w-app-container></div>
		</body>
	</html>
	`)

	server := hm.NewMock(map[string]hm.Responder{
		"/pages.html": hm.NewOKResponse(`<div>
		<div w-belong="pg-home">Home</div>
		<div w-belong="grp-parent">
			<div>Parent</div>
			<winclude w-belong="pg-child-1" src="/child1.html"></winclude>
			<winclude w-belong="pg-child-2" src="/child2.html"></winclude>
		</div>
	</div>`),
		"/child1.html": hm.NewOKResponse(`Child 1`),
		"/child2.html": hm.NewOKResponse(`Child 2`),
	})

	globalCalled := false
	mess := make(chan int, 5)
	container := doc.Find("body").First()
	b := &NoopBindEngine{}

	app, err := newApp(AppConfig{BasePath: "/"}, func(app *Application) {
		app.Router.Handle("/", Redirecter{"/home"}).
			Handle("/home", Page{Id: "pg-home"}).
			Handle("/child/:name", Page{Id: "pg-child-1"}).
			Handle("/child/:name/:gender", Page{Id: "pg-child-2"})

		app.Register.PageGroup("grp-parent", []string{"pg-child-1", "pg-child-2"})

		app.Register.Controller(GlobalDisplayScope, func(s *PageScope) (err error) {
			globalCalled = true
			s.SetModelNamed("global", Struct1{
				A: 0,
			})

			return
		})

		app.Register.Controller("pg-home", func(s *PageScope) (err error) {
			mess <- 1
			s.SetModel(Struct2{B: 1})

			return
		})

		app.Register.Controller("grp-parent", func(s *PageScope) (err error) {
			s.SetModelNamed("parent", Struct1{A: 2})
			return
		})

		app.Register.Controller("pg-child-1", func(s *PageScope) (err error) {
			s.SetModel(Struct2{B: 3})
			return
		})

		app.Register.Controller("pg-child-2", func(s *PageScope) (err error) {
			s.SetModel(Struct3{C: 4})
			return
		})
	},

		RenderBackend{
			JsBackend: &NoopJsBackend{
				BasicWatchBackend: bind.BasicWatchBackend{},
				JsHistory:         NewNoopHistory("/"),
			},
			HttpBackend: server,
			Document:    doc,
		},

		b,
	)

	if err != nil {
		t.Fatal(err)
	}

	app.Start()

	require.Equal(t, globalCalled, true)
	require.Equal(t, <-mess, 1)
	require.Equal(t, icommon.RemoveAllSpaces(container.Text()), "Home")

	s := bind.ScopeFromModels(b.models)
	v, _ := s.LookupValue("global.A")
	require.Equal(t, v.(int), 0)

	v, _ = s.LookupValue("B")
	require.Equal(t, v.(int), 1)

	app.CurrentPage().GoToUrl("/child/vuong")

	require.Equal(t, icommon.RemoveAllSpaces(container.Text()), "ParentChild1")

	s = bind.ScopeFromModels(b.models)
	v, err = s.LookupValue("parent.A")
	require.Equal(t, v.(int), 2)

	v, _ = s.LookupValue("B")
	require.Equal(t, v.(int), 3)

	app.CurrentPage().GoToUrl("/child/vuong/nam")

	s = bind.ScopeFromModels(b.models)
	v, _ = s.LookupValue("C")
	require.Equal(t, v.(int), 4)

	require.Equal(t, icommon.RemoveAllSpaces(container.Text()), "ParentChild2")
}
