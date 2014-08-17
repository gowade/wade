package goquery

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	Awesome = "Awesome!"
	HaiNT   = "<chmk>HaiNT</chmk>"
	P       = "<p>:D</p>"
)

func TestEverything(t *testing.T) {
	d := Dom{}
	s := d.NewFragment("<div><wade>" + Awesome + "</wade></div>")

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
}
