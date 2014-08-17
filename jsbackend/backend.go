package jsbackend

import "github.com/gopherjs/gopherjs/js"

var (
	gGlobal  js.Object = js.Global
	gBackend *BackendImp
)

func Get() *BackendImp {
	if gBackend == nil {
		gBackend = &BackendImp{
			History: History{js.Global.Get("history")},
		}
	}
	return gBackend
}

type BackendImp struct {
	History History
}

func (b *BackendImp) CheckJsDep(symbol string) bool {
	if gGlobal.Get(symbol).IsUndefined() {
		return false
	}

	return true
}
