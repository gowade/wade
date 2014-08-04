package menu

import (
	"fmt"

	wd "github.com/phaikawl/wade"
)

type SwitchMenu struct {
	Current     string
	ActiveClass string
}

func (sm *SwitchMenu) Init(ce *wd.CustomElem) error {
	if sm.Current == "" {
		return fmt.Errorf(`"Current" attribute must be set.`)
	}

	cl := ce.Contents.Children("")
	if cl.Length != 1 || !cl.First().Is("ul") {
		return fmt.Errorf("switchmenu's contents must have exactly 1 child which is an <ul> element.")
	}

	for _, item := range wd.ToElemSlice(cl.First().Children("")) {
		if !item.Is("li") {
			return fmt.Errorf(`Direct children of the <ul> must be <li>.`)
		}

		if itemid := item.Attr("itemid"); itemid != "" {
			if itemid == sm.Current {
				item.AddClass(sm.ActiveClass)
			} else {
				item.RemoveClass(sm.ActiveClass)
			}
		} else {
			return fmt.Errorf(`"itemid" attribute must be set for each <li>.`)
		}
	}

	ce.Elem.Append(ce.Contents.Html())

	return nil
}

type Pagemenu struct {
	SwitchMenu
}

func Spec() map[string]interface{} {
	return map[string]interface{}{
		"switchmenu": SwitchMenu{},
		"pagemenu":   Pagemenu{},
	}
}
