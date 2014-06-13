/* This package is a wade service that provides persistent data storage and retrieving. */
package pdata

import (
	"encoding/json"

	"github.com/gopherjs/gopherjs/js"
)

var (
	gStorage Storage
)

type Storage struct {
	js.Object //localStorage
}

func (stg *Storage) get(key string, outVal interface{}) (ok bool) {
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

func (stg *Storage) GetStr(key string) (v string, ok bool) {
	ok = stg.get(key, &v)
	return
}

func (stg *Storage) GetInt(key string) (v int, ok bool) {
	ok = stg.get(key, &v)
	return
}

func (stg *Storage) GetFloat(key string) (v float64, ok bool) {
	ok = stg.get(key, &v)
	return
}

//Get the stored value with key key and store it in v.
//Typically used for struct values.
func (stg *Storage) GetTo(key string, v interface{}) bool {
	return stg.get(key, v)
}

func (stg *Storage) Set(key string, v interface{}) {
	s, err := json.Marshal(v)
	if err != nil {
		panic(err.Error())
	}
	stg.Object.Call("setItem", key, string(s))
}

func init() {
	gStorage = Storage{js.Global.Get("localStorage")}
}

func Service() *Storage {
	return &gStorage
}
