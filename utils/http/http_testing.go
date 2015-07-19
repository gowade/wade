package http

type testResponse struct {
	StatusCode int
	Body       string
}

type stubBackend struct {
	testResponse
}

func (sb *stubBackend) Do(r *Request) (*Response, error) {
	return &Response{
		Body:       []byte(sb.Body),
		StatusCode: sb.StatusCode,
	}, nil
}

func (sb *stubBackend) Response(status int, data string) {
	sb.StatusCode = status
	sb.Body = data
}
