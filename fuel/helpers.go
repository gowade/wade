package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

func efmt(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func sfmt(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func execTplBuf(tpl *template.Template, data interface{}) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	err := tpl.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

// textPart represents either a typical HTML text node or a {{mustache node}}
type textPart struct {
	content    string
	isMustache bool
}

var (
	MustacheRegex = regexp.MustCompile("{{((?:[^{}]|{[^{]|}[^}])+)}}")
)

// parseTextMustache splits HTML text into a list of text and mustaches.
//
// "ABC: {{mustache}} DEF" would be splitted into
// "ABC: ", {{mustache}} and "DEF"
func parseTextMustache(text string) []textPart {
	matches := MustacheRegex.FindAllStringSubmatch(text, -1)

	if matches == nil {
		return []textPart{{text, false}}
	}

	parts := []textPart{}
	splitted := MustacheRegex.Split(text, -1)

	for i, m := range matches {
		if splitted[i] != "" {
			parts = append(parts, textPart{splitted[i], false})
		}

		parts = append(parts, textPart{strings.TrimSpace(m[1]), true})
	}

	if splitted[len(splitted)-1] != "" {
		parts = append(parts, textPart{splitted[len(splitted)-1], false})
	}

	return parts
}
