package compiler

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"strings"

	strutils "github.com/naoina/go-stringutil"

	"github.com/phaikawl/wade/core"
	"github.com/phaikawl/wade/page"
	"github.com/phaikawl/wade/vquery"
)

const (
	ComponentDeclTagName = "component"
	HanldeTagName        = "handle"
	RouterDeclTagName    = "router"
	PageDeclTagName      = "page"
	tmplVarPrefix        = "Tmpl_"
	OriginFileAttrName   = "_orig"
	PGroupDeclTagName    = "pagegroup"
)

var (
	prelude = "package %v\n" +
		"import (\n" +
		"\t" + `. "fmt"` + "\n" +
		"\t" + `"strings"` + "\n" +
		"\t" + `. "github.com/phaikawl/wade/utils"` + "\n" +
		"\t" + `. "github.com/phaikawl/wade/core"` + "\n" +
		"\t" + `. "github.com/phaikawl/wade/rt/utils"` + "\n" +
		"\t" + `. "github.com/phaikawl/wade/rtbinders"` + "\n" +
		"\t" + `"github.com/phaikawl/wade/dom"` + "\n" +
		")\n\n%v\n\n" +
		"func init() {_ = Url; _ = strings.Join; _ = ToString; _ = Sprintf; _ = dom.DebugInfo; _ = RTBinder_value}"
	mainVarName   = "main"
	fileOf        = map[*core.VNode]string{}
	displayScopes = map[string]page.DisplayScope{}
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

func (g *Compiler) genPageCode(page, vm string, n *core.VNode, rootHtmlFile string) {
	funcEm := ""
	namePrefix := page
	if vm != "" {
		funcEm = "(" + vm + ")"
		namePrefix = ""
	}
	g.writeTemplate(page+".html", fmt.Sprintf(`
func %v %vTemplate() *VNode {
return VPrep(&VNode%v)
}
`, funcEm, namePrefix, g.Process(n, 0, rootHtmlFile)))
}

func pageMarkup(page string, root *core.VNode) *core.VNode {
	return root.CloneWithCond(func(n *core.VNode) bool {
		if n.TagName() == "meta" {
			return false
		}

		if belongStr := n.StrAttr("!belong"); belongStr != "" {
			belongs := strings.Split(belongStr, " ")
			for _, belong := range belongs {
				if ds, ok := displayScopes[belong]; ok {
					if ds.HasPage(page) {
						return true
					}
				} else {
					fmt.Printf(`In !belong specification %v:
				no such page or page group with id "%v"`, belongStr, belong)
				}
			}

			return false
		}

		return true
	})
}

type pageDecl struct {
	title      string
	cons       string
	controller string
	vm         string
}

func (g *Compiler) routerDecl(n, root *core.VNode, objStr string, rootHtmlFile string) func() {
	routeCode := ""
	pageConsts := ""
	fns := make([]func(), 0)
	for _, c := range n.ChildElems() {
		if c.TagName() == HanldeTagName {
			route := c.StrAttr("route")
			redirect := c.StrAttr("redirect")
			if redirect != "" {
				routeCode += fmt.Sprintf("\t\t"+
					`r.Handle("%v", page.Redirecter{"%v"})`+"\n",
					route, redirect)
			} else {
				pg := c.ChildElems()[0]
				cons := pg.StrAttr("const")
				if cons == "" {
					continue
				}
				err := (page.Page{Id: cons}).AddTo(displayScopes)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fns = append(fns, func() {
					title := pg.Text()
					pageConsts += cons + ` = "` + cons + `"` + "\n"
					routeId := `"` + route + `"`
					if route == page.NotFoundRoute {
						routeId = "page.NotFoundRoute"
					}
					routeCode += fmt.Sprintf("\t\t"+
						`r.Handle(%v, page.Page{`+"\n"+
						"\t\t\t"+`Id: %v,`+"\n"+
						"\t\t\t"+`Title: "%v",`+"\n", routeId, cons, title)

					ctrlCode := pg.StrAttr("controller")
					if ctrlCode == "" {
						ctrlCode = fmt.Sprintf("func(ctx *page.Context) *VNode { return %vTemplate() }",
							cons)
					}

					routeCode += "\t\t\t" +
						"Controller: " + ctrlCode + ",\n"

					g.genPageCode(cons, pg.StrAttr("vm"), pageMarkup(cons, root), rootHtmlFile)
					routeCode += "\t\t" + `})` + "\n"
				})
			}
		}
	}

	return func() {
		for _, fn := range fns {
			fn()
		}

		g.writeFile("router", fmt.Sprintf(`package %v
		
import "github.com/phaikawl/wade/page"
import . "github.com/phaikawl/wade/core"
		
const (
%v
)
		
func (%v) Setup(r page.Router) {
%v

}`, g.PackageName, pageConsts, objStr, routeCode))
	}
}

func (g *Compiler) componentDecl(n *core.VNode, fromFile string) {
	comName := strings.TrimSpace(n.StrAttr("name"))
	modelName := n.StrAttr("model")
	if modelName == "" {
		printErr(fmt.Sprintf(`No model specified for component "%v"`, comName), "root")
		return
	}

	outputFile := "component_" + comName + ".html"
	varName := "component_" + comName
	if impSrc := n.StrAttr("import_src"); impSrc != "" {
		if impCom := n.StrAttr("import_com"); impCom != "" {
			data := fmt.Sprintf("package %v\n"+`import __imported "%v"`+"\n%v = %v", g.PackageName, impSrc,
				varName, "__imported."+tmplVarPrefix+"component_"+impCom)
			g.writeFile(outputFile, data)
		} else {
			printErr(fmt.Sprintf("No import_com specified for component '%v' importing '%v'\n", comName, impSrc), "root")
		}
	} else {
		comTemp := core.VPrep(&core.VNode{
			Data: comName,
		})
		comTemp.Children = n.Children
		src := fmt.Sprintf("var %v = func(M *%v) *VNode {\n\treturn VPrep(&VNode%v)\n}",
			tmplVarPrefix+varName, modelName, g.Process(comTemp, 1, fromFile))
		g.writeTemplate(outputFile, src)
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

func (g *Compiler) CompileRoot(masterFile string, root *core.VNode) {
	metaElems := vq.New(root).Find(vq.Selector{Tag: "w_meta"})
	var routerDeclFn func()
	if len(metaElems) != 0 {
		for _, meta := range metaElems {
			g.Process(meta, 0, masterFile)
			for _, n := range meta.ChildElems() {
				switch n.TagName() {
				case RouterDeclTagName:
					objStr := n.StrAttr("object")
					if objStr == "" {
						fmt.Println("Error: No object string for router specified.")
					}
					routerDeclFn = g.routerDecl(n, root, objStr, masterFile)
				case ComponentDeclTagName:
					from := fileOf[n]
					g.componentDecl(n, from)
				case PGroupDeclTagName:
					if con := n.StrAttr("const"); con != "" {
						list := []string{}
						for _, p := range n.ChildElems() {
							if p.TagName() == "page" {
								list = append(list, strings.TrimSpace(p.Text()))
							}
						}
						err := (page.PageGroup{
							Id:       con,
							Children: list,
						}).AddTo(displayScopes)
						if err != nil {
							fmt.Println(err)
						}
					}
				default:
					fmt.Printf("Unknown Wade meta tag name: %v\n", n.TagName())
				}
			}
		}
	}

	if routerDeclFn != nil {
		routerDeclFn()
	}
}

func (g Compiler) templateCode(data string) string {
	return fmt.Sprintf(prelude, g.PackageName, data)
}

func (g *Compiler) writeTemplate(htmlFile, data string) {
	g.writeFile(htmlFile, g.templateCode(data))
}

func (g *Compiler) writeFile(htmlFile, content string) {
	filePath := path.Join(g.OutputDir, "compiled_"+path.Base(htmlFile)+".go")
	ioutil.WriteFile(filePath, []byte(content), 0644)
}

var (
	NameRegexp = regexp.MustCompile(`\w*`)
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
		argStr := astr[lp+1 : len(astr)-1]
		if argStr == "" {
			args = []string{}
		} else {
			args = strings.Split(argStr, ",")
		}
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
					for i := range args {
						args[i] = "`" + args[i] + "`"
					}
					fStr = fmt.Sprintf(`RTBinder(RTBinder_%v(func() interface{} {return %v}, []string{%v})),`,
						binder, v, strings.Join(args, ","))
				} else {
					fStr = "func(__node *VNode) {\n"
					fStr += fn(cplData, args, v)
					fStr += "\n" + cplData.Idt + "\t" + "},"
				}
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

	if src := node.StrAttr(OriginFileAttrName); src != "" {
		file = src
	}
	fileOf[node] = file

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
