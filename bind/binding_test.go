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
	}

	Model struct {
		Name  string
		Value int
	}

	Scope struct {
		Name string
		Num  int
	}
)

func (ct *customTag) NewModel(elem dom.Selection) interface{} {
	return &Model{}
}

func (ct *customTag) PrepareTagContents(elem dom.Selection, model interface{}, fn func(dom.Selection)) error {
	return nil
}

func (cem *ceManager) GetCustomTag(elem dom.Selection) (CustomTag, bool) {
	if tn, err := elem.TagName(); err != nil && tn == "test" {
		return cem.ct, true
	}

	return nil, false
}

func (w *watcher) Watch(modelRefl reflect.Value, field string, callback func()) {
	w.watches = append(w.watches, callback)
}

func TestBinding(t *testing.T) {
	wc := &watcher{}
	cem := &ceManager{ct: &customTag{}}
	b := NewBindEngine(cem, wc)
	sc := &Scope{"a", 9000}
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
	b.processFieldBind("Name: 'Hai'; Value: Num", elem, bs, false, model)
	require.Equal(t, model.Name, "Hai")
	require.Equal(t, model.Value, 9000)
	sc.Num = 9999
	wc.watches[1]()
	require.Equal(t, model.Value, 9999)
}
