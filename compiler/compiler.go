package compiler

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"strings"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/vquery"
)

const (
	tmplVarPrefix = "Tmpl_"
)

var (
	prelude = "package %v\n" +
		"import (\n" +
		"\t" + `. "fmt"` + "\n" +
		"\t" + `. "strings"` + "\n" +
		"\t" + `. "github.com/phaikawl/wade/utils"` + "\n" +
		"\t" + `. "github.com/phaikawl/wade/core"` + "\n" +
		")\n\n" +
		"var %v = VPrep(VNode%v)\n"
)

type CTBFunc func(TempComplData) string

type Compiler struct {
	OutputDir          string
	PackageName        string
	Includes           map[string]int
	IncludeIdx         int
	CompileTimeBinders map[string]CTBFunc
	PreventProcessing  map[*core.VNode]bool
}

func NewCompiler(outputDir, pkgName string, ctBinders map[string]CTBFunc) *Compiler {
	return &Compiler{
		OutputDir:          outputDir,
		PackageName:        pkgName,
		CompileTimeBinders: ctBinders,
		Includes:           map[string]int{},
		PreventProcessing:  map[*core.VNode]bool{},
	}
}

func (g *Compiler) Compile(htmlFile string, varName string, node *core.VNode) {
	g.writeContent(htmlFile, varName, g.Process(node, 0, htmlFile))
}

func (g *Compiler) CompileRoot(htmlFile string, node *core.VNode) {
	metaElems := vq.New(node).Find(vq.Selector{Tag: "w_meta"})
	if len(metaElems) == 0 {
		return
	}

	for _, meta := range metaElems {
		for _, n := range meta.Children {
			if n.Data == core.ComponentTagName {
				comNameI, ok := n.Attr("name")
				comName := strings.TrimSpace(comNameI.(string))
				if !ok || comName == "" {
					continue
				}

				outputFile := "component_" + comName + ".html.go"
				varName := "component_" + comName
				if impSrc, _ := n.Attr("import_src"); impSrc != nil {
					if impCom, _ := n.Attr("import_com"); impCom != nil {
						data := fmt.Sprintf("package %v\n"+`import __imported "%v"`+"\n%v = %v", g.PackageName, impSrc.(string),
							varName, "__imported."+tmplVarPrefix+"component_"+impCom.(string))
						g.writeFile(outputFile, data)
					} else {
						fmt.Printf("No import_com specified for component '%v' importing '%v'\n", comName, impSrc.(string))
					}
				} else {
					comTemp := core.VPrep(&core.VNode{
						Data: comName,
					})
					comTemp.Children = n.Children
					g.Compile(outputFile, varName, comTemp)
				}
			}
		}
	}

	g.Compile(htmlFile, "main", node)
}

func (g Compiler) fileContent(varName, data string) string {
	return fmt.Sprintf(prelude, g.PackageName, varName, data)
}

func (g *Compiler) getInclVar(htmlFile string) string {
	idx, ok := g.Includes[htmlFile]
	if !ok {
		g.IncludeIdx++
		idx = g.IncludeIdx
		g.Includes[htmlFile] = g.IncludeIdx
	}

	return fmt.Sprintf("include%d", idx)
}

func (g *Compiler) writeContent(htmlFile, varName, data string) {
	if varName == "" {
		varName = g.getInclVar(htmlFile)
	}

	g.writeFile(htmlFile, g.fileContent(tmplVarPrefix+varName, data))
}

func (g *Compiler) writeFile(htmlFile, content string) {
	filePath := path.Join(g.OutputDir, path.Base(htmlFile)+".go")
	ioutil.WriteFile(filePath, []byte(content), 0644)
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

type TempComplData struct {
	Args     []string
	Node     *core.VNode
	Depth    int
	Idt      string //indentation
	File     string
	Expr     string
	Compiler *Compiler
}

func (g *Compiler) Process(node *core.VNode, depth int, file string) string {
	if prevent, ok := g.PreventProcessing[node]; ok && prevent {
		return ""
	}

	if node.Data == "w_meta" {
		return ""
	}

	if depth > 0 && node.Data == core.GroupNodeTagName {
		if iSrc, ok := node.Attr("src"); ok {
			src := iSrc.(string)
			g.Compile(src, "", node)
			return tmplVarPrefix + g.getInclVar(src)
		}
	}

	switch node.Type {
	case core.TextNode:
		if strings.TrimSpace(node.Data) == "" {
			return ""
		}

		return fmt.Sprintf("VText(`%s`)", node.Data)
	case core.MustacheNode:
		return fmt.Sprintf(`VMustache(func() interface{} { return %s })`, strings.TrimSpace(node.Data))
	case core.ElementNode, core.DeadNode, core.GroupNode:
		cidt := g.getIndent(depth)
		idt := cidt + "\t"

		childrenStr := ""
		if len(node.Children) > 0 {
			childrenStr = idt + "Children: []VNode{"
			for _, c := range node.Children {
				cr := g.Process(c, depth+1, file)
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
			bStr = idt + "Binds: []BindFunc{"
			for k, v := range node.Attrs {
				kr := []rune(k)
				vstr := v.(string)

				if kr[0] == '@' || kr[0] == '#' || kr[0] == '*' {
					delete(node.Attrs, k)
					name := string(kr[1:])
					var fStr string
					switch kr[0] {
					case '@':
						fStr = fmt.Sprintf(`func(n *VNode){ n.Attrs["%v"] = %v }`, name, vstr)
					case '#':
						binder, args, err := parseBinderLHS(k)
						if err != nil {
							fmt.Printf(`Error '%v' while processing file "%v".`+"\n", err.Error(), file)
							continue
						}

						fn, ok := g.CompileTimeBinders[binder]
						if !ok {
							fmt.Printf(`Error 'No such binder "%v"' while processing file "%v".`+"\n", binder, file)
							continue
						}

						fStr = "func(__node *VNode) {\n"
						fStr += fn(TempComplData{
							Args:     args,
							Node:     node,
							Depth:    depth,
							Idt:      idt,
							File:     file,
							Expr:     vstr,
							Compiler: g,
						})
						fStr += "\n" + idt + "\t" + "},"
					}
					bStr += "\n" + idt + "\t" + fStr
				}
			}
			bStr += "\n" + idt + "},\n"
		}

		if len(node.Attrs) != 0 {
			attrStr = idt + "Attrs: Attributes{"
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

func (g Compiler) getIndent(depth int) (s string) {
	for i := 0; i < depth; i++ {
		s += "\t\t"
	}

	return
}

func (g Compiler) format(lines []string, depth int) (s string) {
	for i := range lines {
		s += g.getIndent(depth) + lines[i]
		if i < len(lines)-1 {
			s += "\n"
		}
	}

	return
}
