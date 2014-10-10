package bind

import "reflect"

func compareRefl(ap, bp reflect.Value) (comp int, ok bool) {
	ok = true

	switch ap.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Chan, reflect.Func:
		app, bpp := ap.Pointer(), bp.Pointer()
		switch {
		case app == bpp:
			comp = 0
		case app < bpp:
			comp = -1
		default:
			comp = 1
		}
		return
	}

	ai, b := ap.Interface(), bp.Interface()
	switch a := ai.(type) {
	case bool:
		bi := b.(bool)
		switch {
		case a == bi:
			comp = 0
		case a == false && bi == true:
			comp = -1
		default:
			comp = 1
		}
	case string:
		bi := b.(string)
		switch {
		case a == bi:
			comp = 0
		case a < bi:
			comp = -1
		default:
			comp = 1
		}
	case int:
		bi := b.(int)
		switch {
		case a == bi:
			comp = 0
		case a < bi:
			comp = -1
		default:
			comp = 1
		}
	case int32:
		bi := b.(int32)
		switch {
		case a == bi:
			comp = 0
		case a < bi:
			comp = -1
		default:
			comp = 1
		}
	case int64:
		bi := b.(int64)
		switch {
		case a == bi:
			comp = 0
		case a < bi:
			comp = -1
		default:
			comp = 1
		}
	case uint:
		bi := b.(uint)
		switch {
		case a == bi:
			comp = 0
		case a < bi:
			comp = -1
		default:
			comp = 1
		}
	case uint32:
		bi := b.(uint32)
		switch {
		case a == bi:
			comp = 0
		case a < bi:
			comp = -1
		default:
			comp = 1
		}
	case uint64:
		bi := b.(uint64)
		switch {
		case a == bi:
			comp = 0
		case a < bi:
			comp = -1
		default:
			comp = 1
		}
	case float32:
		bi := b.(float32)
		switch {
		case a == bi:
			comp = 0
		case a < bi:
			comp = -1
		default:
			comp = 1
		}
	case float64:
		bi := b.(float64)
		switch {
		case a == bi:
			comp = 0
		case a < bi:
			comp = -1
		default:
			comp = 1
		}
	default:
		if ap == bp {
			comp = 0
			ok = true
		} else {
			ok = false
		}

		return
	}

	return
}

type sliceRepr interface {
	Add(index int, value reflect.Value)
	Remove(index int)
}

func performChange(repr sliceRepr, oa, na reflect.Value) {
	if oa.Len() < na.Len() {
		for i := 0; i < oa.Len(); i++ {
			nv := na.Index(i)
			if comp, _ := compareRefl(oa.Index(i), nv); comp != 0 {
				repr.Add(i, nv)
				return
			}
		}

		repr.Add(na.Len()-1, na.Index(na.Len()-1))
	} else if oa.Len() > na.Len() {
		for i := 0; i < na.Len(); i++ {
			if comp, _ := compareRefl(oa.Index(i), na.Index(i)); comp != 0 {
				repr.Remove(i)
				return
			}
		}

		repr.Remove(oa.Len() - 1)
	}
}
