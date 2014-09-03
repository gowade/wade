package wade

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/dom/goquery"
	"github.com/phaikawl/wade/icommon"
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

func (b *NoopBindEngine) Bind(relem dom.Selection, model interface{}, once bool, bindrelem bool) {}

func (b *NoopBindEngine) BindModels(relem dom.Selection, models []interface{}, once bool, bindrelem bool) {
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
	pm := newPageManager(NewNoopHistory("/"),
		AppConfig{StartPage: "pg-home", BasePath: "/web"},
		doc,
		template,
		b)

	container := doc.Find("body").First()

	pm.registerDisplayScopes([]PageDesc{
		MakePage("pg-home", "/home", "Home"),
		MakePage("pg-child-1", "/child/:name", "Child 1"),
		MakePage("pg-child-2", "/child/:name/:gender", "Child 2"),
	}, []PageGroupDesc{MakePageGroup("grp-parent", "pg-child-1", "pg-child-2")})

	mess := make(chan int, 5)

	globalCalled := false

	pm.registerController(GlobalDisplayScope, func(p *PageCtrl) interface{} {
		globalCalled = true
		return Struct1{A: 0}
	})

	pm.registerController("pg-home", func(p *PageCtrl) interface{} {
		mess <- 1
		return Struct2{B: 1}
	})

	pm.prepare()

	require.Equal(t, globalCalled, true)
	require.Equal(t, <-mess, 1)
	require.Equal(t, icommon.RemoveAllSpaces(container.Text()), "Home")

	require.Equal(t, b.models[0].(Struct1).A, 0)
	require.Equal(t, b.models[1].(Struct2).B, 1)

	pm.registerController("grp-parent", func(p *PageCtrl) interface{} {
		return Struct1{A: 2}
	})

	pm.registerController("pg-child-1", func(p *PageCtrl) interface{} {
		return Struct2{B: 3}
	})

	pm.registerController("pg-child-2", func(p *PageCtrl) interface{} {
		return Struct3{C: 4}
	})

	pm.updatePage("/child/vuong", false)
	require.Equal(t, b.models[0].(Struct1).A, 0)
	require.Equal(t, b.models[1].(Struct1).A, 2)
	require.Equal(t, b.models[2].(Struct2).B, 3)
	require.Equal(t, icommon.RemoveAllSpaces(container.Text()), "ParentChild1")

	pm.updatePage("/child/vuong/nam", false)
	require.Equal(t, b.models[0].(Struct1).A, 0)
	require.Equal(t, b.models[1].(Struct1).A, 2)
	require.Equal(t, b.models[2].(Struct3).C, 4)
	require.Equal(t, icommon.RemoveAllSpaces(container.Text()), "ParentChild2")
}
