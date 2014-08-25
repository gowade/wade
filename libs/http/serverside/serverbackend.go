package serverside

import (
	"bufio"
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
	buf := bytes.NewBufferString("")
	b.ClientReq.Write(buf)
	req, _ := http.ReadRequest(bufio.NewReader(buf))
	req.Method = r.Method
	req.URL, _ = url.Parse(r.Url)
	req.Body = ioutil.NopCloser(bytes.NewBufferString(r.Data()))

	resp := httptest.NewRecorder()
	b.Server.ServeHTTP(resp, req)

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
