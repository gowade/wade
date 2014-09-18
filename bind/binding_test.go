package bind

import (
	"testing"

	"github.com/phaikawl/wade/custom"
	"github.com/phaikawl/wade/dom"
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
		Name:       "test",
		Attributes: []string{"Name", "Value", "Test"},
		Html: `
			<div><wcontents></wcontents></div>
			<span #html="$Name"></span>
			<p #html="$Value"></p>
			<div #html="$Test.A.B"></div>
			<div><wcontents></wcontents></div>
			`,
		Prototype: &Model{},
	}
	b.tm.RegisterTags([]custom.HtmlTag{
		testct,
	})

	binder, args, err := parseBinderLHS("if")
	require.Equal(t, binder, "if")
	require.Equal(t, len(args), 0)
	require.Equal(t, err, nil)

	binder, args, err = parseBinderLHS("each(_,val)")
	require.Equal(t, binder, "each")
	require.Equal(t, len(args), 2)
	require.Equal(t, args[0], "_")
	require.Equal(t, args[1], "val")
	require.Equal(t, err, nil)

	binder, args, err = parseBinderLHS("ech(_,val")
	require.NotEqual(t, err, nil)

	//test processDomBind
	elem := goquery.GetDom().NewFragment("<div></div>")

	b.processBinderBind("html", "$Name", elem, bs, false)
	require.Equal(t, elem.Html(), "a")
	sc.Name = "b"
	b.watcher.Digest(&sc.Name)
	require.Equal(t, elem.Html(), "b")

	//test processFieldBind
	elem = goquery.GetDom().NewFragment("<test></test>")
	ct, _ := testct.NewElem(elem)
	model := ct.Model().(*Model)
	b.processFieldBind("Name", "':Hai;'", elem, bs, false, ct)
	b.processFieldBind("Value", "$Num", elem, bs, false, ct)
	require.Equal(t, model.Name, ":Hai;")
	require.Equal(t, model.Value, 9000)

	sc.Num = 9999
	b.watcher.Digest(&sc.Num)
	require.Equal(t, model.Value, 9999)

	//full test
	src := `
		<wroot>
			<ww @title="$Name" #class(awesome)="true">
				<div id="0">da{{ $Num }}n</div>
				<test @Value="$Num" @Name="'abc'" @Test="$Test" id="1">{{ $Name }}<!-- --></test>
				<div id="2">{{ $Test.A.B }}</div>
			</ww>
		</wroot>
	`
	sc.Name = "scope"
	root := goquery.GetDom().NewFragment(src)
	b.Bind(root, sc, false)
	getAttr := func(elem dom.Selection, attr string) string {
		a, _ := elem.Attr(attr)
		return a
	}
	require.Equal(t, getAttr(root.Find("#0"), "title"), sc.Name)
	require.Equal(t, getAttr(root.Find("#1"), "title"), sc.Name)
	require.Equal(t, root.Find("#0").HasClass("awesome"), true)
	require.Equal(t, root.Find("#1").Length(), 1)
	require.Equal(t, root.Find("#1").HasClass("awesome"), true)
	require.Equal(t, root.Find("#2").Html(), "true")

	felems := root.Find("#1").Children().Elements()
	require.Equal(t, felems[0].Text(), sc.Name)
	require.Equal(t, felems[1].Html(), "abc")
	require.Equal(t, felems[2].Html(), "9999")
	require.Equal(t, felems[3].Html(), "true")

	require.Equal(t, root.Find("#0").Html(), "da9999n")
	sc.Num = 6666
	b.Watcher().Digest(&sc.Num)
	require.Equal(t, root.Find("#0").Html(), "da6666n")
}
