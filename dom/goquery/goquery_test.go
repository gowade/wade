package goquery

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	Awesome    = "Awesome!"
	Aw50m3n355 = "aw50m3n355"
	Aw         = "Aw50m3n355"
	E11t3      = "e11t3"
	N0rm41     = "n0rm41"
	HaiNT      = `<person class="` + N0rm41 + `">HaiNT</person>`
	P          = "<p>:D</p>"
)

func TestEverything(t *testing.T) {
	d := Dom{}
	s := d.NewFragment(`
	` + "<div><wade>" + Awesome + "</wade></div><div></div>")

	empty := d.NewFragment("")
	require.Equal(t, empty.Length(), 0)

	require.Equal(t, len(s.Elements()), 2)

	s = s.First()

	tag, err := s.TagName()
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, tag, "div")

	wade := s.Find("wade")
	require.Equal(t, wade.Html(), Awesome)

	haint := d.NewFragment(HaiNT)
	wade.ReplaceWith(haint)

	require.Equal(t, s.Html(), HaiNT)

	tf := d.NewFragment("<div>" + P + "</div>")
	p := tf.Find("p")
	s.Append(p)

	require.Equal(t, s.OuterHtml(), "<div>"+HaiNT+P+"</div>")
	require.Equal(t, tf.Html(), "")

	a, ok := haint.Attr(Aw50m3n355)
	require.Equal(t, ok, false)
	haint.SetAttr(Aw50m3n355, "5000")
	a, ok = haint.Attr(Aw)
	require.Equal(t, ok, true)
	require.Equal(t, a, "5000")
	haint.SetAttr(Aw50m3n355, "over 9000")
	a, _ = haint.Attr(Aw50m3n355)
	require.Equal(t, a, "over 9000")

	tn, _ := haint.Next().TagName()
	require.Equal(t, tn, "p")

	haint.Before(d.NewFragment(`<input value="NTH"></a>`))
	require.Equal(t, haint.Prev().Val(), "NTH")
	haint.After(d.NewFragment(`<input value="G"></input>`))
	require.Equal(t, haint.Next().Val(), "G")

	require.Equal(t, haint.HasClass(N0rm41), true)
	require.Equal(t, haint.HasClass(E11t3), false)
	haint.AddClass(E11t3)
	require.Equal(t, haint.HasClass(E11t3), true)
	haint.RemoveClass(N0rm41)
	require.Equal(t, haint.HasClass(N0rm41), false)
}
