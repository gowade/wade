package jsdom

import (
	"github.com/gopherjs/gopherjs/js"
)

type FormEl struct{ Node }

func (e FormEl) IsValid() bool {
	if e.Get("checkValidity") != js.Undefined {
		return e.Call("checkValidity").Bool()
	}

	return true
}

type InputEl struct{ Node }

func (e InputEl) Checked() bool {
	return e.Get("checked").Bool()
}

func (e InputEl) SetChecked(checked bool) {
	e.Set("checked", checked)
}

func (e InputEl) Value() string {
	return e.Get("value").String()
}

func (e InputEl) JS() *js.Object {
	return e.Object
}

func (e InputEl) SetValue(value string) {
	e.Set("value", value)
}
