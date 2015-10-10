package htmlutils

import (
	"bytes"
	"strings"

	html "github.com/gowade/whtml"
)

func FragmentFromString(htmlCode string) *html.Node {
	buf := bytes.NewBufferString(strings.TrimSpace(htmlCode))
	nodes, err := html.Parse(buf)

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
	node.Render(&buf)
	return buf.String()
}
