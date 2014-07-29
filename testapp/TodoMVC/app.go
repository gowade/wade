package main

import wd "github.com/phaikawl/wade"

// the different states a TodoEntry can be in
const (
	stateEditing   = "editing"
	stateCompleted = "completed"
)

// TodoEntry represents a single entry in the todo list
type TodoEntry struct {
	Text  string
	Done  bool
	State string
}

type todoEntryTag struct {
	Entry  *TodoEntry
	Key    int
	Delete func(int) func()
}

// ToggleEdit updates the state for the TodoEntry
func (t *TodoEntry) ToggleEdit() {
	if t.State == stateEditing {
		t.setCompleteState()
	} else {
		t.State = stateEditing
	}
}

// ToggleDone switches the Done field on or off
func (t *TodoEntry) ToggleDone() {
	println("clicked ToggleDone:" + t.Text)
	t.Done = !t.Done
	t.setCompleteState()
}

// setCompleteState is just a small helper to reuse this if
func (t *TodoEntry) setCompleteState() {
	if t.Done {
		t.State = stateCompleted
	} else {
		t.State = ""
	}
}

type TodoView struct {
	NewEntry string
	Entries  []*TodoEntry
}

//
func (t *TodoView) ToggleAll() {
	println("clicked ToggleAll")
	for _, e := range t.Entries {
		e.ToggleDone()
	}
}

func (t *TodoView) AddEntry() {
	if t.NewEntry != "" {
		println("Adding:'" + t.NewEntry + "'")
		t.Entries = append(t.Entries, &TodoEntry{Text: t.NewEntry})
		t.NewEntry = ""
	}
}

func (t *TodoView) DeleteEntry(i int) func() {
	return func() {
		t.Entries = append(t.Entries[0:i], t.Entries[i+1:]...)
	}
}

func main() {
	wadeApp := wd.WadeUp("pg-main", "/todo", func(wade *wd.Wade) {
		wade.Pager().RegisterPages("wpage-root")

		// our custom tags
		wade.Custags().RegisterNew("todoentry", "t-todoentry", todoEntryTag{})

		// our main controller
		wade.Pager().RegisterController("pg-main", func(p *wd.PageCtrl) interface{} {
			println("called RegisterController for pg-main")
			view := new(TodoView)

			view.Entries = []*TodoEntry{
				&TodoEntry{Text: "create a datastore for entries", Done: true},
				&TodoEntry{Text: "add new entries", Done: true},
				&TodoEntry{Text: "toggle edit off - click anywhere else"},
				&TodoEntry{Text: "ToggleAll should do something", Done: true},
				&TodoEntry{Text: "destroy -> delete from the list", Done: true},
				&TodoEntry{Text: "add filters for state"},
				&TodoEntry{Text: "update counters in footer"},
			}

			// update the t.State
			// might be better to bind to Done directly
			for _, e := range view.Entries {
				e.setCompleteState()
			}
			return view
		})
	})

	wadeApp.Start()
}
