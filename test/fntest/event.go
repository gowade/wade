package test

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/dom/goquery"
)

type (
	Event interface {
		dom.Event
		Event() *event
	}

	event struct {
		target       dom.Selection
		typ          string
		propaStopped bool
	}

	KeyEvent struct {
		*event
		keyCode int
	}

	MouseEvent struct {
		*event
		button     int
		posX, posY int
	}
)

// NewEvent creates a new event
func NewEvent(eventType string) Event {
	return &event{nil, eventType, false}
}

// NewEvent creates a new key event
func NewKeyEvent(eventType string, keyCode int) dom.KeyEvent {
	return &KeyEvent{NewEvent(eventType).(*event), keyCode}
}

// NewEvent creates a new mouse event
func NewMouseEvent(eventType string, button, posX, posY int) dom.MouseEvent {
	return &MouseEvent{NewEvent(eventType).(*event), button, posX, posY}
}

func (e *event) StopPropagation() {
	e.propaStopped = true
}

func (e *event) Target() dom.Selection {
	return e.target
}

func (e *event) PreventDefault() {}

func (e *event) Type() string {
	return e.typ
}

func (e *event) Event() *event {
	return e
}

func (e *event) Js() js.Object { return nil }

func (e *KeyEvent) Which() int {
	return e.keyCode
}

func (e *MouseEvent) Which() int {
	return e.button
}

func (e *MouseEvent) Pos() (int, int) {
	return e.posX, e.posY
}

func triggerRec(selection dom.Selection, event Event) {
	goquery.TriggerEvent(selection, event)

	if !event.Event().propaStopped {
		if parent := selection.Parent(); parent.Length() > 0 {
			triggerRec(parent, event)
		}
	}
}
