package core

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)

func toString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func jsGetType(obj js.Object) string {
	return js.Global.Get("Object").Get("prototype").Get("toString").Call("call", obj).Str()
}
