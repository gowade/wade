package bind

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
	s1 := newModelScope(Test{C9: 1})
	_, err := s1.lookup("Nonexistant")
	require.Equal(t, strings.Contains(err.Error(), "Unable to find"), true)
	symbol, err := s1.lookup("C9")
	require.Equal(t, err, nil)
	v, _ := symbol.value()
	require.Equal(t, v.Int(), 1)

	s2 := &scope{[]symbolTable{newHelpersSymbolTable(map[string]interface{}{
		"testHelper": func() bool {
			return true
		},
	})}}

	s1.merge(s2)
	symbol, err = s1.lookup("testHelper")
	require.Equal(t, err, nil)
	v, _ = symbol.call([]reflect.Value{})
	require.Equal(t, v.Bool(), true)

	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": true,
		},
		"b": &Test{2},
	}

	st := newModelScope(m)
	symbol, err = st.lookup("a.b")
	require.Equal(t, err, nil)
	v, _ = symbol.value()
	require.Equal(t, v.Interface(), true)

	symbol, err = st.lookup("b.C9")
	require.Equal(t, err, nil)
	v, _ = symbol.value()
	require.Equal(t, v.Interface(), 2)
}
