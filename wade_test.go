package wade

import (
	"testing"

	"github.com/phaikawl/wade/dom/goquery"
	"github.com/phaikawl/wade/icommon"
	"github.com/phaikawl/wade/libs/http"
	hm "github.com/phaikawl/wade/test/httpmock"
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

	NoSrc = `<div><winclude></winclude></div>`

	SrcA = `<winclude src="/d"></winclude>`
	SrcB = `b`
	SrcC = `c`
	SrcD = `a`
)

func TestHtmlImport(t *testing.T) {
	mb := hm.NewMock(map[string]hm.Responder{
		"/a": hm.NewListResponder([]hm.Responder{hm.NewOKResponse(SrcA), hm.NewOKResponse(SrcB)}),
		"/b": hm.NewOKResponse(SrcB),
		"/c": hm.NewOKResponse(SrcC),
		"/d": hm.NewOKResponse(SrcD),
	})

	client := http.NewClient(mb)

	root := goquery.GetDom().NewFragment(Src)
	err := htmlImport(client, root, "/")
	require.Equal(t, err, nil)
	require.Equal(t, icommon.RemoveAllSpaces(root.Text()), `abc`)
	root = goquery.GetDom().NewFragment(Src)
	htmlImport(client, root, "/")
	require.Equal(t, icommon.RemoveAllSpaces(root.Text()), `bbc`)

	root = root.NewFragment(NoSrc)
	err = htmlImport(client, root, "/")
	require.NotEqual(t, err, nil)
}
