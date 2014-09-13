package wade

import (
	"testing"

	"github.com/phaikawl/wade/dom/goquery"
	"github.com/phaikawl/wade/icommon"
	"github.com/phaikawl/wade/libs/http"
	"github.com/stretchr/testify/require"
)

const (
	Src = `<div>
<winclude src="/a"></winclude>
<winclude src="/b"></winclude>
<div>
	<winclude src="/c"></winclude>
</div>
</div>`

	FailSrc = `<div><winclude src="/kdkfk"></winclude></div>`
	NoSrc   = `<div><winclude></winclude></div>`

	SrcA = `<winclude src="/d"></winclude>`
	SrcB = `b`
	SrcC = `c`
	SrcD = `a`
)

func TestHtmlImport(t *testing.T) {
	mb := http.NewMockBackend(map[string]http.TestResponse{
		"/a": http.FakeOK(SrcA),
		"/b": http.FakeOK(SrcB),
		"/c": http.FakeOK(SrcC),
		"/d": http.FakeOK(SrcD),
	})

	client := http.NewClient(mb)

	root := goquery.GetDom().NewFragment(Src)
	err := htmlImport(client, root, "/")
	require.Equal(t, err, nil)
	require.Equal(t, icommon.RemoveAllSpaces(root.Html()), `ab<div>c</div>`)

	root = root.NewFragment(FailSrc)
	err = htmlImport(client, root, "/")
	require.NotEqual(t, err, nil)

	root = root.NewFragment(NoSrc)
	err = htmlImport(client, root, "/")
	require.NotEqual(t, err, nil)
}
