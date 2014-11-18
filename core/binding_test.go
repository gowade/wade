package core

import (
	"testing"

	"github.com/phaikawl/wade/utils"
	"github.com/stretchr/testify/require"
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

func (b *TextBinder) Update(d DomBind) (err error) {
	d.Node.Children = []VNode{VText(utils.ToString(d.Value))}
	return
}

func (b *TextBinder) BindInstance() Binder { return b }

func TestBinding(t *testing.T) {
	b := NewTestBindEngine()
	sc := &Sc{"a", 9000, &TestModel{A{true}}}
	b.RegisterBinder("text", &TextBinder{})
	bs := b.newModelScope(sc)

	testct := ComponentView{
		Name: "test",
		Template: VNodeTemplate(
			VWrap("div", []VNode{
				VWrap("div", []VNode{
					VEmpty(CompInner),
				}),
				VElem("span", NoAttr(), []Bindage{BindBinder("text", "Name")}, []VNode{}),
				VElem("div", NoAttr(), []Bindage{BindBinder("text", "Test.A.B")}, []VNode{}),
			})),
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
	elem := NodeRoot(VElem("test", NoAttr(), NoBind(), []VNode{}))

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

	elem = NodeRoot(V(GhostNode, "div", NoAttr(), []Bindage{BindAttr("title", "Name")}, []VNode{
		VElem("test", map[string]interface{}{"name": "abc"}, []Bindage{
			BindAttr("Value", "Num"),
			BindAttr("Test", "Test"),
		}, []VNode{
			VMustache("Name"),
		}),

		VWrap("div", []VNode{
			VMustache("Test.A.B"),
		}),
	}))

	sc.Name = "scope"
	b.Bind(elem, sc)
	elem.Update()

	first := &elem.Children[0]
	second := &elem.Children[1]
	require.Equal(t, first.Attrs["title"].(string), sc.Name)
	require.Equal(t, second.Attrs["title"].(string), sc.Name)
	require.Equal(t, second.Text(), "true")

	fChildren := first.Children
	require.Equal(t, fChildren[0].Text(), sc.Name)
	require.Equal(t, fChildren[1].Text(), "abc")
	require.Equal(t, fChildren[2].Text(), "true")
}
