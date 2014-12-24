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
		templates map[string]*core.VNode
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
		template, ok := tc.templates[templateId]
		if !ok {
			panic(fmt.Errorf(`Cannot find HTML Template "%v".`, templateId))
		}

		*templatePtr = *template
	})

	// return a temporary dummy template, replaced later
	return core.VPrep(core.VNode{
		Type: core.GroupNode,
		Data: "template",
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
		templates: make(map[string]*core.VNode),
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
		//gopherjs:blocking
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
		if id, ok := tmpl.Attr("id"); ok && id != "" {
			mm.templates[id.(string)] = tmpl
		}
	}

	for _, imp := range vq.New(mm.origVdom).Find(vq.Selector{Tag: IncludeTag}) {
		imp.Type = core.GroupNode
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
			//gopherjs:blocking
			html, err = fetcher.FetchFile(src)
			if err != nil {
				finishChan <- err
				return
			}

			elem.SetHtml(html)
			err = HTMLImports(fetcher, elem)
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
