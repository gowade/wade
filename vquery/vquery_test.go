package vq

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/phaikawl/wade/dom/gonet"
)

func TestVQuery(t *testing.T) {
	gq := gonet.GetDom()
	vdom := gq.NewFragment(`<div>
		<div class="wrapper">
			<a href="zz">A</a>
			<div>
				<b id="b">B</b>
			</div>
		</div>
	</div>`).ToVNode()

	q := New(vdom.Ptr())

	a := q.Find(Selector{Class: "wrapper", Tag: "div"}).Find(Selector{Tag: "a"})[0]
	require.Equal(t, a.Text(), "A")
	a = q.Find(Selector{Attrs: M{"href": "zz"}})[0]
	require.Equal(t, a.Text(), "A")
	b := q.Find(Selector{Id: "b"})[0]
	require.Equal(t, b.Text(), "B")

	require.Equal(t, Parent(Parent(b)).HasClass("wrapper"), true)
}
