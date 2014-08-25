package clientside

import "github.com/gopherjs/gopherjs/js"

type (
	History struct {
		js.Object
	}
)

func (h History) ReplaceState(title, path string) {
	h.Object.Call("replaceState", nil, title, path)
}

func (h History) PushState(title, path string) {
	h.Object.Call("pushState", nil, title, path)
}

func (h History) CurrentPath() string {
	location := h.Get("location")
	if location.IsNull() || location.IsUndefined() {
		location = js.Global.Get("document").Get("location")
	}

	return location.Get("pathname").Str()
}

func (h History) OnPopState(fn func()) {
	gJQ(js.Global.Get("window")).On("popstate", fn)
}

func (h History) RedirectTo(url string) {
	js.Global.Get("window").Set("location", url)
}
