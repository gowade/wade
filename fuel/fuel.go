package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"

	"gopkg.in/fsnotify.v1"

	"github.com/gowade/html"
	"github.com/gowade/wade/utils/htmlutils"
)

const (
	fuelSuffix = ".fuel.go"
)

var (
	gSrcPath string
)

func srcPath() string {
	if gSrcPath == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			fatal("GOPATH environment variable has not been set, please set it to a correct value.")
		}

		gSrcPath = filepath.Join(gopath, "src")
	}

	return gSrcPath
}

type htmlInfo struct {
	file   string
	markup *html.Node
}

type stateInfo struct {
	field     string
	typ       string
	structTyp *ast.StructType
	isPointer bool
}

type componentInfo struct {
	htmlInfo  htmlInfo
	prefix    string
	name      string
	argFields map[string]bool
	state     stateInfo
}

func (z componentInfo) fullName() string {
	return z.prefix + z.name
}

type componentMap map[string]componentInfo

type Fuel struct {
	dir        string
	components componentMap
}

func NewFuel() *Fuel {
	return &Fuel{
		components: componentMap{},
	}
}

func (f *Fuel) runGopherjs(indexFile string) {
	serveDir := filepath.Dir(indexFile)
	cmd := exec.Command("gopherjs", "build", "-o", filepath.Join(serveDir, "main.js"))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()

	if err != nil {
		fatal(`gopherjs failed with %v`, err.Error())
	}
}

func fileIsRelevant(fileName string) bool {
	return (strings.HasSuffix(fileName, ".html") ||
		strings.HasSuffix(fileName, ".go")) &&
		!strings.HasSuffix(fileName, fuelSuffix)
}

func (f *Fuel) serveHTTP(idxFile string, port string) {
	serveDir := filepath.Dir(idxFile)
	servePath := path.Join("/", filepath.ToSlash(serveDir))
	http.Handle(servePath+"/", http.StripPrefix(servePath,
		http.FileServer(http.Dir(serveDir))))

	http.Handle("/gopath/", http.StripPrefix("/gopath", http.FileServer(http.Dir(os.Getenv("GOPATH")))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexBytes, err := ioutil.ReadFile(idxFile)
		if err != nil {
			panic(err)
		}

		w.Write(indexBytes)
	})

	fmt.Printf("Serving at :%v\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fatal(err.Error())
	}
}

func (f *Fuel) Serve(dir string, indexFile string, port string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fatal(err.Error())
	}

	f.BuildPackage(dir, "", watcher)
	f.runGopherjs(indexFile)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					fileName := event.Name
					if fileIsRelevant(fileName) {
						log.Println("modified " + fileName)
						f.BuildPackage(filepath.Dir(fileName), "", nil)
						f.runGopherjs(indexFile)
					}
				}

			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	f.serveHTTP(indexFile, port)
}

func (f *Fuel) watch(watcher *fsnotify.Watcher, file string) {
	if strings.HasSuffix(file, fuelSuffix) {
		return
	}

	err := watcher.Add(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error watching file %v: %v\n", file, err)
	}
}

func (f *Fuel) watchImports(imports []*ast.ImportSpec, watcher *fsnotify.Watcher) {
	for _, imp := range imports {
		path := imp.Path.Value[1 : len(imp.Path.Value)-1]
		pdir := filepath.Join(srcPath(), filepath.FromSlash(path))

		if _, err := os.Stat(pdir); err == nil {
			f.BuildPackage(pdir, "", watcher)
		}

		//watcher.Add(pdir)

		//files, err := ioutil.ReadDir(pdir)
		//checkFatal(err)

		//for _, file := range files {
		//watcher.Add(file.Name())
		//}

		//fset := token.NewFileSet()
		//pkgs, err := parser.ParseDir(fset, pdir, func(fi os.FileInfo) bool {
		//return !strings.HasSuffix(fi.Name(), fuelSuffix)
		//}, 0)
		//checkFatal(err)

		//for _, pkg := range pkgs {
		//if !strings.HasSuffix(pkg.Name, "_test") {
		//for _, file := range pkg.Files {
		//f.watchImports(file.Imports, watcher)
		//}
		//}
		//}
		//}
	}
}

