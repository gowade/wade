package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/dom/gonet"
)

var (
	Prelude = "package %v\n" +
		"import (\n" +
		"\t " + `. "github.com/phaikawl/wade/app/helpers"` + "\n" +
		"\t " + `wc "github.com/phaikawl/wade/core"` + "\n" +
		")\n\n" +
		"var %v = wc.VPrep(wc.VNode%v)\n"
)

var (
	gd           = gonet.GetDom()
	groupNodeStr = fmt.Sprintf("<%v></%v>", core.GroupNodeTagName)
)

func parseHTML(filePath string) dom.Selection {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	file, err := os.Open(path.Join(wd, filePath))
	if err != nil {
		panic(err)
	}

	return gonet.NewFragment(file)
}

func importHtml(node dom.Selection) {
	for _, inclNode := range node.Find(core.IncludeTagName).Elements() {
		src, ok := inclNode.Attr("src")
		if !ok {
			continue
		}

		repl := gd.NewFragment(groupNodeStr)
		repl.SetAttr("src", src)

		if belongStr, hasBelong := inclNode.Attr("_belong"); hasBelong {
			repl.SetAttr("_belong", belongStr)
		}

		repl.Append(parseHTML(src))
		importHtml(repl)
		inclNode.ReplaceWith(repl)
	}
}

type generator struct {
	outputDir   string
	packageName string
	includes    map[string]int
	includeIdx  int
}

func (g *generator) generate(htmlFile string, node *core.VNode) {
	g.writeFile(htmlFile, g.processRec(node, 0))
}

func (g generator) fileContent(varName string, data string) string {
	return fmt.Sprintf(Prelude, g.packageName, varName, data)
}

func (g *generator) getInclVar(htmlFile string) string {
	idx, ok := g.includes[htmlFile]
	if !ok {
		g.includeIdx++
		idx = g.includeIdx
		g.includes[htmlFile] = g.includeIdx
	}

	return fmt.Sprintf("include%d", idx)
}

func (g *generator) writeFile(htmlFile string, data string) {
	filePath := path.Join(g.outputDir, path.Base(htmlFile)+".go")
	ioutil.WriteFile(filePath, []byte(g.fileContent(g.getInclVar(htmlFile), data)), 0644)
}

func (g *generator) processRec(node *core.VNode, depth int) string {
	if depth > 0 {
		if iSrc, ok := node.Attr("src"); ok {
			src := iSrc.(string)
			g.generate(src, node)
			return g.getInclVar(src)
		}
	}

	switch node.Type {
	case core.TextNode:
		if strings.TrimSpace(node.Data) == "" {
			return ""
		}

		return fmt.Sprintf("wc.VText(`%s`)", node.Data)
	case core.MustacheNode:
		return fmt.Sprintf(`wc.VMustache(func() interface{} { return %s })`, node.Data)
	case core.ElementNode, core.DeadNode, core.GroupNode:
		if len(node.Children) == 0 {
			if len(node.Attrs) == 0 {
				return fmt.Sprintf(`{Data: "%v"}`, node.Data)
			}

			if _, ok := node.Attrs["class"]; len(node.Attrs) == 1 && ok {
				return fmt.Sprintf(`VElem("%v", "%v")`, node.Data, node.ClassStr())
			}
		}

		cidt := g.getIndent(depth)
		idt := cidt + "\t"

		childrenStr := ""
		if len(node.Children) > 0 {
			childrenStr = idt + "Children: []wc.VNode{"
			for _, c := range node.Children {
				cr := g.processRec(c, depth+1)
				if cr == "" {
					continue
				}
				childrenStr += "\n" + idt + "\t" + cr + ","
			}
			childrenStr += "\n" + idt + "},\n"
		}

		attrStr := ""
		if len(node.Attrs) != 0 {
			attrStr = idt + "Attrs: wc.Attributes{"
			for k, v := range node.Attrs {
				attrStr += "\n" + idt + "\t" + fmt.Sprintf(`"%v": "%v",`, k, v.(string))
			}
			attrStr += idt + "\n" + idt + "},\n"
		}

		return "{\n" +
			idt + fmt.Sprintf(`Data: "%v",`, node.Data) + "\n" +
			attrStr +
			childrenStr +
			cidt + "}"
	}

	panic(fmt.Sprintf("Unhandled node type %v!", node.Type))
	return ""
}

func (g generator) getIndent(depth int) (s string) {
	for i := 0; i < depth; i++ {
		s += "\t\t"
	}

	return
}

func (g generator) format(lines []string, depth int) (s string) {
	for i := range lines {
		s += g.getIndent(depth) + lines[i]
		if i < len(lines)-1 {
			s += "\n"
		}
	}

	return
}

func main() {
	var (
		flagMasterFile = flag.String("f", "public/main.html", "main template file")
		flagOutputDir  = flag.String("o", "client", "output directory")

		masterFile = *flagMasterFile
		outputDir  = *flagOutputDir
	)

	root := gd.NewFragment(groupNodeStr)
	root.Append(parseHTML(masterFile))

	importHtml(root)
	pkgName := path.Base(outputDir)

	vRoot := root.ToVNode()
	gen := &generator{outputDir, pkgName, make(map[string]int), 0}
	gen.generate(masterFile, vRoot)
}
