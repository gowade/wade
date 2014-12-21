package page

type noopHistory struct {
	path  string
	title string
}

func NewNoopHistory(path string) *noopHistory {
	return &noopHistory{path: path}
}

func (h *noopHistory) ReplaceState(title string, path string) {
	h.path = path
	h.title = title
}

func (h *noopHistory) PushState(title string, path string) {
	h.path = path
	h.title = title
}

func (h *noopHistory) OnPopState(fn func()) {}

func (h *noopHistory) CurrentPath() string {
	return h.path
}

func (h *noopHistory) RedirectTo(url string) {
}
