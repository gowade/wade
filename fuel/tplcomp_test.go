package main

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/gowade/wade/utils/htmlutils"
)

const (
	RETURN_START = 2
)

type CompileTestSuite struct {
	suite.Suite
}

func (s *CompileTestSuite) TestBasicTree() {
	root := generate(htmlutils.FragmentFromString(`
		<div>
			<div class="wrapper {{ this.aClass }}">
				<ul><li>Prefix: {{ this.HeadItem }}</li><li>Second</li></ul>
			</div>
		</div>
	`))

	root = root.children[RETURN_START]
	s.Equal(root.typ, FuncCallCodeNode)

	s.Equal(root.children[0].typ, StringCodeNode)
	s.Contains(root.children[0].code, "div")

	s.Equal(root.children[1].typ, NakedCodeNode)
	s.Equal(root.children[1].code, "nil")

	s.Equal(root.children[2].typ, ElemListCodeNode)

	rchild := root.dCh(0)
	s.Equal(rchild.typ, FuncCallCodeNode)
	attrCode := rchild.children[1].children[0].code
	s.Contains(attrCode, `class`)
	s.Contains(attrCode, "`wrapper %v`, this.aClass)")

	// ul should contain 2 li
	ulChildren := rchild.dCh(0).dChn()
	s.Len(ulChildren, 2)
	s.Equal(ulChildren[1].typ, FuncCallCodeNode)
	s.Equal(ulChildren[1].dCh(0).children[0].code, "Second")

	// "Prefix:" part
	liPrefix := ulChildren[0].dCh(0)
	s.Equal(liPrefix.typ, FuncCallCodeNode)
	s.Equal(liPrefix.code, CreateTextNodeOpener)
	s.Equal(liPrefix.children[0].typ, StringCodeNode)
	s.Equal(liPrefix.children[0].code, "Prefix: ")

	// {{ this.HeadItem }} part
	liMustache := ulChildren[0].dCh(1)
	s.Equal(liMustache.children[0].typ, NakedCodeNode)
	s.Equal(liMustache.children[0].code, "this.HeadItem")
}

func (s *CompileTestSuite) TestForAndIf() {
	root := generate(htmlutils.FragmentFromString(`
		<ul>
			<for k="i" v="item" range="{{ this.Items }}">
				<if cond="{{ i == 0 }}">
					<li>Even {{ i }}</li>
				</if>
				<li>{{ v }}</li>	
			</for>
		</ul>
	`))

	varDecls := root.children[0]

	// check for loop declaration
	s.Equal(varDecls.typ, VarDeclAreaCodeNode)
	s.Equal(varDecls.children[0].typ, NakedCodeNode)
	s.Equal(varDecls.children[1].typ, BlockCodeNode)
	s.Contains(varDecls.children[1].code, "for")

	// Check if inside for
	s.Equal(varDecls.children[1].children[1].typ, VarDeclAreaCodeNode)
	s.Equal(varDecls.children[1].children[1].children[1].typ, BlockCodeNode)
	s.Contains(varDecls.children[1].children[1].children[1].code, "if")

	s.Equal(root.children[RETURN_START].dCh(0).typ, SliceVarCodeNode)
}

func TestCompile(t *testing.T) {
	suite.Run(t, new(CompileTestSuite))
}
