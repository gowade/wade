/* This package is a wade service that provides persistent data storage and retrieving. */
package pdata

import (
	"encoding/json"
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)

var (
	gLocalStorage   Storage = storage{js.Global.Get("localStorage")}
	gSessionStorage Storage = storage{js.Global.Get("sessionStorage")}
)

type StorageType int

const (
	LocalStorage   StorageType = 0
	SessionStorage StorageType = 1
)

type Storage interface {
	GetBool(key string) (v bool, ok bool)
	GetStr(key string) (v string, ok bool)
	GetInt(key string) (v int, ok bool)
	GetTo(key string, v interface{}) bool
	Set(key string, v interface{})
}

type storage struct {
	js.Object
}

func (stg storage) get(key string, outVal interface{}) (ok bool) {
	jsv := stg.Object.Call("getItem", key)
	ok = !jsv.IsNull() && !jsv.IsUndefined()
	if ok {
		gv := jsv.Str()
		err := json.Unmarshal([]byte(gv), &outVal)
		if err != nil {
			panic(err.Error())
		}
	}
	return
}

func (stg storage) GetBool(key string) (v bool, ok bool) {
	ok = stg.get(key, &v)
	return
}

func (stg storage) GetStr(key string) (v string, ok bool) {
	ok = stg.get(key, &v)
	return
}

func (stg storage) GetInt(key string) (v int, ok bool) {
	ok = stg.get(key, &v)
	return
}

//Get the stored value with key key and store it in v.
//Typically used for struct values.
func (stg storage) GetTo(key string, v interface{}) bool {
	return stg.get(key, v)
}

func (stg storage) Set(key string, v interface{}) {
	s, err := json.Marshal(v)
	if err != nil {
		panic(err.Error())
	}
	stg.Object.Call("setItem", key, string(s))
}

func init() {
}

func Service(storageType StorageType) Storage {
	switch storageType {
	case LocalStorage:
		return gLocalStorage
	case SessionStorage:
		return gSessionStorage
	}

	panic(fmt.Sprintf(`Invalid storage type "%v".`, storageType))
	return gLocalStorage
}
