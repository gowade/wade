package menu

import (
	"fmt"
	"strings"

	"github.com/phaikawl/wade/core"
)

type SwitchMenu struct {
	core.BaseProto
	Current         string
	ActiveClassName string
	Choices         map[string]*core.VNode
}

func (sm *SwitchMenu) ProcessInner(node core.VNode) {
	sm.Choices = make(map[string]*core.VNode)
	sm.ActiveClassName = strings.TrimSpace(sm.ActiveClassName)
	if sm.ActiveClassName == "" {
		sm.ActiveClassName = "active"
	}

	if sm.Current == "" {
		panic(`Field "Current" has not been set.`)
	}

	children := node.ChildElems()
	if len(children) != 1 || children[0].Data != "ul" {
		panic(`Must have 1 child and it must be an "ul" element.`)
	}

	for _, item := range children[0].ChildElems() {
		if item.TagName() != "li" {
			continue
		}

		if casestr, ok := item.Attr("case"); ok {
			cases := strings.Split(casestr.(string), " ")
			for _, id := range cases {
				if _, exists := sm.Choices[id]; exists {
					panic(fmt.Errorf("case %v is duplicated in multiple items.", id))
				}

				sm.Choices[strings.TrimSpace(id)] = item
			}
		}
	}
}

func (sm *SwitchMenu) Update(node core.VNode) {
	for key, item := range sm.Choices {
		item.SetClass(sm.ActiveClassName, key == sm.Current)
	}
}

func Components() []core.ComponentView {
	return []core.ComponentView{
		{
			Name:      "w-switch-menu",
			Prototype: &SwitchMenu{},
			Template:  core.VNode{Data: core.CompInner},
		},
	}
}