func (f *Fuel) BuildPackage(dir string, prefix string, fswatcher *fsnotify.Watcher) {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), fuelSuffix)
	}, 0)
	checkFatal(err)

	if fswatcher != nil {
		f.watch(fswatcher, dir)

		fset.Iterate(func(file *token.File) bool {
			f.watch(fswatcher, file.Name())
			return true
		})
	}

	htmlComs, htmlFiles := f.parseHtmlTemplates(dir)
	var pkgName string

	for _, pkg := range pkgs {
		if !strings.HasSuffix(pkg.Name, "_test") {
			if pkgName == "" {
				pkgName = pkg.Name
			}

			ast.PackageExports(pkg)
			for _, file := range pkg.Files {
				f.getComponents(file, htmlComs, prefix)

				if fswatcher != nil {
					f.watchImports(file.Imports, fswatcher)
				}
			}
		}
	}

	var needGen bool
	htmlCompiler := NewHTMLCompiler(f.components)
	var pcoms []componentInfo
	for _, htmlFile := range htmlFiles {
		if fswatcher != nil {
			f.watch(fswatcher, filepath.Join(dir, htmlFile.name))
		}

		for _, imp := range htmlFile.imports {
			pdir := filepath.Join(srcPath(), filepath.FromSlash(imp.path))
			if _, err := os.Stat(pdir); err == nil {
				prefix := imp.as
				if prefix == "" {
					prefix = filepath.Base(pdir)
				}

				f.BuildPackage(pdir, prefix+".", fswatcher)
			}
		}

		extcut := len(htmlFile.name) - len(".html")
		ofilename := "g." + string([]rune(htmlFile.name[:extcut])) + fuelSuffix
		ofilepath := filepath.Join(dir, ofilename)
		w, err := os.Create(ofilepath)
		if err != nil {
			fatal(err.Error())
		}
		defer w.Close()
		write(w, prelude(pkgName, htmlFile.imports))

		for _, comName := range htmlFile.components {
			if com, ok := f.components[prefix+comName]; ok {
				needGen = true
				write(w, "\n\n")
				ctree, err := htmlCompiler.Generate(com.htmlInfo.markup, &com)
				if err != nil {
					fatal(err.Error())
				}

				emitDomCode(w, ctree)
				pcoms = append(pcoms, com)
			} else {
				fatal("No struct definition for %v component.", comName)
			}
		}

		runGofmt(ofilepath)
	}

	if !needGen {
		return
	}

	mfile, err := os.Create(filepath.Join(dir, "g.methods.fuel.go"))
	if err != nil {
		fatal(err.Error())
	}
	defer mfile.Close()

	write(mfile, prelude(pkgName, nil))
	for _, com := range pcoms {
		if com.state.field != "" {
			write(mfile, stateMethsCode(com, fset))
		}

		write(mfile, comRefsDeclCode(com.name, htmlCompiler.comRefs[com.name]))
		write(mfile, comRefsMethsCode(com.name))
		write(mfile, fmt.Sprintf(`func (this *%v) Rerender() {
	r := this.Render(nil)
	vdom.PerformDiff(r, this.VNode.Render(), this.VNode.DOMNode())
	this.VNode.ComRend = r
	this.VNode = r
}

`, com.name))
	}
}

func comRefsMethsCode(comName string) string {
	return fmt.Sprintf(`func (this *%v) Refs() %vRefs {
	return this.Com.InternalRefsHolder.(%vRefs)	
}

`, comName, comName, comName)
}

func comRefsDeclCode(comName string, refs []comRef) string {
	fields := make([]string, 0, len(refs))
	if refs != nil {
		fields = append(fields, "")
		for _, ref := range refs {
			elTp, _ := domElType(ref.elTag)
			fields = append(fields, fmt.Sprintf("\t%v %v", ref.name, elTp))
		}
		fields = append(fields, "")
	}
	return fmt.Sprintf(`type %vRefs struct {%v}
`, comName, strings.Join(fields, "\n"))
}

func stateMethsCode(com componentInfo, fset *token.FileSet) string {
	setters := ""
	for _, f := range com.state.structTyp.Fields.List {
		pos := fset.Position(f.Type.Pos())
		end := fset.Position(f.Type.End())
		file, err := os.Open(pos.Filename)
		if err != nil {
			fatal(err.Error())
		}

		buf := make([]byte, end.Offset-pos.Offset)
		_, err = file.ReadAt(buf, int64(pos.Offset))
		if err != nil {
			fatal(err.Error())
		}

		fname := fieldName(f)
		setters += fmt.Sprintf(`func (this *%v) Set%v(v %v) {
	this.%v.%v = v
	this.Rerender()
}

`, com.name, fname, string(buf), com.state.field, fname)
	}

	return fmt.Sprintf(`func (this *%v) InternalState() interface{} {
	return this.%v
}

`,
		//com.name, com.stateField, com.stateType,
		com.name, com.state.field) + setters
}

