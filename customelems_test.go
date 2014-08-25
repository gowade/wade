package wade

import (
	"strings"
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
		Str  string
		Num  int
		Fnum float32
		Tf   bool
	}
)

func (t *Test) Init(ce CustomElem) error {
	ce.Contents.Find("smile").ReplaceWith(ce.Dom.NewFragment(":D"))
	return nil
}

func TestCustomTag(t *testing.T) {
	d := goquery.GetDom()
	tm := newCustagMan()
	err := tm.registerTags([]CustomTag{CustomTag{
		Name:       "testfail",
		Attributes: []string{"Id", "Gender"},
		Prototype:  BaseProto{},
		Html:       ``,
	}, CustomTag{
		Name:       "test",
		Attributes: []string{"Str", "Num", "Fnum", "Tf"},
		Prototype:  &Test{},
		Html:       `<span><wcontents></wcontents></span>`,
	}})

	require.NotEqual(t, err, nil)
	require.Equal(t, strings.Contains(err.Error(), "forbidden"), true)

	tag, ok := tm.GetCustomTag(d.NewFragment("<div></div>"))
	require.Equal(t, ok, false)

	re := d.NewFragment(Real)
	tag, ok = tm.GetCustomTag(re)
	require.Equal(t, ok, true)

	model := tag.NewModel(re).(*Test)
	require.Equal(t, model.Str, "Awesome!")
	require.Equal(t, model.Num, 69)
	require.Equal(t, model.Fnum, 669.99)
	require.Equal(t, model.Tf, true)

	tag.PrepareTagContents(re, model, func(s dom.Selection) {})
	require.Equal(t, re.Find("span").Text(), ":D_:D")
}
