package vdom

type TreeModifier interface {
	Render(Node, DomNode)
	Insert(Node, DomNode)
	Delete(DomNode)
}

type DomNode interface {
	Children() []DomNode
	AppendChild(DomNode)
	Remove()
}

func nodeCompat(a, b Node) bool {
	aie, bie := a.IsElement(), b.IsElement()
	if aie != bie {
		return false
	}

	if aie {
		return a.(*Element).Tag == b.(*Element).Tag
	}

	return a.(*TextNode).Data == b.(*TextNode).Data
}

func PerformDiff(a, b *Element, dNode DomNode, m TreeModifier) {
	if b == nil || a.Tag != b.Tag {
		m.Render(a, dNode)
		return
	}

	dChildren := dNode.Children()

	i := 0
	for ; i < len(a.Children); i++ {
		aCh := a.Children[i]
		if i > len(b.Children)-1 {
			m.Insert(aCh, dNode)
			continue
		}

		bCh := b.Children[i]
		dCh := dChildren[i]
		if nodeCompat(aCh, bCh) && a.IsElement() {
			PerformDiff(aCh.(*Element), bCh.(*Element), dCh, m)
		} else {
			m.Render(aCh, dCh)
		}
	}

	//println(a.Tag, len(a.Children), b.Tag, len(b.Children), i)
	for ; i < len(b.Children); i++ {
		m.Delete(dChildren[i])
	}
}
