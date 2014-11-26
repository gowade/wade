package markman

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/phaikawl/wade/dom/goquery"
	"github.com/phaikawl/wade/libs/http"
	hm "github.com/phaikawl/wade/test/httpmock"
	"github.com/phaikawl/wade/utils"
)

const (
	Src = `<div>
<w-include src="/a"></w-include>
<w-include src="/b"></w-include>
<div>
	<w-include src="/c"></w-include>
</div>
</div>`

	NoSrc = `<div><w-include></w-include></div>`

	SrcA = `<w-include src="/d"></w-include>`
	SrcB = `b`
	SrcC = `c`
	SrcD = `a`
)

type fetcher struct {
	http *http.Client
}

func (f fetcher) FetchFile(src string) (html string, err error) {
	resp, err := f.http.GET(src)
	html = resp.Data
	return
}

func TestHtmlImport(t *testing.T) {
	mb := hm.NewMock(map[string]hm.Responder{
		"/a": hm.NewListResponder([]hm.Responder{hm.NewOKResponse(SrcA), hm.NewOKResponse(SrcB)}),
		"/b": hm.NewOKResponse(SrcB),
		"/c": hm.NewOKResponse(SrcC),
		"/d": hm.NewOKResponse(SrcD),
	})

	client := http.NewClient(mb)

	root := goquery.GetDom().NewFragment(Src)
	fetcher := fetcher{client}
	err := HTMLImports(fetcher, root)
	return
	require.Equal(t, err, nil)
	require.Equal(t, utils.NoSp(root.Text()), `abc`)
	root = goquery.GetDom().NewFragment(Src)

	HTMLImports(fetcher, root)
	require.Equal(t, utils.NoSp(root.Text()), `bbc`)

	root = root.NewFragment(NoSrc)
	err = HTMLImports(fetcher, root)
	require.NotEqual(t, err, nil)
}
