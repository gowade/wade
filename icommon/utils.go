package icommon

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/phaikawl/wade/dom"
)

const WrapperTag = "ww"

func WrapperUnwrap(elem dom.Selection) {
	for _, e := range elem.Find(WrapperTag).Elements() {
		e.Unwrap()
	}
}

func IsWrapperElem(elem dom.Selection) bool {
	return elem.Is(WrapperTag)
}

func RemoveAllSpaces(src string) string {
	r := ""
	for _, c := range src {
		if !unicode.IsSpace(c) {
			r += string(c)
		}
	}

	return r
}

var (
	TempReplaceRegexp = regexp.MustCompile(`<%([^"<>]+)%>`)
)

// parseTemplate replaces "<% bindstr %>" with <span bind-html="bindstr"></span>
func ParseTemplate(source string) string {
	return TempReplaceRegexp.ReplaceAllStringFunc(source, func(m string) string {
		bindstr := strings.TrimSpace(TempReplaceRegexp.FindStringSubmatch(m)[1])
		return fmt.Sprintf(`<span bind-html="%v"></span>`, bindstr)
	})
}
