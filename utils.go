package wade

import neturl "net/url"

// UrlQuery adds query arguments (?arg1=value1&arg2=value2...)
// specified in the given map args to a given url and returns the new result
func UrlQuery(url string, args map[string][]string) string {
	qs := neturl.Values(args).Encode()
	if qs == "" {
		return url
	}

	return url + "?" + qs
}

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
