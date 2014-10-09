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
	pm := pageManager{}
	pm.displayScopes = make(map[string]displayScope)
	route := "/:testparam/:testparam2/*testparam3"
	pm.registerDisplayScopes([]PageDesc{
		MakePage("test", route, ""),
	}, nil)

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

	pm.registerDisplayScopes([]PageDesc{
		MakePage("pg-home", "/", "Home"),
		MakePage("pg-child-1", "/child/:name", "Child 1"),
		MakePage("pg-child-2", "/child/:name/:gender", "Child 2"),
	}, []PageGroupDesc{
		MakePageGroup("grp-parent", []string{"pg-child-1", "pg-child-2"}),
	})

	mess := make(chan int, 5)

	globalCalled := false

	pm.registerController(GlobalDisplayScope, func(s *Scope) (err error) {
		globalCalled = true
		s.AddModel(Struct1{
			A: 0,
		})

		return
	})

	pm.registerController("pg-home", func(s *Scope) (err error) {
		mess <- 1
		s.AddModel(Struct2{B: 1})

		return
	})

	pm.prepare()

	require.Equal(t, globalCalled, true)
	require.Equal(t, <-mess, 1)
	require.Equal(t, icommon.RemoveAllSpaces(container.Text()), "Home")

	require.Equal(t, b.models[0].(Struct1).A, 0)
	require.Equal(t, b.models[1].(Struct2).B, 1)

	pm.registerController("grp-parent", func(s *Scope) (err error) {
		s.AddModel(Struct1{A: 2})
		return
	})

	pm.registerController("pg-child-1", func(s *Scope) (err error) {
		s.AddModel(Struct2{B: 3})
		return
	})

	pm.registerController("pg-child-2", func(s *Scope) (err error) {
		s.AddModel(Struct3{C: 4})
		return
	})

	pm.updateUrl("/child/vuong", false, false)
	require.Equal(t, b.models[0].(Struct1).A, 0)
	require.Equal(t, b.models[1].(Struct1).A, 2)
	require.Equal(t, b.models[2].(Struct2).B, 3)
	require.Equal(t, icommon.RemoveAllSpaces(container.Text()), "ParentChild1")

	pm.updateUrl("/child/vuong/nam", false, false)
	require.Equal(t, b.models[0].(Struct1).A, 0)
	require.Equal(t, b.models[1].(Struct1).A, 2)
	require.Equal(t, b.models[2].(Struct3).C, 4)
	require.Equal(t, icommon.RemoveAllSpaces(container.Text()), "ParentChild2")
}
