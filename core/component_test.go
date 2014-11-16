package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type (
	Test struct {
		BaseProto
		Str  string
		Num  int
		Fnum float32
		Tf   bool
	}
)

func (t *Test) Init(node VNode) error {
	NodeWalk(&node, func(node *VNode) {
		if node.TagName() == "smile" {
			*node = VText(":D")
		}
	})

	return nil
}

func TestComponent(t *testing.T) {
	tm := NewComManager(nil)
	err := tm.Register(ComponentView{
		Name:      "test",
		Prototype: &Test{},
		Template: VNodeTemplate(
			VElem("span", NoAttr(), NoBind(), []VNode{
				VElem(CompInner, NoAttr(), NoBind(), []VNode{}),
			})),
	})

	if err != nil {
		panic(err)
	}

	_, ok := tm.GetComponent("div")
	require.Equal(t, ok, false)

	re := NodeRoot(
		VElem("test", map[string]interface{}{
			"str":  "Awesome!",
			"num":  "69",
			"fnum": "699.69",
			"tf":   "true",
		}, NoBind(), []VNode{
			VElem("smile", NoAttr(), NoBind(), []VNode{}),
			VText("_"),
			VElem("smile", NoAttr(), NoBind(), []VNode{}),
		}))
	cv, ok := tm.GetComponent("test")
	require.Equal(t, ok, true)

	ci, _ := cv.NewInstance(re)
	model := ci.Model().(*Test)
	require.Equal(t, model.Str, "Awesome!")
	require.Equal(t, model.Num, 69)
	require.Equal(t, model.Fnum, 699.69)
	require.Equal(t, model.Tf, true)

	ci.prepareInner(nil)
	require.Equal(t, re.Text(), ":D_:D")
}
