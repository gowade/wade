package markman

import (
	"fmt"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/vquery"
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
		importCtn dom.Selection
		origVdom  *core.VNode
		vdom      *core.VNode
		onReady   []func()
	}

	NopFetcher struct{}

	TemplateConverter struct {
		*MarkupManager
		queue []func()
	}
)

func (f NopFetcher) FetchFile(file string) (string, error) {
	return "", nil
}

func (tc TemplateConverter) FromString(template string) core.VNode {
	return tc.document.NewFragment(template).ToVNode()
}

func (tc *TemplateConverter) processQueue() {
	for _, fn := range tc.queue {
		fn()
	}
}

func (tc *TemplateConverter) FromHTMLTemplate(templatePtr *core.VNode, templateId string) core.VNode {
	tc.queue = append(tc.queue, func() {
		vn := core.VNode{
			Type:     core.GroupNode,
			Children: []core.VNode{},
		}

		template := tc.importCtn.Find("template#" + templateId)
		if template.Length() == 0 {
			panic(fmt.Errorf(`Cannot find HTML Template "%v".`, templateId))
		}

		children := template.Children().Elements()

		for _, c := range children {
			vn.Children = append(vn.Children, c.ToVNode())
		}

		*templatePtr = vn
	})

	// return a temporary dummy template, replaced later
	return core.VPrep(core.VNode{
		Type: core.GroupNode,
		Data: "component",
	})
}

func (m *MarkupManager) TemplateConverter() *TemplateConverter {
	tc := &TemplateConverter{m, make([]func(), 0)}
	m.onReady = append(m.onReady, tc.processQueue)
	return tc
}

func (m MarkupManager) Document() dom.Selection {
	return m.document
}

func New(document dom.Selection, fetcher SrcFetcher) (mm *MarkupManager) {
	c := document.Find("[\\" + AppViewAttr + "]")
	if c.Length() == 0 {
		panic(fmt.Errorf(`No view container (element with "%v" attribute found.`, AppViewAttr))
	}
	c = c.First()

	mm = &MarkupManager{
		document:  document,
		fetcher:   fetcher,
		container: c,
		importCtn: c,
		onReady:   make([]func(), 0),
	}

	return
}

func (mm MarkupManager) Container() dom.Selection {
	return mm.container
}

func (mm MarkupManager) Render() {
	mm.vdom.Update()
	mm.container.Render(mm.vdom)
}

func (mm MarkupManager) VDom() *core.VNode {
	return mm.vdom
}

func (mm *MarkupManager) MarkupPage(title string, condFn core.CondFn) *core.VNode {
	if mm.origVdom == nil {
		panic("View has not been loaded.")
	}

	mm.vdom = mm.origVdom.CloneWithCond(condFn).Ptr()

	headElem := mm.document.Find("head").First()
	titleElem := headElem.Find("title")
	if titleElem.Length() == 0 {
		titleElem = mm.document.NewFragment("<title></title>")
		headElem.Append(titleElem)
	}

	titleElem.SetHtml(title)

	return mm.vdom
}

func (mm *MarkupManager) LoadView() (err error) {
	file, ok := mm.container.Attr(AppViewAttr)
	if !ok {
		panic("WTF?")
	}

	importCtn := mm.container

	if file != "" {
		importCtn = mm.container.Clone()

		var src string
		src, err = mm.fetcher.FetchFile(file)
		if err != nil {
			return
		}

		importCtn.Append(importCtn.NewFragment(src))
	}

	err = mm.htmlImports(importCtn)
	if err != nil {
		return
	}

	mm.origVdom = importCtn.ToVNode().Ptr()
	for _, tmpl := range vq.New(mm.origVdom).Find(vq.Selector{Tag: "template"}) {
		tmpl.Type = core.DeadNode
	}
	mm.vdom = mm.origVdom
	mm.importCtn = importCtn
	for _, fn := range mm.onReady {
		fn()
	}

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
			newElem := container.NewFragment("<div !group>" + html + "</div>")
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
