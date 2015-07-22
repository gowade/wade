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
	"strings"
	"unicode"

	"gopkg.in/fsnotify.v1"

	"github.com/gowade/html"
	"github.com/gowade/wade/utils/htmlutils"
)

const (
	fuelSuffix  = ".fuel.go"
	genComsFile = "~generatedComs~" + fuelSuffix
	methodsFile = "~methods~" + fuelSuffix
)

var (
	goSrcPath string
)

func srcPath() string {
	if goSrcPath == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			fatal("GOPATH environment variable has not been set, please set it to a correct value.")
		}

		goSrcPath = filepath.Join(gopath, "src")
	}

	return goSrcPath
}

type htmlInfo struct {
	file   string
	markup *html.Node
}

type componentInfo struct {
	htmlInfo htmlInfo
	prefix   string
	name     string
	state    *fieldInfo
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
	cmd := exec.Command("gopherjs", "build", `--tags="js"`, "-o", filepath.Join(serveDir, "main.js"))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "gopherjs failed with %v\n", err.Error())
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

	files, err := ioutil.ReadDir(srcPath())
	if err != nil {
		fatal("Cannot read %v: %v", srcPath(), err)
	}

	for _, f := range files {
		dir := filepath.Join(srcPath(), f.Name())
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			prefix := "/" + f.Name()
			http.Handle(prefix+"/", http.StripPrefix(prefix,
				http.FileServer(http.Dir(dir))))
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexBytes, err := ioutil.ReadFile(idxFile)
		if err != nil {
			panic(err)
		}

		w.Write(indexBytes)
	})

	fmt.Printf("Serving at :%v\n", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		fatal(err.Error())
	}
}

func (f *Fuel) Serve(dir string, indexFile string, port string, serveOnly bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fatal(err.Error())
	}

	f.BuildPackage(dir, "", watcher, serveOnly)
	f.runGopherjs(indexFile)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					fileName := event.Name
					if fileIsRelevant(fileName) {
						log.Println("modified " + fileName)
						f.components = make(componentMap)
						f.BuildPackage(filepath.Dir(fileName), "", nil, serveOnly)
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

func importPath(imp *ast.ImportSpec) string {
	return imp.Path.Value[1 : len(imp.Path.Value)-1]
}

func importDir(impPath string) string {
	pdir := filepath.Join(srcPath(), filepath.FromSlash(impPath))
	if _, err := os.Stat(pdir); err == nil {
		return pdir
	}

	return ""
}

func importName(imp *ast.ImportSpec) string {
	if imp.Name == nil {
		return path.Base(importPath(imp))
	}

	return imp.Name.String()
}

func (f *Fuel) watchImports(imports []*ast.ImportSpec, watcher *fsnotify.Watcher, serveOnly bool) {
	for _, imp := range imports {
		pdir := importDir(importPath(imp))
		if pdir != "" {
			f.BuildPackage(pdir, "", watcher, serveOnly)
		}
	}
}

func (f *Fuel) BuildPackage(dir string, prefix string, fswatcher *fsnotify.Watcher, serveOnly bool) {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), fuelSuffix)
	}, 0)
	if err != nil {
		printErr(err)
	}

	if fswatcher != nil {
		f.watch(fswatcher, dir)

		fset.Iterate(func(file *token.File) bool {
			f.watch(fswatcher, file.Name())
			return true
		})
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fatal(err.Error())
	}

	htmlComs, htmlFiles, err := f.parseHtmlTemplates(dir, files)
	if err != nil {
		printErr(err)
	}

	var pkg *astPkg
	for _, p := range pkgs {
		if !strings.HasSuffix(p.Name, "_test") {
			pkg = &astPkg{
				Package:    p,
				fset:       fset,
				genImports: map[string]string{},
			}

			pkg.registerImport("wade", "")
			pkg.registerImport("vdom", "")
		}
	}

	if pkg == nil {
		return
	}

	for _, file := range pkg.Files {
		err := f.getComponents(file, htmlComs, prefix, pkg)
		if err != nil {
			printErr(err)
		}

		if fswatcher != nil {
			f.watchImports(file.Imports, fswatcher, serveOnly)
		}
	}

	var needGen bool
	htmlCompiler := NewHTMLCompiler(f.components)
	var pcoms []componentInfo
	var gencoms []string
	generated := make(map[string]bool)

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

				f.BuildPackage(pdir, prefix+".", fswatcher, serveOnly)
			}
		}

		if serveOnly {
			continue
		}

		extcut := len(htmlFile.name) - len(".html")
		ofilename := "~" + string([]rune(htmlFile.name[:extcut])) + fuelSuffix
		ofilepath := filepath.Join(dir, ofilename)
		w, err := os.Create(ofilepath)
		if err != nil {
			fatal(err.Error())
		}
		generated[ofilename] = true
		defer w.Close()
		write(w, prelude(pkg.Name, htmlFile.imports))

		for _, comName := range htmlFile.components {
			if _, ok := f.components[prefix+comName]; !ok {
				gencoms = append(gencoms, comName)
				f.components[prefix+comName] = componentInfo{
					name:     comName,
					prefix:   prefix,
					htmlInfo: htmlComs[comName],
				}
			}

			com, _ := f.components[prefix+comName]
			needGen = true
			write(w, "\n\n")
			//ctree, err := htmlCompiler.Generate(com.htmlInfo.markup, &com)
			//if err != nil {
			//fmt.Fprintf(os.Stderr, "%v\n", err)
			//}

			//emitDomCode(w, ctree, err)
			pcoms = append(pcoms, com)
		}

		runGofmt(ofilepath)
	}

	if !needGen || serveOnly {
		return
	}

	gcf, err := os.Create(filepath.Join(dir, genComsFile))
	write(gcf, prelude(pkg.Name, nil))
	if err != nil {
		fatal(err.Error())
	}
	for _, com := range gencoms {
		comDefTpl.Execute(gcf, comDefTD{
			ComName: com,
		})
	}
	defer gcf.Close()
	generated[genComsFile] = true

	mfile, err := os.Create(filepath.Join(dir, methodsFile))
	if err != nil {
		fatal(err.Error())
	}
	defer mfile.Close()
	generated[methodsFile] = true

	write(mfile, prelude(pkg.Name, pkg.importList()))
	for _, com := range pcoms {
		if com.state != nil {
			err := stateMethodsTpl.Execute(mfile, makeStateMethodsTD(pkg, com))
			if err != nil {
				panic(err)
			}
		}

		err := refsTpl.Execute(mfile, makeRefsTD(com.name, htmlCompiler.comRefs[com.name]))
		if err != nil {
			panic(err)
		}
		writeRerenderMethod(mfile, com.name)
	}

	for _, fi := range files {
		fpath := filepath.Join(dir, fi.Name())
		if strings.HasSuffix(fi.Name(), fuelSuffix) && !generated[filepath.Base(fi.Name())] {
			os.Remove(fpath)
		}
	}
}

