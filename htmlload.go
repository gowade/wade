package wade

import (
	"fmt"
	"path"

	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/libs/http"
)

// GetHtml makes a request and gets the HTML contents
func (wd *wade) getHtml(httpClient *http.Client, href string) (string, error) {
	return getHtmlFile(httpClient, wd.serverBase, href)
}

func getHtmlFile(httpClient *http.Client, serverbase string, href string) (data string, err error) {
	resp, err := httpClient.GET(path.Join(serverbase, href))
	if resp.Failed() || err != nil {
		err = fmt.Errorf(`Failed to load HTML file "%v". Status code: %v. Error: %v.`, href, resp.StatusCode, err)
		return
	}

	data = resp.Data
	return
}

// htmlInclude performs HTML include
func htmlInclude(httpClient *http.Client, elem dom.Selection, serverbase string) (html string, err error) {
	src, ok := elem.Attr("src")
	if !ok {
		panic(dom.ElementError(elem, `winclude element has no "src" attribute`))
	}

	html, err = getHtmlFile(httpClient, serverbase, src)

	return
}
