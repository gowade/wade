package fntest

import (
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/dom/goquery"
)

type (
	TestView struct {
		document dom.Selection
		rsessId  int64
	}

	//Selection struct {
	//	sel     dom.Selection
	//	rsessId int64
	//	view    *TestView
	//}

	selector func(dom.Selection) dom.Selection

	Selection struct {
		selector selector
		parent   *Selection
	}
)

func (s Selection) doSelect() dom.Selection {
	var d dom.Selection
	if s.parent != nil {
		d = s.parent.doSelect()
	}

	return s.selector(d)
}

func (s *Selection) spawn(selector selector) *Selection {
	return &Selection{
		selector: selector,
		parent:   s,
	}
}

func (s *Selection) First() *Selection {
	return s.spawn(func(ds dom.Selection) dom.Selection {
		return ds.First()
	})
}

func (s *Selection) Eq(index int) *Selection {
	return s.spawn(func(ds dom.Selection) dom.Selection {
		return ds.Elements()[index]
	})
}

func (s *Selection) Find(selector string) *Selection {
	return s.spawn(func(ds dom.Selection) dom.Selection {
		return ds.Find(selector)
	})
}

func (s Selection) Text() string {
	return s.doSelect().Text()
}

func (s Selection) HasClass(class string) bool {
	return s.doSelect().HasClass(class)
}

//func (s Selection) First() Selection {
//	return s.spawnSelection(s.sel.First())
//}

//func (s Selection) Eq(index int) Selection {
//	return s.spawnSelection(s.sel.Elements()[index])
//}

//func (s Selection) Find(selector string) Selection {
//	return s.spawnSelection(s.sel.Find(selector))
//}

//func (s Selection) Text() string {
//	s.check()
//	return s.sel.Text()
//}

//func (s Selection) check() {
//	if s.view.rsessId != s.rsessId {
//		panic("Reuse of selection after a render is forbidden (unpredictable behavior).")
//	}
//}

//func (s Selection) spawnSelection(ds dom.Selection) Selection {
//	s.check()
//	return Selection{ds, s.view.rsessId, s.view}
//}

func (tv TestView) Title() string {
	return tv.document.Find("head title").Text()
}

func (tv *TestView) Find(selector string) *Selection {
	return &Selection{
		selector: func(_ dom.Selection) dom.Selection {
			return tv.document.Find(selector)
		},
	}
}

func (tv TestView) Content() string {
	return tv.document.Find("body").Text()
}

// TriggerEvent triggers a given event on the selected elements
func (tv TestView) TriggerEvent(selection *Selection, event Event) {
	for _, elem := range selection.doSelect().Elements() {
		event.Event().propaStopped = false
		event.Event().target = elem
		triggerRec(elem.(goquery.Selection).Nodes[0], event)
	}
}
