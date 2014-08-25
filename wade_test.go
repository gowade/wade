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
<wimport src="/a"></wimport>
<wimport src="/b"></wimport>
<div>
	<wimport src="/c"></wimport>
</div>
</div>`

	FailSrc = `<div><wimport src="/kdkfk"></wimport></div>`
	NoSrc   = `<div><wimport></wimport></div>`

	SrcA = `<wimport src="/d"></wimport>`
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
