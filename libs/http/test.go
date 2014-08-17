package http

import (
	urlrouter "github.com/naoina/kocha-urlrouter"
)

type StubBackend struct {
	TestResponse
}

func (sb *StubBackend) Do(r *Request) error {
	r.Response = &Response{
		Data:       sb.Data,
		StatusCode: sb.StatusCode,
	}

	return nil
}

func (sb *StubBackend) Response(status int, data string) {
	sb.StatusCode = status
	sb.Data = data
}

type TestResponse struct {
	StatusCode int
	Data       string
}

type MockBackend struct {
	Router urlrouter.URLRouter
}

func NewMockBackend(handlers map[string]TestResponse) *MockBackend {
	router := urlrouter.NewURLRouter("regexp")
	records := make([]urlrouter.Record, 0)
	for route, handler := range handlers {
		records = append(records, urlrouter.NewRecord(route, handler))
	}

	router.Build(records)

	return &MockBackend{
		Router: router,
	}
}

func (mb *MockBackend) Do(r *Request) error {
	match, _ := mb.Router.Lookup(r.Url)
	tr := match.(TestResponse)
	r.Response = &Response{
		Data:       tr.Data,
		StatusCode: tr.StatusCode,
	}

	return nil
}

func FakeOK(data string) TestResponse {
	return TestResponse{200, data}
}
