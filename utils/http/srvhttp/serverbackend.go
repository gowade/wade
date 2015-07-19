package serverside

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	wadehttp "github.com/gowade/wade/utils/http"
)

type ServerBackend struct {
	Server    http.Handler
	ClientReq *http.Request
}

func (b *ServerBackend) Do(wr *wadehttp.Request) (*wadehttp.Response, error) {
	buf := bytes.NewBufferString("")
	b.ClientReq.Write(buf)

	req, err := http.ReadRequest(bufio.NewReader(buf))
	if err != nil {
		return nil, err
	}

	req.Method = wr.Method
	req.URL = wr.URL
	req.Body = ioutil.NopCloser(bytes.NewBuffer(wr.Body))

	resp := httptest.NewRecorder()
	b.Server.ServeHTTP(resp, req)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &wadehttp.Response{
		Body:       data,
		StatusCode: resp.Code,
		Status:     fmt.Sprint(resp.Code),
		Header:     wadehttp.Header(resp.Header()),
	}, nil
}
