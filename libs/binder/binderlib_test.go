package binder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCollection(t *testing.T) {
	list := []string{"a", "b", "c"}

	m, err := GetLoopList(list)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range m {
		k, v := item.Key, item.Value
		if k.Int() == 0 {
			require.Equal(t, v.String(), "a")
		}
		if k.Int() == 1 {
			require.Equal(t, v.String(), "b")
		}
		if k.Int() == 2 {
			require.Equal(t, v.String(), "c")
		}
	}

	list = list[1:]
	m, err = GetLoopList(list)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range m {
		k, v := item.Key, item.Value
		if k.Int() == 0 {
			require.Equal(t, v.String(), "b")
		}
		if k.Int() == 1 {
			require.Equal(t, v.String(), "c")
		}
	}

	m, err = GetLoopList(map[string]int{
		"a": 0,
		"b": 1,
		"c": 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, item := range m {
		k, v := item.Key, item.Value
		if k.String() == "a" {
			require.Equal(t, v.Int(), 0)
		}
		if k.String() == "b" {
			require.Equal(t, v.Int(), 1)
		}
		if k.String() == "c" {
			require.Equal(t, v.Int(), 2)
		}
	}
}
