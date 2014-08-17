package wade

import (
	"testing"
	"unicode"

	"github.com/gopherjs/gopherjs/js"
	"github.com/phaikawl/wade/dom/goquery"
	"github.com/phaikawl/wade/libs/http"
	"github.com/stretchr/testify/require"
)

func init() {
	stubInit()
}

type JsStub struct{}

func (j *JsStub) Get(name string) js.Object {
	return j
}
func (j *JsStub) Set(name string, value interface{}) {
}
func (j *JsStub) Delete(name string) {
}
func (j *JsStub) Length() int {
	return 0
}
func (j *JsStub) Index(i int) js.Object {
	return j
}
func (j *JsStub) SetIndex(i int, value interface{}) {}
func (j *JsStub) Call(name string, args ...interface{}) js.Object {
	return j
}
func (j *JsStub) Invoke(args ...interface{}) js.Object {
	return j
}
func (j *JsStub) New(args ...interface{}) js.Object {
	return j
}
func (j *JsStub) Bool() bool {
	return true
}
func (j *JsStub) Str() string {
	return ""
}
func (j *JsStub) Int() int {
	return 0
}
func (j *JsStub) Int64() int64 {
	return 0
}
func (j *JsStub) Uint64() uint64 {
	return 0
}
func (j *JsStub) Float() float64 {
	return 0
}
func (j *JsStub) Interface() interface{} {
	return j
}
func (j *JsStub) Unsafe() uintptr {
	return uintptr(0)
}
func (j *JsStub) IsUndefined() bool {
	return true
}
func (j *JsStub) IsNull() bool {
	return true
}

func stubInit() {
	js.Global = &JsStub{}
}

func TestPageUrl(t *testing.T) {
	pm := pageManager{}
	pm.displayScopes = make(map[string]displayScope)
	route := "/:testparam/:testparam2/*testparam3"
	pm.registerDisplayScopes(map[string]DisplayScope{
		"test": MakePage(route, ""),
	})

	var u string
	var err error
	u, err = pm.PageUrl("test", 12, "abc", "some.go")
	expected := "/12/abc/some.go"
	if err != nil {
		t.Fatalf(err.Error())
	}

	require.Equal(t, u, expected)

	u, err = pm.PageUrl("test", 12, "abc")
	if err == nil {
		t.Fatalf("It should have raised an error for not having enough parameters.")
	}

	u, err = pm.PageUrl("test", 12, "abc", "zz", 22)
	if err == nil {
		t.Fatalf("It should have raised an error for having too many parameters.")
	}
}

const (
	Src = `<div>
<wimport src="/a"></wimport>
<wimport src="/b"></wimport>
<div>
	<wimport src="/c"></wimport>
</div>
</div>`

	SrcA = `<wimport src="/d"></wimport>`
	SrcB = `b`
	SrcC = `c`
	SrcD = `a`
)

func removeSpace(src string) string {
	r := ""
	for _, c := range src {
		if !unicode.IsSpace(c) {
			r += string(c)
		}
	}

	return r
}

func TestHtmlImport(t *testing.T) {
	mb := http.NewMockBackend(map[string]http.TestResponse{
		"/a": http.FakeOK(SrcA),
		"/b": http.FakeOK(SrcB),
		"/c": http.FakeOK(SrcC),
		"/d": http.FakeOK(SrcD),
	})

	root := goquery.GetDom().NewFragment(Src)
	htmlImport(http.NewClient(mb), root, "/")
	require.Equal(t, removeSpace(root.Html()), `ab<div>c</div>`)
}
