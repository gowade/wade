package http

import "testing"

func TestHeader(t *testing.T) {
	h := NewRequest(MethodGet, "/test").Headers
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
	httpSrv := Service()
	httpSrv.AddHttpInterceptor(func(r *Request) {
		r.Headers.Add(tk, tv)
		v = true
	})
	req := httpSrv.NewRequest(MethodGet, "/")
	if !v || req.Headers.Get(tk) != tv {
		t.Fatalf("interceptor has not been called.")
	}
}
