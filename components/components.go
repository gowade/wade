package components

import (
	"path"

	"github.com/gopherjs/gopherjs/js"
	//"github.com/gowade/vdom"
	"github.com/gowade/wade"
	"github.com/gowade/wade/dom"
)

type Link struct {
	Path string
}

func (lnk *Link) OnClick(evt *js.Object) {
	evt.Call("preventDefault")
	wade.App().SetURLPath(lnk.Path)
}

func (lnk *Link) Href() string {
	return path.Join(wade.App().BasePath, lnk.Path)
}

type DocumentTitle struct {
	Text string
}

func (t *DocumentTitle) BeforeMount() {
	dom.GetDocument().SetTitle(t.Text)
}
