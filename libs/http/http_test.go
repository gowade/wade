package http

import "testing"

func TestHeader(t *testing.T) {
	req, _ := NewRequest("GET", "/test")
	h := req.Headers
	k := []string{"a", "b"}
	v := []string{"v1", "v2"}
	h.Add(k[0], v[0])
	if h.Get(k[0]) != v[0] {
		t.Fatalf("expected `%v`, got `%v`", v[0], h.Get(k[0]))
	}
	h.Add(k[1], v[0])
	if h.Get(k[1]) != v[0] {
		t.Fatalf("expected `%v`, got `%v`", v[0], h.Get(k[1]))
	}
	h.Add(k[0], v[1])
	if h.Get(k[0]) != v[0] {
		t.Fatalf("expected `%v`, got `%v`", v[0], h.Get(k[0]))
	}
	h.Set(k[0], v[1])
	if h.Get(k[0]) != v[1] {
		t.Fatalf("expected %v, got %v", v[1], h.Get(k[0]))
	}
	h.Del(k[0])
	if h.Get(k[0]) != "" {
		t.Fatalf("expected `%v`, got `%v`", "", h.Get(k[0]))
	}
}

func TestInterceptor(t *testing.T) {
	v := false
	tk, tv := "yes", "here"
	http := NewClient(&stubBackend{testResponse{200, ""}})
	http.AddRequestInterceptor(func(r *Request) {
		r.Headers.Add(tk, tv)
		v = true
	})

	req, _ := http.NewRequest("GET", "/")
	if !v || req.Headers.Get(tk) != tv {
		t.Fatalf("interceptor has not been called.")
	}

	//Test the http API with something like authentication handling
	sb := &stubBackend{testResponse{401, ""}}
	client := NewClient(sb)

	var pendingRequest *Request
	client.AddRequestInterceptor(func(r *Request) {
		pendingRequest = r
	})

	ok := false
	client.AddResponseInterceptor(func(finish chan bool, r *Request) {
		if r.Response.StatusCode == 401 {
			sb.Response(200, "true")
			client.Do(pendingRequest)

			ok = true
			finish <- true
		}
	})

	go func() {
		var b bool
		resp, err := client.GET("/zzz")
		resp.ParseJSON(&b)
		if !ok || err != nil || !b {
			t.Fatalf("Expected %v, got %v", true, false)
		}
	}()
}
