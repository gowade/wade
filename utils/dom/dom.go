package dom

var (
	document Node
)

func Document() Node {
	if document == nil {
		panic("DOM document has not been set.")
	}
	return document
}

func SetDocument(node Node) {
	document = node
}

type NodeType int

const (
	NopNode NodeType = iota
	ElementNode
	TextNode
)

type Node interface {
	Type() NodeType
	Find(query string) []Node
	Data() string
	Children() []Node
}