type importInfo struct {
	path string
	as   string
}

type htmlFileInfo struct {
	name       string
	imports    []importInfo
	components []string
}

func (f *Fuel) parseHtmlTemplates(dir string) (map[string]htmlInfo, []htmlFileInfo) {
	files, err := ioutil.ReadDir(dir)
	checkFatal(err)

	m := make(map[string]htmlInfo)
	hfs := make([]htmlFileInfo, 0)
	for _, fileInfo := range files {
		if !strings.HasSuffix(fileInfo.Name(), ".html") {
			continue
		}

		var impList []importInfo
		var comList []string

		file, err := os.Open(filepath.Join(dir, fileInfo.Name()))
		checkFatal(err)
		nodes, err := htmlutils.ParseFragment(file)
		checkFatal(err)

		for _, node := range nodes {
			if node.Type == html.ElementNode {
				if node.Data == "import" {
					var imp importInfo
					for _, attr := range node.Attr {
						switch attr.Key {
						case "from":
							imp.path = attr.Val
						case "as":
							imp.as = attr.Val
						}
					}

					if imp.path == "" {
						fatal(`%v: <import>'s "from" attribute must be set.`, fileInfo.Name())
					}

					impList = append(impList, imp)
				}

				if unicode.IsUpper([]rune(node.Data)[0]) {
					if _, exists := m[node.Data]; exists {
						fatal(`Fatal Error: Found multiple definitions in HTML for component "%v".`, node.Data)
					}

					comList = append(comList, node.Data)
					m[node.Data] = htmlInfo{
						markup: node,
						file:   fileInfo.Name(),
					}
				}
			}
		}

		hfs = append(hfs, htmlFileInfo{
			name:       fileInfo.Name(),
			components: comList,
			imports:    impList,
		})
	}

	return m, hfs
}

func anonFieldName(typ ast.Expr) string {
	switch t := typ.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return anonFieldName(t.X)
	case *ast.SelectorExpr:
		return anonFieldName(t.X)
	}

	panic(fmt.Sprintf("Unhandled ast expression type %T", typ))
	return ""
}

func fieldName(f *ast.Field) string {
	if len(f.Names) > 0 {
		return f.Names[0].Name
	}

	return anonFieldName(f.Type)
}

func extractFields(comName string, fields []*ast.Field) (map[string]bool, stateInfo) {
	argFields := make(map[string]bool)
	var state stateInfo
	for _, f := range fields {
		fname := fieldName(f)
		if f.Tag != nil {
			var stag = reflect.StructTag(f.Tag.Value[1 : len(f.Tag.Value)-1])
			if sf := stag.Get("fuel"); sf == "state" {
				if state.field != "" {
					fatal("Error processing component %v: component can only have 1 state field.", comName)
				}

				var typIden *ast.Ident
				if pt, ok := f.Type.(*ast.StarExpr); ok {
					if ftype, ok := pt.X.(*ast.Ident); ok {
						typIden = ftype
						state.isPointer = true
					}
				} else {
					if ftype, ok := f.Type.(*ast.Ident); ok {
						typIden = ftype
						state.isPointer = false
					}
				}

				if typIden != nil {
					state.field = fname
					state.typ = typIden.Name
					if spec, ok := typIden.Obj.Decl.(*ast.TypeSpec); ok {
						if st, ok := spec.Type.(*ast.StructType); ok {
							state.structTyp = st
						}
					}
					continue
				}

				fatal(`Error processing field "%v" of component %v: state field's type must be a named type (anonymous struct is forbidden).`, fname, comName)
			}
		}

		argFields[fname] = true
	}

	return argFields, state
}

func (f *Fuel) getComponents(file *ast.File, htmlComs map[string]htmlInfo, prefix string) {
	for _, decl := range file.Decls {
		switch gdecl := decl.(type) {
		case *ast.GenDecl:
			if gdecl.Tok == token.TYPE {
				for _, ospec := range gdecl.Specs {
					spec := ospec.(*ast.TypeSpec)
					name := spec.Name.Name
					switch stype := spec.Type.(type) {
					case *ast.StructType:
						if hcom, ok := htmlComs[name]; ok {
							argFields, state := extractFields(name, stype.Fields.List)
							//println("<<<", prefix+name)
							f.components[prefix+name] = componentInfo{
								name:      name,
								prefix:    prefix,
								argFields: argFields,
								state:     state,
								htmlInfo:  hcom,
							}
						}
					}
				}
			}
		}
	}
}
