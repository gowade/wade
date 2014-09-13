package menu

import (
	"testing"

	"github.com/phaikawl/wade/bind"
	"github.com/phaikawl/wade/dom/goquery"
	"github.com/stretchr/testify/require"
)

type (
	Scope struct {
		Choice string
	}
)

func TestSwitchMenu(t *testing.T) {
	b := bind.NewTestBindEngine()
	b.TagManager().RegisterTags(HtmlTags())
	scope := &Scope{
		Choice: "a",
	}

	root := goquery.GetDom().NewFragment(`
	<wroot>
		<switchmenu @Current="Choice">
			<ul>
				<li case="a"></li>
				<li case="b"></li>
				<li case="c"></li>
			</ul>
		</switchmenu>
	</wroot>
	`)

	b.Bind(root, scope, false)
	lis := root.Find("ul").Children().Elements()
	require.Equal(t, lis[0].HasClass("active"), true)
	require.Equal(t, lis[1].HasClass("active"), false)
	require.Equal(t, lis[2].HasClass("active"), false)

	scope.Choice = "b"
	b.Watcher().Apply()
	require.Equal(t, lis[0].HasClass("active"), false)
	require.Equal(t, lis[1].HasClass("active"), true)
	require.Equal(t, lis[2].HasClass("active"), false)

	scope.Choice = "kkf"
	b.Watcher().Apply()
	require.Equal(t, lis[0].HasClass("active"), false)
	require.Equal(t, lis[1].HasClass("active"), false)
	require.Equal(t, lis[2].HasClass("active"), false)
}
