package http

import (
	gourl "net/url"
	"strings"
	"time"
)

var (
	driver Driver
)

func SetDriver(drv Driver) {
	driver = drv
}

func Do(req *Request) (*Response, error) {
	return driver.Do(req)
}

type (
	Header map[string][]string

	Request struct {
		Body            []byte
		Header          Header
		Method          string
		URL             *gourl.URL
		Timeout         time.Duration
		WithCredentials bool
	}

	Driver interface {
		Do(*Request) (*Response, error)
	}

	Response struct {
		Body       []byte
		Status     string
		StatusCode int
		Header     Header
	}
)

// Add adds the key, value pair to the header.
// It appends to any existing values associated with key.
func (h Header) Add(key, value string) {
	if _, ok := h[key]; !ok {
		h[key] = make([]string, 0)
	}
	h[key] = append(h[key], value)
}

// Set sets the header entries associated with key to
// the single element value.  It replaces any existing
// values associated with key.
func (h Header) Set(key, value string) {
	if _, ok := h[key]; ok {
		h[key] = nil
	}
	h[key] = make([]string, 0)
	h[key] = append(h[key], value)
}

// Get gets the first value associated with the given key.
// If there are no values associated with the key, Get returns "".
func (h Header) Get(key string) string {
	if v, ok := h[key]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
}

// Del deletes the values associated with key.
func (h Header) Del(key string) {
	if _, ok := h[key]; ok {
		delete(h, key)
	}
}

func (h Header) String() (s string) {
	for key, values := range h {
		s += key + ": " + strings.Join(values, ";") + "\n"
	}

	return
}

func NewRequest(method string, url string, body []byte) (*Request, error) {
	u, err := gourl.Parse(url)
	if err != nil {
		return nil, err
	}

	return &Request{
		Method: method,
		Header: make(map[string][]string),
		URL:    u,
		Body:   body,
	}, nil
}
