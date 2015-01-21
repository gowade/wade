package compiler

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"strings"

	strutils "github.com/naoina/go-stringutil"

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
		"\t" + `. "github.com/phaikawl/wade/app/utils"` + "\n" +
		"\t" + `"github.com/phaikawl/wade/dom"` + "\n" +
		")\n\n" +
		"var %v = %v\n\n" +
		"func init() {_ = Url; _ = Join; _ = ToString; _ = Sprintf; _ = dom.DebugInfo}"
	mainVarName = "main"
)

type CTBFunc func(TempComplData, []string, string) string

type Compiler struct {
	OutputDir          string
	PackageName        string
	Includes           map[string]int
	IncludeIdx         int
	CompileTimeBinders map[string]CTBFunc
	PreventProcessing  map[*core.VNode]bool
	components         map[string]ComponentInfo
}

type ComponentInfo struct {
	defBinds map[string]string
	model    string
}

func NewCompiler(outputDir, pkgName string, ctBinders map[string]CTBFunc) *Compiler {
	return &Compiler{
		OutputDir:          outputDir,
		PackageName:        pkgName,
		CompileTimeBinders: ctBinders,
		Includes:           map[string]int{},
		PreventProcessing:  map[*core.VNode]bool{},
		components:         map[string]ComponentInfo{},
	}
}

func (g *Compiler) Compile(htmlFile string, varName string, node *core.VNode) {
	g.writeContent(htmlFile, varName, fmt.Sprintf("VPrep(&VNode%v)", g.Process(node, 0, htmlFile)))
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
				if !ok || comNameI.(string) == "" {
					continue
				}
				comName := strings.TrimSpace(comNameI.(string))

				modelNameI, ok := n.Attr("model")
				if !ok || modelNameI.(string) == "" {
					printErr(fmt.Sprintf(`No model specified for component "%v"`, comName), "root")
					continue
				}
				modelName := modelNameI.(string)

				outputFile := "component_" + comName + ".html.go"
				varName := "component_" + comName
				if impSrc, _ := n.Attr("import_src"); impSrc != nil {
					if impCom, _ := n.Attr("import_com"); impCom != nil {
						data := fmt.Sprintf("package %v\n"+`import __imported "%v"`+"\n%v = %v", g.PackageName, impSrc.(string),
							varName, "__imported."+tmplVarPrefix+"component_"+impCom.(string))
						g.writeFile(outputFile, data)
					} else {
						printErr(fmt.Sprintf("No import_com specified for component '%v' importing '%v'\n", comName, impSrc.(string)), "root")
					}
				} else {
					comTemp := core.VPrep(&core.VNode{
						Data: comName,
					})
					comTemp.Children = n.Children
					src := fmt.Sprintf("func(__m *%v) *VNode {\n\treturn VPrep(&VNode%v)\n}",
						modelName, g.Process(comTemp, 1, outputFile))
					g.writeContent(outputFile, varName, src)
				}

				m := map[string]string{}
				for attr, val := range n.Attrs {
					rname := []rune(attr)
					if rname[0] == '*' {
						field := string(rname[1:])
						m[field] = val.(string)
					}
				}

				g.components[comName] = ComponentInfo{
					defBinds: m,
					model:    modelName,
				}
			}
		}
	}

	g.Compile(htmlFile, mainVarName, node)
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
	Node     *core.VNode
	Depth    int
	Idt      string //indentation
	File     string
	Compiler *Compiler
}

func printErr(err string, file string) {
	fmt.Printf(`Error <%v> while processing "%v"`+"\n", err, file)
}

