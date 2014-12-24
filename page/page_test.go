package page

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom/gonet"
	"github.com/phaikawl/wade/markman"
	"github.com/phaikawl/wade/scope"
	"github.com/phaikawl/wade/utils"
)

type (
	Struct1 struct {
		A int
	}

	Struct2 struct {
		B int
	}

	mockFetcher struct{}

	mockBindEngine struct {
		models []interface{}
	}
)

func (f mockFetcher) FetchFile(file string) (string, error) {
	return "", nil
}

func (b *mockBindEngine) Bind(_ *core.VNode, models ...interface{}) {
	b.models = models
}

func TestPageUrl(t *testing.T) {
	pm := &PageManager{}
	pm.displayScopes = make(map[string]displayScope)
	pm.router = newRouter()
	route := "/:testparam/:testparam2/*testparam3"
	r := pm.RouteMgr()
	r.Handle(route,
		Page{
			Id: "test",
		})

	u := pm.PageUrl("test", 12, "abc", "some.go")
	expected := "/12/abc/some.go"

	require.Equal(t, u, expected)

	var err error
	u, err = pm.pageUrl("test", []interface{}{12, "abc"})
	if err == nil {
		t.Fatalf("It should have raised an error for not having enough parameters.")
	}

	u, err = pm.pageUrl("test", []interface{}{12, "abc", "zz", 22})
	if err == nil {
		t.Fatalf("It should have raised an error for having too many parameters.")
	}
}

func TestPageManager(t *testing.T) {
	var (
		root = gonet.GetDom().NewDocument(`<html><head></head>
	<body !appview>
		<div !belong="pg-home">Home</div>
		<div !belong="grp-parent">
			<div>Parent</div>
			<div !belong="pg-child-1">
				Child 1
			</div>
			<div !belong="pg-child-2">
				Child 2
			</div>
		</div>
	</body></html>`)

		err          error
		mess         = make(chan int, 5)
		globalCalled = false
		markman      = markman.New(root, mockFetcher{})
		container    = markman.Container()
		b            = &mockBindEngine{}
		pm           = NewPageManager("/", NewNoopHistory("/"), markman, b)
	)

	err = markman.LoadView()
	require.Equal(t, err, nil)

	r := pm.RouteMgr()
	r.Handle("/", Redirecter{"/home"})
	r.Handle("/home", Page{
		Id: "pg-home",
		Controller: func(ctx Context) Scope {
			mess <- 1
			return Scope{
				"home": "home",
			}
		},
	})

	r.Handle("/child/:name", Page{
		Id: "pg-child-1",
		Controller: func(ctx Context) Scope {
			return Scope{
				"b": Struct2{B: 3},
			}
		},
	})

	r.Handle("/child/:name/:gender", Page{
		Id: "pg-child-2",
		Controller: func(ctx Context) Scope {
			return Scope{
				"p": Struct1{A: 4},
			}
		},
	})

	pm.AddPageGroup(PageGroup{
		Id:       "grp-parent",
		Children: []string{"pg-child-1", "pg-child-2"},
		Controller: func(ctx Context) Scope {
			return Scope{
				"p": Struct1{A: 2},
			}
		},
	})

	GlobalDisplayScope.AddController(func(ctx Context) Scope {
		globalCalled = true
		return Scope{
			"global": Struct1{
				A: 0,
			},
		}
	})

	pm.Start()

	require.Equal(t, globalCalled, true)
	require.Equal(t, <-mess, 1)
	require.Equal(t, utils.NoSp(container.Text()), "Home")

	s := scope.NewScope(b.models...)
	v, err := s.LookupValue("global.A")
	if err != nil {
		panic(err)
	}
	require.Equal(t, v.(int), 0)

	v, err = s.LookupValue("home")
	if err != nil {
		panic(err)
	}

	require.Equal(t, v.(string), "home")

	pm.updateUrl("/child/vuong", false, false)

	//fmt.Printf("%+v", b.models)
	s = scope.NewScope(b.models...)
	v, err = s.LookupValue("p.A")
	if err != nil {
		panic(err)
	}
	require.Equal(t, v.(int), 2)

	v, err = s.LookupValue("b.B")
	if err != nil {
		panic(err)
	}
	require.Equal(t, v.(int), 3)

	require.Equal(t, utils.NoSp(container.Text()), "ParentChild1")

	pm.updateUrl("/child/vuong/nam", false, false)

	s = scope.NewScope(b.models...)

	v, err = s.LookupValue("p.A")
	if err != nil {
		panic(err)
	}
	require.Equal(t, v.(int), 4)

	require.Equal(t, utils.NoSp(container.Text()), "ParentChild2")
}
