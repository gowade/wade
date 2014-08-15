package clientside

import (
	"strings"

	"github.com/phaikawl/wade/libs/http"
	"honnef.co/go/js/xhr"
)

type XhrBackend struct {
}

type XhrResponseHeaders struct {
	request *xhr.Request
}

func (rh XhrResponseHeaders) String() string {
	return rh.request.ResponseHeaders()
}

func (rh XhrResponseHeaders) Get(key string) string {
	return rh.request.ResponseHeader(key)
}

func (b XhrBackend) Do(r *http.Request) (err error) {
	req := xhr.NewRequest(r.Method, r.Url)
	req.ResponseType = r.ResponseType
	req.Timeout = r.Timeout
	req.WithCredentials = r.WithCredentials
	for k, values := range r.Headers {
		req.SetRequestHeader(k, strings.Join(values, ","))
	}
	err = req.Send(r.Data())

	r.Response = &http.Response{
		RawData:    req.Response.Interface(),
		Data:       req.ResponseText,
		StatusCode: req.Status,
		Status:     req.StatusText,
		Type:       req.ResponseType,
		Headers:    XhrResponseHeaders{req},
	}

	return
}