func (g *Compiler) bindCode(binds map[string]string, cplData TempComplData) (bStr string) {
	bStr = cplData.Idt + "Binds: []BindFunc{"
	for k, v := range binds {
		kr := []rune(k)

		if kr[0] == '@' || kr[0] == '#' {
			name := string(kr[1:])
			var fStr string
			switch kr[0] {
			case '@':
				fStr = fmt.Sprintf(`func(n *VNode){ n.Attrs["%v"] = %v },`, name, v)
			case '#':
				binder, args, err := parseBinderLHS(k)
				if err != nil {
					printErr(err.Error(), cplData.File)
					continue
				}

				fn, ok := g.CompileTimeBinders[binder]
				if !ok {
					printErr(fmt.Sprintf(`No such binder "%v"`, binder), cplData.File)
					continue
				}

				fStr = "func(__node *VNode) {\n"
				fStr += fn(cplData, args, v)
				fStr += "\n" + cplData.Idt + "\t" + "},"
			}
			bStr += "\n" + cplData.Idt + "\t" + fStr
		}
	}
	bStr += "\n" + cplData.Idt + "},\n"

	return
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
		if node.Data == "" {
			return ""
		} else if strings.TrimSpace(node.Data) == "" {
			node.Data = " "
		}

		return fmt.Sprintf("VText(`%s`)", node.Data)
	case core.MustacheNode:
		return fmt.Sprintf(`VMustache(func() interface{} { return %s })`,
			strings.TrimSpace(node.Data))
	case core.ElementNode, core.DeadNode, core.GroupNode:
		binds := map[string]string{}
		fieldBinds := map[string]string{}
		exBinds := map[string]string{}

		for k, v := range node.Attrs {
			kr := []rune(k)
			if kr[0] == '@' || kr[0] == '#' || kr[0] == '*' || kr[0] == '!' {
				delete(node.Attrs, k)
				expr := v.(string)
				binds[k] = expr
				name := string(kr[1:])
				switch kr[0] {
				case '@', '#':
					binds[k] = expr
				case '*':
					fieldBinds[name] = expr
				case '!':
					exBinds[name] = expr
				}
			}
		}

		cidt := g.getIndent(depth)
		idt := cidt + "\t"

		attrStr := ""

		if len(node.Attrs)+len(exBinds) != 0 {
			attrStr = idt + "Attrs: Attributes{"
			for k, v := range node.Attrs {
				attrStr += "\n" + idt + "\t" + fmt.Sprintf(`"%v": "%v",`, k, v.(string))
			}
			for k, v := range exBinds {
				attrStr += "\n" + idt + "\t" + fmt.Sprintf(`"%v": %v,`, k, v)
			}
			attrStr += "\n" + idt + "},\n"
		}

		childrenStr := ""
		if strings.HasPrefix(node.Data, core.ComponentTagPrefix) {
			comName := string([]rune(node.Data)[len(core.ComponentTagPrefix):])
			if comInfo, ok := g.components[comName]; ok {
				ret := "VComponent(func() (*VNode, func(*VNode)) {\n"
				ret += idt + fmt.Sprintf("\t\t__m := new(%v); __m.Init(); __node := %v\n",
					comInfo.model, tmplVarPrefix+"component_"+comName+"(__m)")
				ret += idt + "\t\treturn __node, func(_ *VNode) {\n"
				for k, v := range fieldBinds {
					ret += idt + fmt.Sprintf("\t\t\t__m.%v = %v\n", strutils.ToUpperCamelCase(k), v)
				}
				for k, v := range comInfo.defBinds {
					ret += idt + fmt.Sprintf("\t\t\t__m.%v = %v\n", strutils.ToUpperCamelCase(k), v)
				}

				ret += idt + fmt.Sprintf("\t\t\t__m.Update(__node)\n")
				ret += idt + "\t\t}\n"
				ret += idt + "\t})"

				return ret
			} else {
				printErr(fmt.Sprintf(`No component with name "%v" has been defined`, comName), file)
				return ""
			}
		}

		bStr := ""
		if len(binds) > 0 {
			bStr = g.bindCode(binds, TempComplData{
				Node:     node,
				Depth:    depth,
				Idt:      idt,
				File:     file,
				Compiler: g,
			})
		}

		if len(node.Children) > 0 {
			childrenStr = idt + "Children: []*VNode{"
			for _, c := range node.Children {
				cr := g.Process(c, depth+1, file)
				if cr == "" {
					continue
				}
				childrenStr += "\n" + idt + "\t" + cr + ","
			}
			childrenStr += "\n" + idt + "},\n"
		}

		typeStr := "ElementNode"
		switch node.Type {
		case core.GroupNode:
			typeStr = "GroupNode"
		case core.DeadNode:
			typeStr = "DeadNode"
		}

		return "{\n" +
			idt + fmt.Sprintf(`Data: "%v",`, node.Data) + "\n" +
			idt + fmt.Sprintf(`Type: %v,`, typeStr) + "\n" +
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
