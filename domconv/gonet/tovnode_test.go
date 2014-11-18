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
	root := core.VWrap("div", []core.VNode{
		core.VElem("a", map[string]interface{}{
			"w": 123.4,
		}, core.NoBind(), []core.VNode{}),
		core.V(core.GhostNode, "span", core.NoAttr(), core.NoBind(), []core.VNode{
			core.VText("("),
			core.VMustache("empty"),
			core.VElem("div", map[string]interface{}{
				"disabled": true,
			}, []core.Bindage{core.BindAttr("test", "test")}, []core.VNode{
				core.VText(")"),
			}),
		}),
		core.V(core.DataNode, "data", core.NoAttr(), core.NoBind(), []core.VNode{}),
		core.V(core.DeadNode, "dead", core.NoAttr(), core.NoBind(), []core.VNode{}),
		core.VText("t"),
	})

	conv := Converter{node}
	buf := bytes.NewBufferString("")
	conv.Render(root)
	html.Render(buf, node)
	src := `<div>
			<a w="123.4"></a>
			(<div disabled="">
				)
			</div>
			<!--data-->
			t
		</div>
	`
	require.Equal(t, utils.WithoutSpaces(buf.String()), utils.WithoutSpaces(src))

	// From real to virtual
	pnode, err := parseHtml(`
		<div>
			<div !ghost>
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

	vnode := Converter{pnode}.ToVNode()
	conv.Render(vnode)
	require.Equal(t, utils.WithoutSpaces(buf.String()), utils.WithoutSpaces(src))
}
