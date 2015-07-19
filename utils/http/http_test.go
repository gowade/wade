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
