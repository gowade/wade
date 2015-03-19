package main

const (
	StringCodeNode      = "string code"          // "<code>"
	NakedCodeNode       = "naked code"           // <code>
	FuncCallCodeNode    = "function call"        // funcName(<children>)
	CompositeCodeNode   = "composite literal"    // typeName{<children>}
	ElemListCodeNode    = "element list"         // []*Element{<children>}
	AppendListCodeNode  = "append list"          // for1 = append(for1, if1...)
	BlockCodeNode       = "function declaration" // func (m *A) funcName() {<children>}
	VarDeclAreaCodeNode = "var declaration area"
	SliceVarCodeNode    = "list variable" // a list of nodes represented as a variable
)

// codeNode is intermediate tree representation of the generated code
// this help decouple between the syntax and our code generation "business logic"
type codeNode struct {
	typ      string
	code     string
	children []*codeNode
}

func (n codeNode) dCh(i int) *codeNode {
	return n.children[2].children[i]
}

func (n codeNode) dChn() []*codeNode {
	return n.children[2].children
}
