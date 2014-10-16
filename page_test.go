package wade

import (
	"testing"

	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/dom/goquery"
	"github.com/phaikawl/wade/icommon"
	"github.com/stretchr/testify/require"
)

type (
	NoopBindEngine struct {
		models []interface{}
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
		</head>
		<body>
		</body>
	</html>
	`)

	template := goquery.GetDom().NewFragment(`
	<div>
		<div w-belong="pg-home">Home</div>
		<div w-belong="grp-parent">
			<div>Parent</div>
			<div w-belong="pg-child-1">
				Child 1
			</div>
			<div w-belong="pg-child-2">
				Child 2
			</div>
		</div>
	</div>
	`)

	b := &NoopBindEngine{}
	pm := newPageManager(&Application{Config: AppConfig{BasePath: "/web"}}, NewNoopHistory("/"),
		doc,
		template,
		b)

	container := doc.Find("body").First()

	pm.router.Handle("/", Redirecter{"/home"}).
		Handle("/home", Page{Id: "pg-home"}).
		Handle("/child/:name", Page{Id: "pg-child-1"}).
		Handle("/child/:name/:gender", Page{Id: "pg-child-2"})

	pm.registerPageGroup("grp-parent", []string{"pg-child-1", "pg-child-2"})

	mess := make(chan int, 5)

	globalCalled := false

	pm.registerController(GlobalDisplayScope, func(s *PageScope) (err error) {
		globalCalled = true
		s.SetModelNamed("global", Struct1{
			A: 0,
		})

		return
	})

	pm.registerController("pg-home", func(s *PageScope) (err error) {
		mess <- 1
		s.SetModel(Struct2{B: 1})

		return
	})

	pm.prepare()

	require.Equal(t, globalCalled, true)
	require.Equal(t, <-mess, 1)
	require.Equal(t, icommon.RemoveAllSpaces(container.Text()), "Home")

	s := bind.ScopeFromModels(b.models)
	v, _ := s.LookupValue("global.A")
	require.Equal(t, v.(int), 0)

	v, _ = s.LookupValue("B")
	require.Equal(t, v.(int), 1)

	pm.registerController("grp-parent", func(s *PageScope) (err error) {
		s.SetModelNamed("parent", Struct1{A: 2})
		return
	})

	pm.registerController("pg-child-1", func(s *PageScope) (err error) {
		s.SetModel(Struct2{B: 3})
		return
	})

	pm.registerController("pg-child-2", func(s *PageScope) (err error) {
		s.SetModel(Struct3{C: 4})
		return
	})

	pm.updateUrl("/child/vuong", false, false)

	s = bind.ScopeFromModels(b.models)
	v, _ = s.LookupValue("parent.A")
	require.Equal(t, v.(int), 2)

	v, _ = s.LookupValue("B")
	require.Equal(t, v.(int), 3)

	require.Equal(t, icommon.RemoveAllSpaces(container.Text()), "ParentChild1")

	pm.updateUrl("/child/vuong/nam", false, false)
	s = bind.ScopeFromModels(b.models)
	v, _ = s.LookupValue("C")
	require.Equal(t, v.(int), 4)

	require.Equal(t, icommon.RemoveAllSpaces(container.Text()), "ParentChild2")
}
