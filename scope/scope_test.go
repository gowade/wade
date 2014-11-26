package scope

import (
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
	s1 := NewScope(Test{C9: 1})
	_, err := s1.Lookup("Nonexistant")
	require.Equal(t, strings.Contains(err.Error(), "Unable to find"), true)
	symbol, err := s1.Lookup("C9")
	require.Equal(t, err, nil)
	v, _ := symbol.Value()
	require.Equal(t, v.Int(), 1)

	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": true,
		},
		"b": &Test{2},
	}

	st := NewScope(m)
	symbol, err = st.Lookup("a.b")
	require.Equal(t, err, nil)
	v, _ = symbol.Value()
	require.Equal(t, v.Interface(), true)

	symbol, err = st.Lookup("b.C9")
	require.Equal(t, err, nil)
	v, _ = symbol.Value()
	require.Equal(t, v.Interface(), 2)
}
