package main

const (
	stringCodeNode    = "string code"          // "<code>"
	nakedCodeNode     = "naked code"           // <code>
	funcCallCodeNode  = "function call"        // funcName(<children>)
	compositeCodeNode = "composite literal"    // typeName{<children>}
	funcDeclCodeNode  = "function declaration" // func (m *A) funcName() {<children>}
)

// codeNode is intermediate tree representation of the generated code
// this help decouple between the syntax and our code generation "business logic"
type codeNode struct {
	typ      string
	code     string
	children []*codeNode
}

func (n codeNode) domChildren() []*codeNode {
	return n.children[2].children
}
