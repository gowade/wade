package http

type testResponse struct {
	StatusCode int
	Data       string
}

type stubBackend struct {
	testResponse
}

func (sb *stubBackend) Do(r *Request) error {
	r.Response = &Response{
		Data:       sb.Data,
		StatusCode: sb.StatusCode,
	}

	return nil
}

func (sb *stubBackend) Response(status int, data string) {
	sb.StatusCode = status
	sb.Data = data
}
