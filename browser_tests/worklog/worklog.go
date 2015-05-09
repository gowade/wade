package worklog

import (
	"strings"
	"time"

	"github.com/gowade/wade"
)

type Project struct {
	wade.Com
	ID    int
	Title string
}

type WorklogState struct {
	FilterText string
	Projects   []*Project
}

type Worklog struct {
	wade.Com
	State *WorklogState `fuel:"state"`
}

func (this *Worklog) handleSearch(filterText string) {
	this.SetFilterText(strings.ToLower(filterText))
}

type SearchBar struct {
	wade.Com
	FilterText string
	OnSearch   func(string)
}

func (this *SearchBar) handleSearch() {
	this.OnSearch(this.Refs().filterTextInput.Value())
}

type LogTable struct {
	wade.Com
	FilterText string
	Projects   []*Project
}

func (this LogTable) filterCheck(text string) bool {
	return strings.Contains(strings.ToLower(text), this.FilterText)
}

type LogRowTimerState struct {
	Elapsed float32
	Running bool
}

type LogRow struct {
	wade.Com
	*Project

	ticker *time.Ticker
	State  *LogRowTimerState `fuel:"state"`
}

func (this *LogRow) toggleClock() {
	if !this.State.Running {
		this.State.Running = true
		this.ticker = time.NewTicker(100 * time.Millisecond)
		go func() {
			for {
				<-this.ticker.C
				this.SetElapsed(this.State.Elapsed + 0.1)
			}
		}()
	} else {
		this.ticker.Stop()
		this.SetRunning(false)
	}
}
