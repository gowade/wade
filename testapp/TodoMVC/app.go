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

func (t *TodoEntry) ToggleEdit() {
	if t.State == stateEditing {
		t.setCompleteState()
	} else {
		t.State = stateEditing
	}
}

func (t *TodoEntry) Destroy() {
	println("Would Destroy:" + t.Text)
}

func (t *TodoEntry) ToggleDone() {
	println("Would ToggleDone:" + t.Text)
	t.Done = !t.Done
	t.setCompleteState()

}

func (t *TodoEntry) setCompleteState() {
	if t.Done {
		t.State = stateCompleted
	} else {
		t.State = ""
	}
}

func main() {
	wade := wd.WadeUp("pg-home", "/web", func(wade *wd.Wade) {
		wade.Pager().RegisterPages("wpage-root")

		wade.Custags().RegisterNew("todoentry", "t-todoentry", TodoEntry{})
	})

	wade.Start()
}
