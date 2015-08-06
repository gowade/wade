package vdom

func Debug(node Node) {
	debug(">>>", node, 0)
}

func debug(prefix string, node Node, depth int) {
	var sp string
	for i := 0; i < depth; i++ {
		sp += "  "
	}

	if e, ok := node.(*Element); ok {
		println(prefix+sp+e.Tag, e.Attrs["class"], e.Attrs["id"], e.Attrs)
		for _, c := range e.Children {
			debug("", c, depth+1)
		}
	} else {
		println(sp + `"` + node.(*TextNode).Data + `"`)
	}
}
