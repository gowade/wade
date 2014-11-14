package core

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRefComp(t *testing.T) {
	v1, v2 := reflect.ValueOf(1), reflect.ValueOf(2)
	comp, ok := compareRefl(v1, v2)
	require.Equal(t, comp, -1)
	require.Equal(t, ok, true)

	v1, v2 = reflect.ValueOf(1.2), reflect.ValueOf(1.1)
	comp, _ = compareRefl(v1, v2)
	require.Equal(t, comp, 1)
	require.Equal(t, ok, true)

	a := 2
	b := 2
	v3, v4 := reflect.ValueOf(&a), reflect.ValueOf(&b)
	comp, _ = compareRefl(v3, v4)
	require.Equal(t, comp, -1)
	require.Equal(t, ok, true)
}

type repr struct {
	slice []int
}

func (r *repr) Remove(idx int) {
	r.slice = append(r.slice[:idx], r.slice[idx+1:]...)
}

func (r *repr) Add(idx int, value reflect.Value) {
	r.slice = append(r.slice[:idx], append([]int{int(value.Int())}, r.slice[idx:]...)...)
}

func TestSliceChange(t *testing.T) {
	a := []int{1, 2, 3, 4, 6}
	b := []int{1, 2, 3, 4}

	r := &repr{a}
	performChange(r, reflect.ValueOf(a), reflect.ValueOf(b))
	require.Equal(t, reflect.DeepEqual(r.slice, b), true)

	a = []int{1, 2, 3, 4}
	b = []int{1, 2, 3, 5, 4}

	r = &repr{a}
	performChange(r, reflect.ValueOf(a), reflect.ValueOf(b))
	require.Equal(t, reflect.DeepEqual(r.slice, b), true)
}
