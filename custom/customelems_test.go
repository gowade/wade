package custom

import (
	"testing"

	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/dom/goquery"
	"github.com/stretchr/testify/require"
)

const (
	Real = `
	<test str="Awesome!" num="69" fnum="669.99" Tf="true"><smile></smile>_<smile></smile></test>
	`
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

func (t *Test) ProcessContents(ctl ContentsCtl) error {
	ctl.Contents().Filter("smile").ReplaceWith(ctl.Dom().NewFragment(":D"))
	return nil
}

func TestCustomTag(t *testing.T) {
	d := goquery.GetDom()
	tm := NewTagManager()
	err := tm.RegisterTags([]HtmlTag{HtmlTag{
		Name:      "test",
		Prototype: &Test{},
		Html:      `<span><wcontents></wcontents></span>`,
	}})

	tag, ok := tm.GetTag(d.NewFragment("<div></div>"))
	require.Equal(t, ok, false)

	re := d.NewFragment(Real)
	tag, ok = tm.GetTag(re)
	require.Equal(t, ok, true)

	elem, _ := tag.NewElem(re, nil)
	model := elem.Model().(*Test)
	require.Equal(t, model.Str, "Awesome!")
	require.Equal(t, model.Num, 69)
	require.Equal(t, model.Fnum, 669.99)
	require.Equal(t, model.Tf, true)

	err = elem.PrepareContents(func(s dom.Selection, once bool) {})
	if err != nil {
		panic(err)
	}
	require.Equal(t, re.Find("span").Text(), ":D_:D")
}
