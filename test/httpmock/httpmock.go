package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	gohttp "net/http"
	"os"
	"path"

	urlrouter "github.com/naoina/kocha-urlrouter"
	"github.com/phaikawl/wade/libs/http"
)

type Responder interface {
	Response(c *Context) Response
}

type Context struct {
	NamedParams *http.NamedParams
	Request     *http.Request
}

func (c *Context) Json(data interface{}) string {
	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return string(bytes[:])
}

type Response struct {
	StatusCode int
	Data       string
}

func (r Response) Response(c *Context) Response {
	return r
}

type ListResponder struct {
	List  []Responder
	Index int
}

func NewListResponder(list []Responder) *ListResponder {
	return &ListResponder{list, 0}
}

func (l *ListResponder) Response(c *Context) (r Response) {
	r = l.List[l.Index].Response(c)
	l.Index++
	return
}

type FuncResponder func(c *Context) Response

func (rf FuncResponder) Response(c *Context) Response {
	return rf(c)
}

type FileResponder struct {
	ParamName string
	Directory string
}

// NewFileResponder returns a Responder that serves static file.
// It receives the file path from the given named parameter paramName
// and opens file in the given base directory
func NewFileResponder(paramName, directory string) *FileResponder {
	_, err := os.Stat(directory)
	if err != nil {
		panic(err)
	}

	return &FileResponder{paramName, directory}
}

func (fr *FileResponder) Response(c *Context) Response {
	filepath, ok := c.NamedParams.Get(fr.ParamName)
	if !ok {
		panic(fmt.Errorf(`There must be a "%v" parameter (usually we use "*%v") in the url pattern for a FileResponder`, fr.ParamName, fr.ParamName))
	}

	file, err := os.Open(path.Join(fr.Directory, filepath))
	if err != nil {
		return Response{
			StatusCode: 404,
		}
	}

	bytes, _ := ioutil.ReadAll(file)
	return Response{
		StatusCode: 200,
		Data:       string(bytes[:]),
	}
}

type HttpMock struct {
	Router            urlrouter.URLRouter
	BlockChan         chan bool
	ResponseCount     int
	WaitResponseCount int
}

func NewMock(handlers map[string]Responder) *HttpMock {
	router := urlrouter.NewURLRouter("regexp")
	records := make([]urlrouter.Record, 0)
	for route, handler := range handlers {
		records = append(records, urlrouter.NewRecord(route, handler))
	}

	router.Build(records)

	return &HttpMock{
		Router:    router,
		BlockChan: make(chan bool, 1),
	}
}

func (mb *HttpMock) Wait(operation func(), responseCount int) {
	mb.ResponseCount = 0
	mb.WaitResponseCount = responseCount
	operation()
	<-mb.BlockChan
	mb.WaitResponseCount = 0
}

func (mb *HttpMock) responseFinish() {
	if mb.WaitResponseCount > 0 {
		mb.ResponseCount++
		if mb.ResponseCount >= mb.WaitResponseCount {
			mb.BlockChan <- false
		}
	}
}

func (mb *HttpMock) Do(r *http.Request) (err error) {
	match, params := mb.Router.Lookup(r.URL.Path)
	if match == nil {
		mb.responseFinish()
		panic(fmt.Errorf(`404 no handler found for path "%v".`, r.URL.Path))
	}

	tr := match.(Responder).Response(&Context{http.NewNamedParams(params), r})

	r.Response = &http.Response{
		Data:       tr.Data,
		StatusCode: tr.StatusCode,
		Status:     gohttp.StatusText(tr.StatusCode),
	}

	mb.responseFinish()

	return nil
}

func NewOKResponse(data string) Response {
	return Response{200, data}
}

func NewJsonResponse(data interface{}) Response {
	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return Response{
		Data:       string(bytes[:]),
		StatusCode: 200,
	}
}
