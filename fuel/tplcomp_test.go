package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/net/html"
)

type CompileTestSuite struct {
	suite.Suite
}

func fragmentFromString(htmlCode string) *html.Node {
	buf := bytes.NewBufferString(strings.TrimSpace(htmlCode))
	nodes, err := parseFragment(buf)

	if err != nil {
		panic(err)
	}

	return nodes[0]
}

func (s *CompileTestSuite) TestBasicTree() {
	root := generate(fragmentFromString(`
		<div>
			<div class="wrapper {{ this.aClass }}">
				<ul>
					<li>Prefix: {{ this.HeadItem }}</li>
					<li>2</li>
				</ul>
			</div>
		</div>
	`))

	s.Equal(root.typ, funcCallCodeNode)

	s.Equal(root.children[0].typ, stringCodeNode)
	s.Contains(root.children[0].code, "div")

	s.Equal(root.children[1].typ, nakedCodeNode)
	s.Equal(root.children[1].code, "nil")

	s.Equal(root.children[2].typ, compositeCodeNode)
	s.Contains(root.children[2].code, elementListOpener)

	rchild := root.domChildren()[0]
	s.Equal(rchild.typ, funcCallCodeNode)
	attrCode := rchild.children[1].children[0].code
	s.Contains(attrCode, `class`)
	s.Contains(attrCode, "`wrapper %v`, this.aClass)")

	// ul should contain 2 li
	ulChildren := rchild.domChildren()[0].domChildren()
	s.Len(ulChildren, 2)

	// "Prefix:" part
	liPrefix := ulChildren[0].domChildren()[0]
	s.Equal(liPrefix.typ, funcCallCodeNode)
	s.Equal(liPrefix.code, createTextNodeOpener)
	s.Equal(liPrefix.children[0].typ, stringCodeNode)
	s.Equal(liPrefix.children[0].code, "Prefix: ")

	// {{ this.HeadItem }} part
	liMustache := ulChildren[0].domChildren()[1]
	s.Equal(liMustache.children[0].typ, nakedCodeNode)
	s.Equal(liMustache.children[0].code, "this.HeadItem")
}

func TestCompile(t *testing.T) {
	suite.Run(t, new(CompileTestSuite))
}
