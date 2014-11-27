package core

import (
	"testing"

	"github.com/phaikawl/wade/scope"
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
	tm := NewComManager()
	err := tm.Register(ComponentView{
		Name:      "test",
		Prototype: &Test{},
		Template: VPrep(VNode{
			Data:     "span",
			Children: []VNode{{Data: CompInner}},
		}),
	})

	if err != nil {
		panic(err)
	}

	_, ok := tm.GetComponent("div")
	require.Equal(t, ok, false)

	re := VPrep(VNode{
		Data: "test",
		Attrs: Attributes{
			"str":  "Awesome!",
			"num":  "69",
			"fnum": "699.69",
			"tf":   "true",
		},
		Children: []VNode{
			{Data: "smile"},
			VText("_"),
			{Data: "smile"},
		},
	}).Ptr()

	cv, ok := tm.GetComponent("test")
	require.Equal(t, ok, true)

	ci, _ := cv.NewInstance(re)
	model := ci.Model().(*Test)
	require.Equal(t, model.Str, "Awesome!")
	require.Equal(t, model.Num, 69)
	require.Equal(t, model.Fnum, 699.69)
	require.Equal(t, model.Tf, true)

	ci.prepareInner(scope.NewScope())
	require.Equal(t, re.Text(), ":D_:D")
}
