package components

import (
	"path"

	"github.com/gopherjs/gopherjs/js"
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

type Title struct{}

func (t *Title) BeforeMount() {
	var title string
	children := t.VDOMChildren()
	if len(children) > 0 {
		title = children[0].Text()
	}

	dom.GetDocument().SetTitle(title)
}
