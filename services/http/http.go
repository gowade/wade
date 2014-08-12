package http

import (
	"encoding/json"
	"net/url"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

var (
	gService HttpService
)

type Response struct {
	data       string
	status     int
	textStatus string
}

func NewResponse(data string, xhr js.Object) *Response {
	return &Response{data, xhr.Get("status").Int(), xhr.Get("textStatus").Str()}
}

func (r *Response) Data() string {
	return r.data
}

func (r *Response) Status() int {
	return r.status
}

func (r *Response) TextStatus() string {
	return r.textStatus
}

func (r *Response) DecodeDataTo(dest interface{}) error {
	if r.data == "" {
		panic("No data received.")
	}
	err := json.Unmarshal([]byte(r.data), dest)
	if err != nil {
		panic(err.Error())
	}
	return err
}

func (r *Response) Bool() (b bool) {
	r.DecodeDataTo(&b)
	return
}

type Deferred struct {
	jquery.Deferred
}

type HttpDoneHandler func(*Response)

func (d Deferred) Done(fn HttpDoneHandler) Deferred {
	d.Deferred.Done(func(data string, _ string, jqxhr js.Object) {
		fn(NewResponse(data, jqxhr))
	})
	return d
}

func (d Deferred) Fail(fn HttpDoneHandler) Deferred {
	d.Deferred.Fail(func(jqxhr js.Object, textStatus string, errorThrown js.Object) {
		fn(NewResponse("", jqxhr))
	})
	return d
}

func (d Deferred) Always(fn HttpDoneHandler) Deferred {
	d.Deferred.Always(func(jqxhr js.Object, textStatus string, errorThrown js.Object) {
		fn(NewResponse("", jqxhr))
	})
	return d
}

type HttpMethod string

const (
	MethodGet  HttpMethod = "GET"
	MethodPost HttpMethod = "POST"
	MethodPut  HttpMethod = "PUT"
)

type HttpHeader map[string][]string

// Add adds the key, value pair to the header.
// It appends to any existing values associated with key.
func (h HttpHeader) Add(key, value string) {
	if _, ok := h[key]; !ok {
		h[key] = make([]string, 0)
	}
	h[key] = append(h[key], value)
}

// Set sets the header entries associated with key to
// the single element value.  It replaces any existing
// values associated with key.
func (h HttpHeader) Set(key, value string) {
	if _, ok := h[key]; ok {
		h[key] = nil
	}
	h[key] = make([]string, 0)
	h[key] = append(h[key], value)
}

// Get gets the first value associated with the given key.
// If there are no values associated with the key, Get returns "".
func (h HttpHeader) Get(key string) string {
	if v, ok := h[key]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
}

// Del deletes the values associated with key.
func (h HttpHeader) Del(key string) {
	if _, ok := h[key]; ok {
		delete(h, key)
	}
}

type Request struct {
	Headers HttpHeader
	Method  HttpMethod
	data    []byte
	Url     *url.URL
}

func NewRequest(method HttpMethod, reqUrl string) *Request {
	u, err := url.Parse(reqUrl)
	if err != nil {
		panic(err.Error())
	}
	return &Request{
		Method:  method,
		Headers: make(map[string][]string),
		Url:     u,
	}
}

func (r *Request) SetData(d interface{}) {
	var err error
	r.data, err = json.Marshal(d)
	if err != nil {
		panic(err.Error())
	}
}

func (r *Request) makeJqConfig() map[string]interface{} {
	desturl := r.Url.String()
	m := map[string]interface{}{
		"type":        string(r.Method),
		"url":         desturl,
		"dataType":    "text",
		"processData": false,
		"headers":     r.Headers,
	}
	if len(r.data) != 0 {
		m["data"] = r.data
	}

	return m
}

// Do does an asynchronous http request, yet the API is blocking, just like Go's http
func (r *Request) Do() *Response {
	ch := make(chan *Response, 1)
	cb := func(r *Response) {
		go func() {
			ch <- r
		}()
	}
	Deferred{jquery.Ajax(r.makeJqConfig())}.Done(cb).Always(cb)
	resp := <-ch
	gService.TriggerResponseInterceptors(resp)
	return resp
}

// DoSync does a synchronous http request and directly returns a response.
// This doesn't apply interceptors to the response.
// This method will freeze everything even in a goroutine, so it is only
// suitable for tasks like app initialization. Please use Do() instead for
// the vast majority of cases.
func (r *Request) DoSync() (resp *Response) {
	conf := r.makeJqConfig()
	conf["async"] = false
	def := Deferred{jquery.Ajax(conf)}
	setResp := func(r *Response) {
		resp = r
	}
	def.Done(setResp)
	def.Fail(setResp)
	return
}

type RequestInterceptor func(*Request)
type ResponseInterceptor func(chan bool, *Response)

type HttpService struct {
	reqInts  []RequestInterceptor
	respInts []ResponseInterceptor
}

func (s *HttpService) AddRequestInterceptor(hi RequestInterceptor) {
	s.reqInts = append(s.reqInts, hi)
}

func (s *HttpService) AddResponseInterceptor(hi ResponseInterceptor) {
	s.respInts = append(s.respInts, hi)
}

func (s *HttpService) NewRequest(method HttpMethod, reqUrl string) *Request {
	request := NewRequest(method, reqUrl)
	for _, intrFn := range s.reqInts {
		intrFn(request)
	}

	return request
}

func (s *HttpService) TriggerResponseInterceptors(r *Response) {
	finishChannel := make(chan bool, 1)
	for _, intrFn := range s.respInts {
		intrFn(finishChannel, r)
		<-finishChannel
	}
}

func Service() *HttpService {
	return &gService
}

func init() {
	gService = HttpService{make([]RequestInterceptor, 0), make([]ResponseInterceptor, 0)}
}
