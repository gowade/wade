package markman

import (
	"fmt"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
)

const (
	IncludeTag  = "w-include"
	AppViewAttr = "!appview"
)

type (
	SrcFetcher interface {
		FetchFile(file string) (string, error)
	}

	MarkupManager struct {
		document  dom.Selection
		fetcher   SrcFetcher
		container dom.Selection
		vdom      core.VNode
	}
)

func New(document dom.Selection, fetcher SrcFetcher) (mm *MarkupManager, err error) {
	c := mm.document.Find("[" + AppViewAttr + "]")
	if c.Length() == 0 {
		err = fmt.Errorf("No view container (element with !appview attribute found.")
		return
	}
	c = c.First()

	mm = &MarkupManager{
		document:  document,
		fetcher:   fetcher,
		container: c,
	}

	err = mm.loadView()

	return
}

func (mm *MarkupManager) loadView() (err error) {
	file, ok := mm.container.Attr(AppViewAttr)
	if !ok {
		panic("WTF?")
	}

	importCtn := mm.container.Clone()

	src, err := mm.fetcher.FetchFile(file)
	if err != nil {
		return
	}

	importCtn.Append(importCtn.NewFragment(src))

	err = mm.htmlImports(importCtn)
	if err != nil {
		return
	}

	mm.vdom = importCtn.ToVNode()
	return
}

func (mm MarkupManager) htmlImports(container dom.Selection) (err error) {
	return HTMLImports(mm.fetcher, container)
}

func HTMLImports(fetcher SrcFetcher, container dom.Selection) (err error) {
	imports := container.Find(IncludeTag).Elements()
	if len(imports) == 0 {
		return nil
	}

	queueChan := make(chan bool, len(imports))
	finishChan := make(chan error, 1)

	for _, elem := range imports {
		src, ok := elem.Attr("src")
		if !ok {
			return dom.ElementError(elem, IncludeTag+` element has no "src" attribute`)
		}

		go func(elem dom.Selection) {
			var err error
			var html string
			html, err = fetcher.FetchFile(src)
			if err != nil {
				finishChan <- err
				return
			}

			// the go html parser will refuse to work if the content is only text, so
			// we put a wrapper here
			newElem := container.NewFragment("<div !ghost>" + html + "</div>")
			if belong, hasbelong := elem.Attr(core.BelongAttrName); hasbelong {
				newElem.SetAttr(core.BelongAttrName, belong)
			}

			elem.ReplaceWith(newElem)

			err = HTMLImports(fetcher, newElem)
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

	err = <-finishChan

	return
}
