package scope

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type (
	Test struct {
		C9 int
	}
)

func TestScope(t *testing.T) {
	s1 := NewModelScope(Test{C9: 1})
	_, err := s1.Lookup("Nonexistant")
	require.Equal(t, strings.Contains(err.Error(), "Unable to find"), true)
	symbol, err := s1.Lookup("C9")
	require.Equal(t, err, nil)
	v, _ := symbol.Value()
	require.Equal(t, v.Int(), 1)

	s2 := &Scope{[]SymbolTable{NewHelpersSymbolTable(map[string]interface{}{
		"testHelper": func() bool {
			return true
		},
	})}}

	s1.Merge(s2)
	symbol, err = s1.Lookup("testHelper")
	require.Equal(t, err, nil)
	v, _ = symbol.Call([]reflect.Value{}, false)
	require.Equal(t, v.Bool(), true)

	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": true,
		},
		"b": &Test{2},
	}

	st := NewModelScope(m)
	symbol, err = st.Lookup("a.b")
	require.Equal(t, err, nil)
	v, _ = symbol.Value()
	require.Equal(t, v.Interface(), true)

	symbol, err = st.Lookup("b.C9")
	require.Equal(t, err, nil)
	v, _ = symbol.Value()
	require.Equal(t, v.Interface(), 2)
}
