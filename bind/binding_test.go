package bind

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/dom/goquery"
)

type (
	watcher struct {
		watches []func()
	}

	ceManager struct {
		ct CustomTag
	}

	customTag struct {
		html string
	}

	Model struct {
		Name  string
		Value int
		Test  *TestModel
	}

	TestModel struct {
		A A
	}

	A struct {
		B bool
	}

	Scope struct {
		Name string
		Num  int
		Test *TestModel
	}
)

func (ct *customTag) NewModel(elem dom.Selection) interface{} {
	return &Model{}
}

func (ct *customTag) PrepareTagContents(elem dom.Selection, model interface{}, fn func(dom.Selection)) error {
	ce := elem.Clone()
	elem.SetHtml(ct.html)
	elem.Find("wcontents").ReplaceWith(ce)
	fn(ce.Contents())
	ce.Unwrap()
	return nil
}

func (cem *ceManager) GetCustomTag(elem dom.Selection) (CustomTag, bool) {
	if tn, err := elem.TagName(); err == nil && tn == "test" {
		return cem.ct, true
	}

	return nil, false
}

func (w *watcher) Watch(modelRefl reflect.Value, field string, callback func()) {
	w.watches = append(w.watches, callback)
}

func initTestBind() (wc *watcher, cem *ceManager, b *Binding) {
	wc = &watcher{make([]func(), 0)}
	cem = &ceManager{ct: &customTag{}}
	return wc, cem, NewBindEngine(cem, wc)
}

func TestBinding(t *testing.T) {
	wc, cem, b := initTestBind()
	sc := &Scope{"a", 9000, &TestModel{A{true}}}
	bs := &bindScope{b.newModelScope(sc)}

	//Test parse dom bind string
	tbs := "kdfk(dfdf)"
	bexpr, outputs, err := parseDomBindstr(tbs)
	require.Equal(t, err, nil)
	require.Equal(t, bexpr, tbs)
	require.Equal(t, len(outputs), 0)

	tbs = "zzz"
	bexpr, outputs, err = parseDomBindstr(tbs + "   ->  abc, def")
	require.Equal(t, err, nil)
	require.Equal(t, bexpr, tbs)
	require.Equal(t, len(outputs), 2)
	require.Equal(t, outputs[0], "abc")
	require.Equal(t, outputs[1], "def")

	//test processDomBind
	elem := goquery.GetDom().NewFragment("<div></div>")

	b.processDomBind("bind-html", "Name", elem, bs, false)
	require.Equal(t, elem.Html(), "a")
	sc.Name = "b"
	wc.watches[0]()
	require.Equal(t, elem.Html(), "b")

	//test processFieldBind
	elem = goquery.GetDom().NewFragment("<test></test>")
	model := &Model{}
	b.processFieldBind("Name: |':Hai;'; Value: Num;", elem, bs, false, model)
	require.Equal(t, model.Name, ":Hai;")
	require.Equal(t, model.Value, 9000)

	sc.Num = 9999
	wc.watches[1]()
	require.Equal(t, model.Value, 9999)

	//full test
	tag, ok := cem.GetCustomTag(elem)
	if !ok {
		panic("WTF is wrong with the tag?")
	}
	tag.(*customTag).html = `
	<div>
		<wcontents></wcontents>
		<span bind-html="Name"></span>
		<p bind-html="Value"></p>
		<div bind-html="Test.A.B"></div>
	</div>
	`
	src := `
	<div>
		<ww bind-attr-class="Name">
			<div id="0" bind-html="Num"></div>
			<test bind="Value: Num; Name: |'abc'; Test: Test" bind-attr-id="|1"><span bind-html="Name"></span></test>
			<div id="2" bind-html="Test.A.B"></div>
		</ww>
	</div>
	`
	sc.Name = "scope"
	root := goquery.GetDom().NewFragment(src)
	b.Bind(root, sc, false, false)
	require.Equal(t, root.Find("#0").HasClass(sc.Name), true)
	require.Equal(t, root.Find("#1").HasClass(sc.Name), true)
	require.Equal(t, root.Find("#0").Html(), "9999")
	require.Equal(t, root.Find("#1").Length(), 1)
	require.Equal(t, root.Find("#2").Html(), "true")
	tn, _ := root.Find("#1").TagName()
	require.Equal(t, tn, "div")
	felems := root.Find("#1").Children().Elements()
	require.Equal(t, felems[0].Html(), sc.Name)
	require.Equal(t, felems[1].Html(), "abc")
	require.Equal(t, felems[2].Html(), "9999")
	require.Equal(t, felems[3].Html(), "true")
}
