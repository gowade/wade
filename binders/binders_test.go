package binders

import (
	"testing"

	"github.com/gopherjs/gopherjs/js"
	"github.com/stretchr/testify/require"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/dom/gonet"
	"github.com/phaikawl/wade/utils"
)

type (
	sliceModel struct {
		List  []*okScope
		List2 []string
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

	evtModel struct {
		ok bool
	}

	dummyEvent struct{}
)

func (e dummyEvent) StopPropagation() {
}

func (e dummyEvent) Target() dom.Selection {
	return nil
}

func (e dummyEvent) PreventDefault() {}

func (e dummyEvent) Type() string {
	return ""
}

func (e dummyEvent) Js() js.Object {
	return nil
}

var (
	gq = gonet.Dom{}
	b  = core.NewBindEngine(nil, make(map[string]interface{}))
)

func init() {
	Install(b)
}

func (m *evtModel) Ok(evt *dom.Event) {
	_, m.ok = (*evt).(dummyEvent)
}
func TestEvent(t *testing.T) {
	m := &evtModel{false}
	vRoot := gq.NewFragment(`<div><a #on(click)="@Ok($event)"></a></div>`).ToVNode().Ptr()
	a := vRoot.ChildElems()[0]
	b.Bind(vRoot, m)
	v, ok := a.Attr("onclick")
	fn := v.(func(dom.Event))
	require.Equal(t, ok, true)
	fn(dummyEvent{})
	require.Equal(t, m.ok, true)
}

func TestEach(t *testing.T) {
	// Test with slice
	b.ComponentManager().Register(core.ComponentView{
		Name:      "Test",
		Prototype: &ceModel{},
		Template:  gq.NewFragment(`<span>{{ A.B }}z</span>`).ToVNode(),
	})

	m1 := &sliceModel{[]*okScope{
		&okScope{true, "a", &aStr{"a"}},
		&okScope{false, "b", &aStr{"b"}},
		&okScope{true, "c", &aStr{"c"}},
	}, []string{"1", "2"}}

	src := `<ul>
			<div !group #range(key,item)="List">
				<li>
					#<span>{{ key }}</span>
					<span>{{ item.A }}</span>
					<test @A="item.B"></test>
				</li>
			</div>
			<div !group #range(_,item)="List2"><li>{{ item }}</li></div>
		</ul>`

	rRoot := gq.NewFragment(src)
	vRoot := rRoot.ToVNode().Ptr()
	b.Bind(vRoot, m1)

	vRoot.Update()
	//n := vRoot.Children[1].Children[1]
	//fmt.Printf("%v %v %v", n.Type, n.Data, n.Attrs())
	rRoot.Render(vRoot)

	list := rRoot.Children().Filter("li").Elements()
	require.Equal(t, utils.NoSp(list[0].Text()), "#0aaz")
	require.Equal(t, utils.NoSp(list[1].Text()), "#1bbz")
	require.Equal(t, utils.NoSp(list[2].Text()), "#2ccz")

	require.Equal(t, utils.NoSp(list[3].Text()), "1")
	require.Equal(t, utils.NoSp(list[4].Text()), "2")

	m1.List = m1.List[1:]

	vRoot.Update()
	rRoot.Render(vRoot)

	list = rRoot.Children().Filter("li").Elements()
	require.Equal(t, utils.NoSp(list[0].Text()), "#0bbz")
	require.Equal(t, utils.NoSp(list[1].Text()), "#1ccz")

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
	vRoot = rRoot.ToVNode().Ptr()
	b.Bind(vRoot, m3)

	vRoot.Update()
	rRoot.Render(vRoot)

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
	vRoot := rRoot.ToVNode().Ptr()
	b.Bind(vRoot, s)

	vRoot.Update()
	rRoot.Render(vRoot)

	require.Equal(t, rRoot.Children().Length(), 1)

	require.Equal(t, rRoot.Find("span").Text(), "ZZZ")

	s.Ok = true

	vRoot.Update()

	rRoot.Render(vRoot)
	require.Equal(t, rRoot.Find("span").Text(), ":D")
}
