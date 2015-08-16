package worklog

import (
	gourl "net/url"
	"strings"
	"time"

	//"github.com/gowade/wade"
	dummy "github.com/gowade/wade/browser_tests/worklog/dummypkg"
)

type Project struct {
	ID    int
	Title string
}

type Worklog struct {
	Projects   []*Project `fstate`
	FilterText string     `fstate`
	demoState  *gourl.URL `fstate`
	dummy.DemoStruct
	StateField int `fstate`
}

func (this *Worklog) handleSearch(filterText string) {
	this.setFilterText(strings.ToLower(filterText))
}

type SearchBar struct {
	FilterText string
	OnSearch   func(string)
}

func (this *SearchBar) handleSearch() {
	this.OnSearch(this.Refs().filterTextInput.Value())
}

type LogTable struct {
	FilterText string
	Projects   []*Project
}

func (this LogTable) filterCheck(text string) bool {
	return strings.Contains(strings.ToLower(text), this.FilterText)
}

type LogRow struct {
	*Project

	ticker  *time.Ticker
	Elapsed float32 `fstate`
	Running bool    `fstate`
}

func (this *LogRow) toggleClock() {
	if !this.Running {
		this.Running = true
		this.ticker = time.NewTicker(100 * time.Millisecond)
		go func() {
			for {
				<-this.ticker.C
				this.setElapsed(this.Elapsed + 0.1)
			}
		}()
	} else {
		this.ticker.Stop()
		this.setRunning(false)
	}
}

type Hello struct {
	Name string
}
