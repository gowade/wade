package core

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom/goquery"
	"github.com/phaikawl/wade/utils"
)

type (
	sliceModel struct {
		List []*okScope
	}

	mapModel struct {
		Map map[string]string
	}

	okScope struct {
		Ok bool
		A  string
		B  *aStr
	}

	aStr struct {
		B string
	}

	ceModel struct {
		core.BaseProto
		A *aStr
	}

	Nest struct {
		List []string
	}

	nestedModel struct {
		List []Nest
	}
)

var (
	gq = goquery.Dom{}
	b  = core.NewTestBindEngine()
)

func init() {
	for name, binder := range Binders {
		b.RegisterBinder(name, binder)
	}
}

func TestEach(t *testing.T) {
	// Test with slice
	b.ComponentManager().Register(core.ComponentView{
		Name:      "test",
		Prototype: &ceModel{},
		Template:  gq.NewFragment(`<span>{{ A.B }}</span>`).ToVNode(),
	})

	m1 := &sliceModel{[]*okScope{
		&okScope{true, "a", &aStr{"a"}},
		&okScope{false, "b", &aStr{"b"}},
		&okScope{true, "c", &aStr{"c"}}},
	}

	src := `<ul>
			<div !group #range(key,item)="List">
				<li>
					#<span>{{ key }}</span>
					<span>{{ item.A }}</span>
					<test @A="item.B"></test>
				</li>
			</div>
		</ul>`

	rRoot := gq.NewFragment(src)
	vRoot := core.NodeRoot(rRoot.ToVNode())
	b.Bind(vRoot, m1)

	vRoot.Update()
	//n := vRoot.Children[1].Children[1]
	//fmt.Printf("%v %v %v", n.Type, n.Data, n.Attrs())
	rRoot.Render(*vRoot)

	list := rRoot.Children().Filter("li").Elements()
	require.Equal(t, utils.NoSp(list[0].Text()), "#0aa")
	require.Equal(t, utils.NoSp(list[1].Text()), "#1bb")
	require.Equal(t, utils.NoSp(list[2].Text()), "#2cc")

	m1.List = m1.List[1:]

	vRoot.Update()
	rRoot.Render(*vRoot)

	list = rRoot.Children().Filter("li").Elements()
	require.Equal(t, utils.NoSp(list[0].Text()), "#0bb")
	require.Equal(t, utils.NoSp(list[1].Text()), "#1cc")

	m3 := &nestedModel{[]Nest{Nest{[]string{"a", "b"}}, Nest{[]string{"b"}}}}
	src = `<ul>
			<div !group #range(key,item)="List">
				<li>
					<div !group #range(_,subitem)="item.List">
						{{ subitem }}
					</div>
				</li>
			</div>
		</ul>
	</ul>`

	rRoot = gq.NewFragment(src)
	vRoot = core.NodeRoot(rRoot.ToVNode())
	b.Bind(vRoot, m3)

	vRoot.Update()
	rRoot.Render(*vRoot)

	list = rRoot.Children().Filter("li").Elements()
	require.Equal(t, utils.NoSp(list[0].Text()), "ab")
	require.Equal(t, utils.NoSp(list[1].Text()), "b")
}

func TestIf(t *testing.T) {
	s := &okScope{
		Ok: false,
		A:  ":D",
	}
	rRoot := gq.NewFragment(`<div><span #if="Ok">{{ A }}</span><span #ifn="Ok">ZZZ</span></div>`)
	vRoot := core.NodeRoot(rRoot.ToVNode())
	b.Bind(vRoot, s)

	vRoot.Update()
	rRoot.Render(*vRoot)

	require.Equal(t, rRoot.Children().Length(), 1)

	require.Equal(t, rRoot.Find("span").Text(), "ZZZ")

	s.Ok = true

	vRoot.Update()
	rRoot.Render(*vRoot)

	require.Equal(t, rRoot.Find("span").Text(), ":D")
}
