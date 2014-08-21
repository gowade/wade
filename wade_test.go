package wade

import (
	"testing"
	"github.com/stretchr/testify/require"

	"github.com/gopherjs/gopherjs/js"
	"github.com/phaikawl/wade/dom/goquery"
	"github.com/phaikawl/wade/icommon"
	"github.com/phaikawl/wade/libs/http"
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

const (
	Src = `<div>
<wimport src="/a"></wimport>
<wimport src="/b"></wimport>
<div>
	<wimport src="/c"></wimport>
</div>
</div>`

	FailSrc = `<div><wimport src="/kdkfk"></wimport></div>`
	NoSrc   = `<div><wimport></wimport></div>`

	SrcA = `<wimport src="/d"></wimport>`
	SrcB = `b`
	SrcC = `c`
	SrcD = `a`
)

func TestHtmlImport(t *testing.T) {
	mb := http.NewMockBackend(map[string]http.TestResponse{
		"/a": http.FakeOK(SrcA),
		"/b": http.FakeOK(SrcB),
		"/c": http.FakeOK(SrcC),
		"/d": http.FakeOK(SrcD),
	})

	client := http.NewClient(mb)

	root := goquery.GetDom().NewFragment(Src)
	err := htmlImport(client, root, "/")
	require.Equal(t, err, nil)
	require.Equal(t, icommon.RemoveAllSpaces(root.Html()), `ab<div>c</div>`)

	root = root.NewFragment(FailSrc)
	err = htmlImport(client, root, "/")
	require.NotEqual(t, err, nil)

	root = root.NewFragment(NoSrc)
	err = htmlImport(client, root, "/")
	require.NotEqual(t, err, nil)
}
