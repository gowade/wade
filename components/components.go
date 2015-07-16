package components

import (
	"path"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gowade/wade"
	"github.com/gowade/wade/utils/dom"
)

type Link struct {
	wade.Com
	Path string
}

func (lnk *Link) OnClick(evt *js.Object) {
	evt.Call("preventDefault")
	wade.App().SetURLPath(lnk.Path)
}

func (lnk *Link) Href() string {
	return path.Join(wade.App().BasePath, lnk.Path)
}

type Title struct {
	wade.Com
}

func (t *Title) BeforeMount() {
	var title string
	if len(t.Children) > 0 {
		title = t.Children[0].Text()
	}

	dom.GetDocument().SetTitle(title)
}
