package compiler

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"strings"

	"github.com/phaikawl/wade/core"
)

var (
	prelude = "package %v\n" +
		"import (\n" +
		"\t" + `. "fmt"` + "\n" +
		"\t" + `. "strings"` + "\n" +
		"\t" + `. "github.com/phaikawl/wade/utils"` + "\n" +
		"\t" + `wc "github.com/phaikawl/wade/core"` + "\n" +
		")\n\n" +
		"var binders = bdrs.Binders" +
		"var %v = wc.VPrep(wc.VNode%v)\n"
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

func (g *Compiler) Compile(htmlFile string, node *core.VNode) {
	g.writeFile(htmlFile, g.Process(node, 0, htmlFile))
}

func (g Compiler) fileContent(varName string, data string) string {
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

func (g *Compiler) writeFile(htmlFile string, data string) {
	filePath := path.Join(g.OutputDir, path.Base(htmlFile)+".go")
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

	if depth > 0 {
		if iSrc, ok := node.Attr("src"); ok {
			src := iSrc.(string)
			g.Compile(src, node)
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
							fmt.Printf(`Error '%v' while processing file "%v".`+"\n", err.Error(), file)
							continue
						}

						fn, ok := g.CompileTimeBinders[binder]
						if !ok {
							fmt.Printf(`Error 'No such binder "%v"' while processing file "%v".`+"\n", binder, file)
							continue
						}

						fStr = "func(__node *wc.VNode) {\n"
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
