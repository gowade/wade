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
	Index = `<html><head></head>
	<body>
		<div !appview="/index">
		</div>
	</body>`

	Src = `	<w-include src="/a"></w-include>
			<w-include src="/b"></w-include>
			<div>
				<w-include src="/c"></w-include>
			</div>`

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

func TestMarkMgr(t *testing.T) {
	mb := hm.NewMock(map[string]hm.Responder{
		"/a":     hm.NewListResponder([]hm.Responder{hm.NewOKResponse(SrcA), hm.NewOKResponse(SrcB)}),
		"/b":     hm.NewOKResponse(SrcB),
		"/c":     hm.NewOKResponse(SrcC),
		"/d":     hm.NewOKResponse(SrcD),
		"/index": hm.NewOKResponse(Src),
	})

	client := http.NewClient(mb)
	fetcher := fetcher{client}

	root := goquery.GetDom().NewDocument(Index)
	markman := New(root, fetcher)
	err := markman.LoadView()
	require.Equal(t, err, nil)
	markman.Render()
	require.Equal(t, utils.NoSp(root.Text()), `abc`)
}
