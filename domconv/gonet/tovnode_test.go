package gonet

import (
	"bytes"
	"testing"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/utils"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestConversion(t *testing.T) {
	// From virtual to real
	node := createElement("zz")

	root := core.VPrep(core.VNode{
		Data: "div",
		Children: []core.VNode{
			core.VNode{
				Data: "b",
			},
			core.VNode{
				Data:  "a",
				Attrs: core.Attributes{"w": 123.4},
			},
			core.VNode{
				Type: core.GroupNode,
				Children: []core.VNode{
					core.VText("("),
					core.VMustache("empty"),
					{
						Data:     "div",
						Attrs:    core.Attributes{"disabled": true},
						Binds:    []core.Bindage{core.BindAttr("test", "test")},
						Children: []core.VNode{core.VText(")")},
					},
				},
			},
			core.VNode{
				Type: core.DataNode,
				Data: "data",
			},
			core.VNode{
				Type: core.DeadNode,
			},
			core.VText("t"),
		},
	}).Ptr()

	buf := bytes.NewBufferString("")
	Render(node, root)
	html.Render(buf, node)
	src := `<div>
			<b></b>
			<a w="123.4"></a>
			(<div disabled="">
				)
			</div>
			<!--data-->
			t
		</div>
	`
	require.Equal(t, utils.NoSp(buf.String()), utils.NoSp(src))

	// From real to virtual
	pnode, err := parseHtml(`
		<div>
			<div !group>
				<b></b>
				<a w="123.4"></a>
				(<div disabled>)</div>
				<!--data-->
				t
			</div>
		</div>
	`)

	if err != nil {
		t.Fatal(err)
	}

	vnode := ToVNode(pnode)
	target := createElement("zz")
	Render(target, &vnode)

	b := bytes.NewBufferString("")
	html.Render(b, target)
	require.Equal(t, utils.NoSp(b.String()), utils.NoSp(src))

	vnode.ChildElems()[0].ChildElems()[0].SetClass("done", true)
	Render(target, &vnode)
	b = bytes.NewBufferString("")
	html.Render(b, target)
	src2 := `<div>
			<b class="done"></b>
			<a w="123.4"></a>
			(<div disabled="">
				)
			</div>
			<!--data-->
			t
		</div>`

	require.Equal(t, utils.NoSp(b.String()), utils.NoSp(src2))
}
