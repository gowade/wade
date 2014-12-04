package menu

import (
	"testing"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom/goquery"
	"github.com/stretchr/testify/require"
)

type (
	Scope struct {
		Choice string
	}
)

func TestSwitchMenu(t *testing.T) {
	b := core.NewBindEngine(nil, map[string]interface{}{})
	b.ComponentManager().Register(Components()...)

	scope := &Scope{
		Choice: "a",
	}

	root := goquery.GetDom().NewFragment(`
	<div>
		<w-switch-menu @Current="Choice">
			<ul>
				<li case="a"></li>
				<li case="b"></li>
				<li case="c"></li>
			</ul>
		</w-switch-menu>
	</div>
	`)

	vroot := root.ToVNode()
	b.Bind(&vroot, scope)

	vroot.Update()

	root.Render(vroot)

	lis := root.Find("ul").Children().Elements()
	require.Equal(t, lis[0].HasClass("active"), true)
	require.Equal(t, lis[1].HasClass("active"), false)
	require.Equal(t, lis[2].HasClass("active"), false)

	scope.Choice = "b"
	vroot.Update()
	root.Render(vroot)
	lis = root.Find("ul").Children().Elements()
	require.Equal(t, lis[0].HasClass("active"), false)
	require.Equal(t, lis[1].HasClass("active"), true)
	require.Equal(t, lis[2].HasClass("active"), false)

	scope.Choice = "kkf"
	vroot.Update()
	root.Render(vroot)
	lis = root.Find("ul").Children().Elements()
	require.Equal(t, lis[0].HasClass("active"), false)
	require.Equal(t, lis[1].HasClass("active"), false)
	require.Equal(t, lis[2].HasClass("active"), false)
}
