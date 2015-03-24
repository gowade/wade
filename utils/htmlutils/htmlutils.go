package htmlutils

import (
	"bytes"
	"io"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func ParseFragment(source io.Reader) ([]*html.Node, error) {
	return html.ParseFragment(source, &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	})
}

func FragmentFromString(htmlCode string) *html.Node {
	buf := bytes.NewBufferString(strings.TrimSpace(htmlCode))
	nodes, err := ParseFragment(buf)

	if err != nil {
		panic(err)
	}

	return nodes[0]
}

func Render(node *html.Node) string {
	var buf bytes.Buffer
	html.Render(&buf, node)
	return buf.String()
}
