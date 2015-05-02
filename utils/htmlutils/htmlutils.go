package htmlutils

import (
	"bytes"
	"io"
	"strings"

	"github.com/gowade/html"
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

func RemoveGarbageTextChildren(node *html.Node) {
	prev := node.FirstChild
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode && strings.TrimSpace(c.Data) == "" {
			if c == node.FirstChild {
				node.FirstChild = c.NextSibling
				prev = node.FirstChild
			} else {
				prev.NextSibling = c.NextSibling
				if c == node.LastChild {
					node.LastChild = prev
				}
			}
		} else {
			prev = c
		}
	}
}

func Render(node *html.Node) string {
	var buf bytes.Buffer
	html.Render(&buf, node)
	return buf.String()
}
