package clientside

import (
	"strings"

	"github.com/gowade/wade/utils/http"
	"honnef.co/go/js/xhr"
)

func init() {
	http.SetDriver(XhrBackend{})
}

type XhrBackend struct {
}

func headers(resp *xhr.Request) http.Header {
	header := make(http.Header)
	str := resp.ResponseHeaders()
	kvlist := strings.Split(str, "\u000d\u000a")
	for _, kv := range kvlist {
		pos := strings.Index(kv, "\u003a\u0020")
		header.Add(kv[:pos], kv[pos+2:])
	}

	return header
}

func (b XhrBackend) Do(r *http.Request) (*http.Response, error) {
	req := xhr.NewRequest(r.Method, r.URL.String())
	req.ResponseType = "text"
	req.Timeout = int(r.Timeout.Seconds())
	req.WithCredentials = r.WithCredentials

	for k, values := range r.Header {
		req.SetRequestHeader(k, strings.Join(values, ","))
	}

	err := req.Send(r.Body)
	if err != nil {
		return nil, err
	}

	return &http.Response{
		Body:       []byte(req.ResponseText),
		StatusCode: req.Status,
		Status:     req.StatusText,
		Header:     headers(req),
	}, nil
}