func makeRefsTD(comName string, refs []comRef) refsTD {
	if refs == nil {
		refs = []comRef{}
	}

	fields := make([]refFieldTD, 0, len(refs))
	for _, ref := range refs {
		elType, _ := domElType(ref.elTag)
		fields = append(fields, refFieldTD{
			Name: ref.name,
			Type: elType,
		})
	}

	return refsTD{
		ComName:  comName,
		TypeName: comName + "Refs",
		Fields:   fields,
	}
}

func makeStateMethodsTD(pkg *astPkg, com componentInfo) (ret stateMethodsTD) {
	ret.Receiver = "*" + com.name
	ret.StateField = com.state.fieldName
	if com.state.typeName[0] != '*' {
		panic("State type must be pointer!") //it should've been reported before this function
	}
	ret.StateType = com.state.typeName[1:]

	ss := com.state.typeStruct
	if ss != nil {
		for _, f := range ss.Fields.List {
			typeName, err := pkg.typeName(f.Type, com.state.typeStructFile)
			if err != nil {
				printErr(err)
				continue
			}

			ret.Setters = append(ret.Setters, stateFieldTD{
				Name: fieldName(f),
				Type: typeName,
			})
		}
	}

	return ret
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

func (f *Fuel) parseHtmlTemplates(dir string, files []os.FileInfo) (m map[string]htmlInfo, hfs []htmlFileInfo, err error) {
	m = make(map[string]htmlInfo)
	hfs = make([]htmlFileInfo, 0)
	for _, fileInfo := range files {
		if !strings.HasSuffix(fileInfo.Name(), ".html") {
			continue
		}

		var impList []importInfo
		var comList []string

		file, err := os.Open(filepath.Join(dir, fileInfo.Name()))
		if err != nil {
			return nil, nil, err
		}

		nodes, err := htmlutils.ParseFragment(file)
		if err != nil {
			return nil, nil, err
		}

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
						return nil, nil, fmt.Errorf(`%v: <import>'s "from" attribute must be set.`, fileInfo.Name())
					}

					impList = append(impList, imp)
				}

				if unicode.IsUpper([]rune(node.Data)[0]) {
					if _, exists := m[node.Data]; exists {
						return nil, nil, fmt.Errorf(`Fatal Error: Found multiple definitions in HTML for component "%v".`, node.Data)
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

	return m, hfs, nil
}

func anonFieldName(typ ast.Expr) string {
	switch t := typ.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return anonFieldName(t.X)
	case *ast.SelectorExpr:
		return anonFieldName(t.Sel)
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

func (f *Fuel) getComponents(file *ast.File, htmlComs map[string]htmlInfo, prefix string,
	pkg *astPkg) error {
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
							state, err := pkg.getStateField(stype.Fields.List, file)
							if err != nil {
								return fmt.Errorf("error while processing component: %v", err)
							}

							f.components[prefix+name] = componentInfo{
								name:     name,
								prefix:   prefix,
								state:    state,
								htmlInfo: hcom,
							}
						}
					}
				}
			}
		}
	}

	return nil
}
