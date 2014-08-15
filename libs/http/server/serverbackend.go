package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"

	wadehttp "github.com/phaikawl/wade/libs/http"
)

type ServerBackend struct {
	Server    http.Handler
	ClientReq *http.Request
}

type Headers struct {
	http.Header
}

func (h Headers) String() string {
	w := bytes.NewBufferString("")
	h.Header.Write(w)
	return w.String()
}

func (b ServerBackend) Do(r *wadehttp.Request) error {
	b.ClientReq.Method = r.Method
	b.ClientReq.URL, _ = url.Parse(r.Url)
	b.ClientReq.Body = ioutil.NopCloser(bytes.NewBufferString(r.Data()))

	resp := httptest.NewRecorder()
	b.Server.ServeHTTP(resp, b.ClientReq)

	dbytes, _ := ioutil.ReadAll(resp.Body)
	data := string(dbytes)
	r.Response = &wadehttp.Response{
		RawData:    data,
		Data:       data,
		StatusCode: resp.Code,
		Status:     fmt.Sprintf("%v", resp.Code),
		Type:       "text",
		Headers:    Headers{resp.Header()},
	}

	return nil
}
