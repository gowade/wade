package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/dom"
	"github.com/phaikawl/wade/dom/gonet"
)

var (
	Prelude = "package %v\n" +
		"import (\n" +
		"\t " + `. "github.com/phaikawl/wade/app/helpers"` + "\n" +
		"\t " + `bdrs "github.com/phaikawl/wade/binders"` + "\n" +
		"\t " + `wc "github.com/phaikawl/wade/core"` + "\n" +
		")\n\n" +
		"var binders = bdrs.Binders" +
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
	g.writeFile(htmlFile, g.processRec(node, 0, htmlFile))
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

var (
	NameRegexp = regexp.MustCompile(`\w+`)
)

func checkName(strs []string) error {
	for _, str := range strs {
		if !NameRegexp.MatchString(str) {
			return fmt.Errorf("Invalid name %v", str)
		}
	}
	return nil
}

func parseBinderLHS(astr string) (binder string, args []string, err error) {
	lp := strings.IndexRune(astr, '(')
	if lp != -1 {
		if astr[len(astr)-1] != ')' {
			err = fmt.Errorf("Invalid syntax for left hand side binding `%v`", astr)
			return
		}
		binder = astr[:lp]
		args = strings.Split(astr[lp+1:len(astr)-1], ",")
	} else {
		binder = astr
		args = []string{}
	}

	binder = binder[1:]

	err = checkName(append(args, binder))
	return
}

func (g *generator) processRec(node *core.VNode, depth int, file string) string {
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
		return fmt.Sprintf(`wc.VMustache(func() interface{} { return %s })`, strings.TrimSpace(node.Data))
	case core.ElementNode, core.DeadNode, core.GroupNode:
		cidt := g.getIndent(depth)
		idt := cidt + "\t"

		childrenStr := ""
		if len(node.Children) > 0 {
			childrenStr = idt + "Children: []wc.VNode{"
			for _, c := range node.Children {
				cr := g.processRec(c, depth+1, file)
				if cr == "" {
					continue
				}
				childrenStr += "\n" + idt + "\t" + cr + ","
			}
			childrenStr += "\n" + idt + "},\n"
		}

		attrStr := ""
		bStr := ""
		nbinds := 0

		for k := range node.Attrs {
			kr := []rune(k)
			if kr[0] == '@' || kr[0] == '#' || kr[0] == '*' {
				nbinds++
			}
		}

		if nbinds > 0 {
			bStr = idt + "Binds: []wc.BindFunc{"
			for k, v := range node.Attrs {
				kr := []rune(k)
				vstr := v.(string)

				if kr[0] == '@' || kr[0] == '#' || kr[0] == '*' {
					delete(node.Attrs, k)
					name := string(kr[1:])
					var fStr string
					switch kr[0] {
					case '@':
						fStr = fmt.Sprintf(`func(n *wc.VNode){ n.Attrs["%v"] = %v }`, name, vstr)
					case '#':
						binder, args, err := parseBinderLHS(k)
						if err != nil {
							panic(fmt.Errorf(`Error "%v" while processing file "%v".`, err.Error(), file))
						}

						fStr = "func() {"
						switch binder {
						case "for":
							key, val := "_", "_"
							if len(args) >= 1 {
								key = args[0]
							}

							if len(args) >= 2 {
								val = args[1]
							}

							fStr += "\n" + idt + fmt.Sprintf("\t\tfor __index, %v := range %v {\n", val, vstr)
							if key != "_" {
								fStr += idt + fmt.Sprintf("\t\t\t%v := __index\n", key)
							}
							fStr += idt + fmt.Sprintf("\t\t\t%v[__index] = %v", vstr, g.processRec(node.ChildElems()[0], depth+2, file))
							fStr += "\n" + idt + "\t\t}"
						}
						fStr += "\n" + idt + "\t" + "}"
					}
					bStr += "\n" + idt + "\t" + fStr
				}
			}
			bStr += "\n" + idt + "},\n"
		}

		if len(node.Attrs) != 0 {
			attrStr = idt + "Attrs: wc.Attributes{"
			for k, v := range node.Attrs {
				attrStr += "\n" + idt + "\t" + fmt.Sprintf(`"%v": "%v",`, k, v.(string))
			}
			attrStr += "\n" + idt + "},\n"
		}

		return "{\n" +
			idt + fmt.Sprintf(`Data: "%v",`, node.Data) + "\n" +
			bStr +
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
