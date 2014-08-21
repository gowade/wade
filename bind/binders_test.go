package bind

import (
	"testing"

	"github.com/phaikawl/wade/dom/goquery"
	"github.com/stretchr/testify/require"
)

type (
	sliceModel struct {
		List []string
	}

	mapModel struct {
		Map map[string]string
	}
)

func TestEach(t *testing.T) {
	// Test with slice
	wc, _, b := initTestBind()
	m1 := &sliceModel{[]string{"a", "b", "c"}}
	src := `
	<ul>
		<li bind-each="List -> key, item">#<span bind-html="key"></span><span bind-html="item"></span></li>
	</ul>
	`
	elem := goquery.GetDom().NewFragment(src)
	b.Bind(elem, m1, false, false)
	lis := elem.Children().Filter("li").Elements()
	require.Equal(t, lis[0].Text(), "#0a")
	require.Equal(t, lis[1].Text(), "#1b")
	require.Equal(t, lis[2].Text(), "#2c")

	m1.List = m1.List[1:]
	wc.watches[0]()
	lis = elem.Children().Filter("li").Elements()
	require.Equal(t, lis[0].Text(), "#0b")
	require.Equal(t, lis[1].Text(), "#1c")

	// Test with map
	m2 := &mapModel{map[string]string{
		"0": "a",
		"1": "b",
		"2": "c",
	}}

	src = `
	<ul>
		<li bind-each="Map -> key, value">#<span bind-html="key"></span><span bind-html="value"></span></li>
	</ul>
	`

	elem = goquery.GetDom().NewFragment(src)
	b.Bind(elem, m2, false, false)
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
	wc.watches[1]()
	lis = elem.Children().Filter("li").Elements()
	require.Equal(t, consists("#0a"), false)
	require.Equal(t, consists("#1bb"), true)
	require.Equal(t, consists("#2c"), true)
}
