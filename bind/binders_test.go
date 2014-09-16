package bind

import (
	"testing"

	"github.com/phaikawl/wade/custom"
	"github.com/phaikawl/wade/dom/goquery"
	"github.com/stretchr/testify/require"
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
		custom.BaseProto
		A *aStr
	}

	Nest struct {
		List []string
	}

	nestedModel struct {
		List []Nest
	}
)

func TestEach(t *testing.T) {
	// Test with slice
	b := NewTestBindEngine()
	b.tm.RegisterTags([]custom.HtmlTag{
		custom.HtmlTag{
			Name:       "test",
			Prototype:  &ceModel{},
			Attributes: []string{"A"},
			Html:       `<span #html="A.B"></span>`,
		},
	})

	m1 := &sliceModel{[]*okScope{&okScope{true, "a", &aStr{"a"}}, &okScope{false, "b", &aStr{"b"}}, &okScope{true, "c", &aStr{"c"}}}}
	src := `<wroot>
		<ul>
			<ww #each(key,item)="List">
				<li>#<span #html="|key"></span><span #html="item.A"></span><test @A="item.B"></test></li>
			</ww>
		</ul>
	</wroot>
	`
	root := goquery.GetDom().NewFragment(src)
	elem := root.Find("ul")
	b.Bind(root, m1, false)
	lis := elem.Children().Filter("li").Elements()
	require.Equal(t, lis[0].Text(), "#0aa")
	require.Equal(t, lis[1].Text(), "#1bb")
	require.Equal(t, lis[2].Text(), "#2cc")

	m1.List = m1.List[1:]
	b.Watcher().Digest(&m1.List)
	lis = elem.Children().Filter("li").Elements()
	require.Equal(t, lis[0].Text(), "#0bb")
	require.Equal(t, lis[1].Text(), "#1cc")

	// Test with map
	m2 := &mapModel{map[string]string{
		"0": "a",
		"1": "b",
		"2": "c",
	}}

	src = `<wroot>
		<ul>
			<li #each(key,value)="Map">#<span #html="|key"></span><span #html="|value"></span></li>
		</ul>
	</wroot>
	`

	elem = goquery.GetDom().NewFragment(src).Find("ul")
	b.Bind(elem, m2, false)
	lis = elem.Children().Filter("li").Elements()
	consists := func(txt string) bool {
		for _, li := range lis {
			if li.Text() == txt {
				return true
			}
		}

		return false
	}

	require.Equal(t, consists("#0a"), true)
	require.Equal(t, consists("#1b"), true)
	require.Equal(t, consists("#2c"), true)

	delete(m2.Map, "0")
	m2.Map["1"] = "bb"
	b.Watcher().Digest(&m2.Map)
	lis = elem.Children().Filter("li").Elements()
	require.Equal(t, consists("#0a"), false)
	require.Equal(t, consists("#1bb"), true)
	require.Equal(t, consists("#2c"), true)

	m3 := &nestedModel{[]Nest{Nest{[]string{"a", "b"}}, Nest{[]string{"b"}}}}
	src = `<wroot>
		<ul>
			<ww #each(key,item)="List">
				<li><ww #each(_,subitem)="|item.List"><span #html="|subitem"></span></ww></li>
			</ww>
		</ul>
	</wroot>
	`
	root = goquery.GetDom().NewFragment(src)
	elem = root.Find("ul")
	b.Bind(elem, m3, false)
	lis = elem.Find("li").Elements()
	require.Equal(t, lis[0].Text(), "ab")
	require.Equal(t, lis[1].Text(), "b")
}

func TestIf(t *testing.T) {
	b := NewTestBindEngine()
	s := &okScope{
		Ok: false,
		A:  ":D",
	}
	root := goquery.GetDom().NewFragment(`<wroot>
		<div #if="Ok"><span #html="A"></span></div>
	</wroot>`)
	b.Bind(root, s, false)
	require.Equal(t, root.Find("div").Length(), 0)
	s.Ok = true
	b.Watcher().Digest(&s.Ok)
	require.Equal(t, root.Find("div span").Text(), ":D")
}
