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

// htmlImport performs HTML include
func htmlImport(httpClient *http.Client, parent dom.Selection, serverbase string) error {
	imports := parent.Find("winclude").Elements()
	if len(imports) == 0 {
		return nil
	}

	queueChan := make(chan bool, len(imports))
	finishChan := make(chan error, 1)

	for _, elem := range imports {
		src, ok := elem.Attr("src")
		if !ok {
			return dom.ElementError(elem, `winclude element has no "src" attribute`)
		}

		go func(elem dom.Selection) {
			var err error
			var html string
			html, err = getHtmlFile(httpClient, serverbase, src)
			if err != nil {
				finishChan <- err
				return
			}

			// the go html parser will refuse to work if the content is only text, so
			// we put a wrapper here
			ne := parent.NewFragment("<ww>" + html + "</ww>")
			if belong, hasbelong := elem.Attr("w-belong"); hasbelong {
				ne.SetAttr("w-belong", belong)
			}

			elem.ReplaceWith(ne)

			err = htmlImport(httpClient, ne, serverbase)
			if err != nil {
				finishChan <- err
				return
			}

			queueChan <- true
			if len(queueChan) == len(imports) {
				finishChan <- nil
			}
		}(elem)
	}

	err := <-finishChan
	if err != nil {
		return err
	}

	return nil
}
