package bind

import (
	"testing"

	"github.com/phaikawl/wade/custom"
	"github.com/phaikawl/wade/dom/goquery"
	"github.com/stretchr/testify/require"
)

type (
	Model struct {
		custom.BaseProto
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

func TestBinding(t *testing.T) {
	b := NewTestBindEngine()
	sc := &Scope{"a", 9000, &TestModel{A{true}}}
	bs := &bindScope{b.newModelScope(sc)}

	testct := custom.HtmlTag{
		Name: "test",
		Html: `
			<wcontents></wcontents>
			<span bind-html="Name"></span>
			<p bind-html="Value"></p>
			<div bind-html="Test.A.B"></div>
			<wcontents></wcontents>
			`,
		Prototype: &Model{},
	}
	b.tm.RegisterTags([]custom.HtmlTag{
		testct,
	})

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
	b.watcher.ApplyChanges(&sc.Name)
	require.Equal(t, elem.Html(), "b")

	//test processFieldBind
	elem = goquery.GetDom().NewFragment("<test></test>")
	ct := testct.NewElem(elem)
	model := ct.Model().(*Model)
	b.processFieldBind("Name: |':Hai;'; Value: Num;", elem, bs, false, ct)
	require.Equal(t, model.Name, ":Hai;")
	require.Equal(t, model.Value, 9000)

	sc.Num = 9999
	b.watcher.ApplyChanges(&sc.Num)
	require.Equal(t, model.Value, 9999)

	//full test
	src := `
		<wroot>
			<ww bind-attr-class="Name">
				<div id="0" bind-html="Num"></div>
				<test bind="Value: Num; Name: |'abc'; Test: Test" bind-attr-id="|1"><span bind-html="Name"></span><!-- --></test>
				<div id="2" bind-html="Test.A.B"></div>
			</ww>
		</wroot>
	`
	sc.Name = "scope"
	root := goquery.GetDom().NewFragment(src)
	b.Bind(root, sc, false)
	require.Equal(t, root.Find("#0").HasClass(sc.Name), true)
	require.Equal(t, root.Find("#1").HasClass(sc.Name), true)
	require.Equal(t, root.Find("#0").Html(), "9999")
	require.Equal(t, root.Find("#1").Length(), 1)
	require.Equal(t, root.Find("#2").Html(), "true")

	felems := root.Find("#1").Children().Elements()
	require.Equal(t, felems[0].Html(), sc.Name)
	require.Equal(t, felems[1].Html(), "abc")
	require.Equal(t, felems[2].Html(), "9999")
	require.Equal(t, felems[3].Html(), "true")
}
