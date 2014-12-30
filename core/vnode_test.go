package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVNode(t *testing.T) {
	vn := VNode{
		Data: "div",
		Children: []*VNode{
			{
				Data:     "div",
				Children: []*VNode{VText("ABCD")},
			},
			{
				Type: GroupNode,
				Children: []*VNode{
					{Data: "hidden", Children: []*VNode{VText("<Should not display>")}},
					VText("hidden"),
					VText("EFGH"),
				},
			},
		},
	}

	nn := vn.CloneWithCond(func(node VNode) bool {
		if node.Data == "hidden" {
			return false
		}

		return true
	})

	NodeWalk(nn, func(n *VNode) {
		require.NotEqual(t, n.Type, NotsetNode)
	})

	require.Equal(t, nn.Text(), "ABCDEFGH")
}
