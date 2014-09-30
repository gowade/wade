package bind

import (
	"reflect"
	"sort"
)

type ByRefl []*SliceChange

func (a ByRefl) Len() int           { return len(a) }
func (a ByRefl) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRefl) Less(i, j int) bool { comp, _ := compareRefl(a[i].val, a[j].val); return comp == -1 }

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
		ok = false
	}

	return
}

type SliceChange struct {
	val reflect.Value
	idx int
}

func SliceDiff(oa, na reflect.Value) (added, removed []*SliceChange) {
	added = make([]*SliceChange, 0)
	removed = make([]*SliceChange, 0)

	if oa.Type() != na.Type() {
		panic("Calling slice diff on values of different types")
	}

	os := make([]*SliceChange, oa.Len())
	ns := make([]*SliceChange, na.Len())
	for i := 0; i < len(os); i++ {
		os[i] = &SliceChange{oa.Index(i), i}
	}

	for i := 0; i < len(ns); i++ {
		ns[i] = &SliceChange{na.Index(i), i}
	}

	sort.Sort(ByRefl(os))
	sort.Sort(ByRefl(ns))

	io, in := 0, 0
	for {
		if io >= oa.Len() || in >= na.Len() {
			break
		}

		switch comp, _ := compareRefl(os[io].val, ns[in].val); comp {
		case 0:
			io++
			in++
		case -1:
			removed = append(removed, os[io])
			io++
		case 1:
			added = append(added, ns[in])
			in++
		}
	}

	if io < oa.Len() {
		removed = append(removed, os[io:]...)
	}

	if in < na.Len() {
		added = append(added, ns[in:]...)
	}

	return
}

type sliceRepr interface {
	Add(index int, value reflect.Value)
	Remove(index int)
}

func performChange(repr sliceRepr, oa, na reflect.Value) {
	added, removed := SliceDiff(oa, na)
	for i, r := range removed {
		repr.Remove(r.idx - i)
	}

	for _, ad := range added {
		repr.Add(ad.idx, ad.val)
	}
}
