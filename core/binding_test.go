package core

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/phaikawl/wade/scope"
	"github.com/phaikawl/wade/utils"
)

type (
	Model struct {
		BaseProto
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

	Sc struct {
		Name string
		Num  int
		Test *TestModel
	}

	TextBinder struct {
		BaseBinder
	}
)

func (m Model) TestFn() string {
	return "ok"
}

func (m Model) Update(vnode VNode) {
	vnode.Children[0].SetClass("updated", true)
}

func (b TextBinder) Update(d DomBind) {
	d.Node.Children = []VNode{VText(utils.ToString(d.Value))}
	return
}

func TestBinding(t *testing.T) {
	b := NewBindEngine(nil, map[string]interface{}{})
	sc := &Sc{"a", 9000, &TestModel{A{true}}}
	b.RegisterBinder("text", TextBinder{})
	bs := bindScope{scope.NewScope(sc)}

	testct := ComponentView{
		Name: "test",
		Template: VPrep(VNode{
			Data: "div",
			Children: []VNode{
				{
					Data:     "div",
					Children: []VNode{{Data: CompInner}},
				},
				{
					Data:  "span",
					Binds: []Bindage{BindBinder("text", "Name")},
				},
				{
					Data:  "div",
					Binds: []Bindage{BindBinder("text", "Test.A.B")},
				},
				{
					Data:  "div",
					Binds: []Bindage{BindBinder("text", "TestFn()")},
				},
			},
		}),
		Prototype: &Model{},
	}

	b.tm.Register(testct)

	binder, args, err := parseBinderLHS("text")
	require.Equal(t, binder, "text")
	require.Equal(t, len(args), 0)
	require.Equal(t, err, nil)

	binder, args, err = parseBinderLHS("text(_,val)")
	require.Equal(t, binder, "text")
	require.Equal(t, len(args), 2)
	require.Equal(t, args[0], "_")
	require.Equal(t, args[1], "val")
	require.Equal(t, err, nil)

	binder, args, err = parseBinderLHS("unregistered(_,val")
	require.NotEqual(t, err, nil)

	//test processDomBind
	elem := VPrep(VNode{Data: "test"}).Ptr()

	b.processBinderBind("text", "Name", elem, bs)
	elem.Update()
	require.Equal(t, elem.Text(), "a")
	sc.Name = "b"

	elem.Update()
	require.Equal(t, elem.Text(), "b")

	//test processFieldBind
	tct, ok := b.tm.GetComponent("test")
	if !ok {
		t.FailNow()
	}

	ct, _ := tct.NewInstance(elem)
	model := ct.model.(*Model)
	b.processFieldBind("Name", "':Hai;'", elem, bs, ct)
	b.processFieldBind("Value", "Num", elem, bs, ct)
	elem.Update()
	require.Equal(t, model.Name, ":Hai;")
	require.Equal(t, model.Value, 9000)

	sc.Num = 9999
	elem.Update()
	require.Equal(t, model.Value, 9999)

	elem = VPrep(VNode{
		Data: "div",
		Children: []VNode{
			{
				Data:  "test",
				Attrs: Attributes{"name": "abc"},
				Binds: []Bindage{
					BindAttr("Value", "Num"),
					BindAttr("Test", "Test"),
				},
				Children: []VNode{VMustache("Name")},
			},
			{
				Type:     GroupNode,
				Data:     "div",
				Children: []VNode{VMustache("Test.A.B")},
			},
		},
	}).Ptr()

	sc.Name = "scope"
	b.Bind(elem, sc)
	elem.Update()

	require.True(t, elem.Children[0].Children[0].HasClass("updated"))

	first := &elem.Children[0]
	second := &elem.Children[1]
	require.Equal(t, second.Text(), "true")

	if len(first.Children) == 1 {
		first = &first.Children[0]
	}

	fChildren := first.Children
	require.Equal(t, fChildren[0].Text(), sc.Name)
	require.Equal(t, fChildren[1].Text(), "abc")
	require.Equal(t, fChildren[2].Text(), "true")
	require.Equal(t, fChildren[3].Text(), "ok")
}
