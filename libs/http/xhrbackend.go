package http

import (
	"strings"

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

func (b XhrBackend) Do(r *Request) (err error) {
	req := xhr.NewRequest(r.Method, r.Url)
	req.ResponseType = r.ResponseType
	req.Timeout = r.Timeout
	req.WithCredentials = r.WithCredentials
	for k, values := range r.Headers {
		req.SetRequestHeader(k, strings.Join(values, ","))
	}
	err = req.Send(r.data)

	r.Response = &Response{
		RawData:    req.Response,
		TextData:   req.ResponseText,
		Status:     req.Status,
		TextStatus: req.StatusText,
		Type:       req.ResponseType,
		Headers:    XhrResponseHeaders{req},
	}

	return
}
