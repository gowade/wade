package menu

import (
	"fmt"
	"strings"

	"github.com/phaikawl/wade"
	"github.com/phaikawl/wade/custom"
	"github.com/phaikawl/wade/dom"
)

type SwitchMenu struct {
	custom.BaseProto
	Current     string
	ActiveClass string
	Choices     map[string]dom.Selection
}

func (sm *SwitchMenu) ProcessContents(ctl custom.ContentsCtl) error {
	sm.Choices = make(map[string]dom.Selection)
	sm.ActiveClass = strings.TrimSpace(sm.ActiveClass)
	if sm.ActiveClass == "" {
		sm.ActiveClass = "active"
	}

	if sm.Current == "" {
		return fmt.Errorf(`"Current" attribute must be set`)
	}

	cl := ctl.Contents()
	ul := cl.Filter("ul")
	if cl.Length() == 0 {
		return fmt.Errorf("switchmenu must have 1 child which is an <ul> element.")
	}

	for _, item := range ul.Children().Elements() {
		if wade.IsWrapperElem(item) {
			item = item.Children().Filter("li").First()
		}

		if !item.Is("li") {
			return fmt.Errorf(`Direct children of the <ul> must be <li>.`)
		}

		if casestr, ok := item.Attr("case"); ok {
			cases := strings.Split(casestr, " ")
			for _, id := range cases {
				if _, exists := sm.Choices[id]; exists {
					return fmt.Errorf("Switchmenu case %v is duplicated in multiple items.", id)
				}

				sm.Choices[strings.TrimSpace(id)] = item
			}

		} else {
			return fmt.Errorf(`"case" attribute must be set for each <li>.`)
		}
	}

	return nil
}

func (sm *SwitchMenu) Update(ctl custom.ElemCtl) error {
	ctl.Element().Find("li." + sm.ActiveClass).RemoveClass(sm.ActiveClass)
	if item, ok := sm.Choices[sm.Current]; ok {
		item.AddClass(sm.ActiveClass)
	}

	return nil
}

func HtmlTags() []custom.HtmlTag {
	return []custom.HtmlTag{
		custom.HtmlTag{
			Name:       "switchmenu",
			Attributes: []string{"Current", "ActiveClass"},
			Prototype:  &SwitchMenu{},
			Html:       `<wcontents></wcontents>`,
		},
	}
}
