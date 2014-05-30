package wade

import "github.com/gopherjs/gopherjs/js"

type Wade struct {
	js.Object
}

func (w *Wade) RegisterElement(tagName string, model interface{}) {
	w.Call("register", tagName, model)
}

type PageHandler func() interface{}

func (w *Wade) RegisterPageHandler(handlerName string, fn PageHandler) {
	w.Call("registerPageHandler", handlerName, fn)
}

func (w *Wade) Start() {
	w.Call("start")
}

func WadeUp(jsVarName string) *Wade {
	wd := js.Global.Get(jsVarName)
	if wd.IsUndefined() || wd.Get("sign").Str() != "1'M_7763_W4D3,_817C76!" {
		panic("Wrong js var, it should be a valid Wade object.")
	}

	return &Wade{
		Object: wd,
	}
}
