package http

import (
	"fmt"
	"reflect"

	urlrouter "github.com/naoina/kocha-urlrouter"
)

// NamedParams holds the values of named parameters for a route
type NamedParams struct {
	m map[string]string
}

func NewNamedParams(params []urlrouter.Param) (np *NamedParams) {
	np = &NamedParams{make(map[string]string)}
	for _, param := range params {
		np.m[param.Name] = param.Value
	}

	return
}

// GetParam use fmt.Sscan to scan the value of the given named parameter to a dest.
func (np *NamedParams) ScanTo(dest interface{}, param string) (err error) {
	v, ok := np.m[param]
	if !ok {
		err = fmt.Errorf("No such parameter %v.", param)
		return
	}

	if reflect.TypeOf(dest).Kind() != reflect.Ptr {
		return fmt.Errorf("The dest for saving the parameter value must be a pointer so that its value could be modified.")
	}
	_, err = fmt.Sscan(v, dest)
	return
}

// Get returns the string value of the given named parameter
func (np *NamedParams) Get(param string) (value string, ok bool) {
	value, ok = np.m[param]
	return
}
