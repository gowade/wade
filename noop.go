package wade

type (
	NoopHistory struct {
		path  string
		title string
	}
)

func newNoopHistory() *NoopHistory {
	return &NoopHistory{path: "/"}
}

func (h *NoopHistory) ReplaceState(title string, path string) {
	h.path = path
	h.title = title
}
func (h *NoopHistory) PushState(title string, path string) {
	h.path = path
	h.title = title
}
func (h *NoopHistory) OnPopState(fn func()) {}
func (h *NoopHistory) CurrentPath() string {
	return h.path
}
