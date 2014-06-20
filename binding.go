package wade

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	jq "github.com/gopherjs/jquery"
)

type UpdateDomFn func(elem jq.JQuery, value interface{}, arg []string)
type ModelUpdateFn func(value string)
type WatchDomFn func(elem jq.JQuery, updateFn ModelUpdateFn)

type DomBinder struct {
	update UpdateDomFn
	watch  WatchDomFn
	bind   UpdateDomFn
}

type binding struct {
	domBinders map[string]DomBinder
}

func newBindEngine() *binding {
	return &binding{
		domBinders: map[string]DomBinder{
			"value": DomBinder{
				update: valueUpdateFn,
				watch:  valueWatchFn,
				bind:   nil,
			},
			"html": DomBinder{
				update: htmlUpdateFn,
				watch:  nil,
				bind:   nil,
			},
			"on": DomBinder{
				update: nil,
				watch:  nil,
				bind:   eventBindFn,
			},
		},
	}
}

func getReflectField(o reflect.Value, field string) (reflect.Value, error) {
	if o.Kind() == reflect.Ptr {
		o = o.Elem()
	}

	var rv reflect.Value
	switch o.Kind() {
	case reflect.Struct:
		rv = o.FieldByName(field)
		if !rv.IsValid() {
			rv = o.Addr().MethodByName(field)
		}
	case reflect.Map:
		rv = o.MapIndex(reflect.ValueOf(field))
	default:
		return rv, fmt.Errorf("Unhandled type for accessing %v.", field)
	}

	if !rv.IsValid() {
		return rv, fmt.Errorf("Unable to access %v, field not available.", field)
	}

	//if !rv.CanSet() {
	//	panic("Unaddressable")
	//}

	return rv, nil
}

func (b *binding) bind(elem jq.JQuery, model interface{}) {
	elem.Find("*").Each(func(i int, elem jq.JQuery) {
		if elem.Length == 0 {
			panic("Incorrect element for bind.")
		}

		attrs := elem.Get(0).Get("attributes")
		for i := 0; i < attrs.Length(); i++ {
			attr := attrs.Index(i)
			name := attr.Get("name").Str()
			if strings.HasPrefix(name, BindPrefix) {
				parts := strings.Split(name, "-")
				if len(parts) <= 1 {
					panic(`Illegal "bind-".`)
				}
				if binder, ok := b.domBinders[parts[1]]; ok {
					args := make([]string, 0)
					if len(parts) >= 2 {
						for _, part := range parts[2:] {
							args = append(args, part)
						}
					}
					flist := strings.Split(attr.Get("value").Str(), ".")
					vals := make([]reflect.Value, len(flist)+1)
					o := reflect.ValueOf(model)
					vals[0] = o
					var err error
					for i, field := range flist {
						o, err = getReflectField(o, field)
						if err != nil {
							panic(err.Error())
						}
						vals[i+1] = o
					}

					fmodel := vals[len(vals)-1]

					(func(args []string) {
						if binder.bind != nil {
							binder.bind(elem, fmodel.Interface(), args)
						}
						//use watchjs to watch for changes to the model
						if binder.update != nil {
							js.Global.Call("watch",
								vals[len(vals)-2].Interface(),
								flist[len(flist)-1],
								func(prop string, action string, newVal interface{}, oldVal js.Object) {
									binder.update(elem, newVal, args)
								})
						}
					})(args)

					if binder.watch != nil {
						binder.watch(elem, func(newVal string) {
							//println(newVal)
							if !fmodel.CanSet() {
								panic("Cannot set field.")
							}
							fmodel.Set(reflect.ValueOf(newVal))
						})
					}
				} else {
					panic(fmt.Sprintf(`Dom binder "%v" does not exist.`, binder))
				}
			}
		}
	})
}
