package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVNode(t *testing.T) {
	vn := VWrap("div", []VNode{
		VWrap("div", []VNode{
			VText("ABCD"),
		}),
		VWrap("div", []VNode{
			VWrap("hidden", []VNode{VText("<This should not display>")}),
			VText("hidden"),
			VText("EFGH"),
		}),
	})

	nn := vn.CloneWithCond(func(node VNode) bool {
		if node.Data == "hidden" {
			return false
		}

		return true
	})

	NodeWalk(&nn, func(n *VNode) {
		require.NotEqual(t, n.Type, UnsetNode)
	})

	require.Equal(t, nn.Text(), "ABCDEFGH")
}
