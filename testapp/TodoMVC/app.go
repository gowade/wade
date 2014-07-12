package main

import (
	wd "github.com/phaikawl/wade"
)

func main() {
	wade := wd.WadeUp("pg-home", "/web", func(wade *wd.Wade) {
		wade.Pager().RegisterPages("wpage-root")
		wade.Pager().SetNotFoundPage("pg-not-found")
	})
	// Should must literally be called at the bottom of every Wade application
	// for whatever the reason
	wade.Start()
}
