package main

import (
	"go/ast"
	"os"
	"strings"
)

type fieldInfo struct {
	name     string
	path     string
	typeName string
}

func newAstPkg(pkg *parsedPkg, pkgs pkgMap) *astPkg {
	return &astPkg{
		pkg:        pkg,
		pkgs:       pkgs,
		genImports: make(map[string]string),
	}
}

type astPkg struct {
	pkg *parsedPkg

	pkgs       pkgMap
	genImports map[string]string
}

func pkgLookup(pkg *ast.Package, sym string) (obj *ast.Object, file *ast.File) {
	for _, f := range pkg.Files {
		obj := f.Scope.Lookup(sym)
		if obj != nil {
			return obj, f
		}
	}

	return nil, file
}

type importVisitor struct {
	pkg  *astPkg
	file *ast.File
}

func (iv importVisitor) Visit(node ast.Node) ast.Visitor {
	switch s := node.(type) {
	case *ast.SelectorExpr:
		for _, imp := range iv.file.Imports {
			impName := importName(imp)
			pkgIdent := s.X.(*ast.Ident)
			if impName == pkgIdent.String() {
				newName := iv.pkg.registerImport(impName, importPath(imp))
				pkgIdent.Name = newName
			}
		}
	}

	return iv
}

func (p *astPkg) registerImport(name string, path string) string {
	newName := name
	var ok bool
	var i int
	for {
		existing := p.genImports[newName]
		ok = existing == "" || existing == path
		if ok {
			break
		}

		i++
		newName = name + string(i)
	}

	p.genImports[newName] = path
	return newName
}

func (p *astPkg) lookupSelector(file *ast.File, selector *ast.SelectorExpr) (
	ast.Expr, *ast.File, *parsedPkg) {

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return nil, file, p.pkg
	}

	sel := selector.Sel.String()
	for _, imp := range file.Imports {
		if importName(imp) == ident.Name {
			impPath := importPath(imp)
			pdir := importDir(impPath)
			if pdir != "" {
				pkg := p.pkgs[impPath]
				if pkg != nil {
					obj, file := pkgLookup(pkg.Package, sel)
					if obj != nil {
						if typeSpec, ok := obj.Decl.(*ast.TypeSpec); ok {
							return typeSpec.Type, file, pkg
						}
					}
				}
			}
		}
	}

	return nil, file, p.pkg
}

func (p *astPkg) typeObj(file *ast.File, ftyp ast.Expr) (obj *ast.Object, rfile *ast.File) {

	return nil, file
}

func (p *astPkg) structType(file *ast.File, ftyp ast.Expr) (
	*ast.StructType, *ast.File, *astPkg) {

	switch typ := ftyp.(type) {
	case *ast.StarExpr:
		return p.structType(file, typ.X)

	case *ast.SelectorExpr:
		itype, file, pkg := p.lookupSelector(file, typ)
		if itype != nil {
			apkg := newAstPkg(pkg, p.pkgs)
			return apkg.structType(file, itype)
		}

	case *ast.StructType:
		return typ, file, p

	case *ast.Ident:
		if typ.Obj == nil {
			typ.Obj, file = pkgLookup(p.pkg.Package, typ.Name)
		}

		if typ.Obj != nil {
			if spec, ok := typ.Obj.Decl.(*ast.TypeSpec); ok {
				if spec.Type != nil {
					switch et := spec.Type.(type) {
					case *ast.StructType:
						return et, file, p
					}
				}
			}
		}
	}

	return nil, file, p
}

func (p *astPkg) typeName(file *ast.File, ftyp ast.Expr) (string, error) {
	pos := p.pkg.fset.Position(ftyp.Pos())
	end := p.pkg.fset.Position(ftyp.End())
	of, err := os.Open(pos.Filename)
	if err != nil {
		return "", err
	}

	buf := make([]byte, end.Offset-pos.Offset)
	_, err = of.ReadAt(buf, int64(pos.Offset))
	if err != nil {
		return "", err
	}

	return string(buf), nil
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

	panic(sfmt("Unhandled ast expression type %T", typ))
	return ""
}

func fieldName(f *ast.Field) string {
	if len(f.Names) > 0 {
		return f.Names[0].Name
	}

	return anonFieldName(f.Type)
}

func (p *astPkg) getStateField(fieldPath, fieldName string, field *ast.Field, file *ast.File) (
	*fieldInfo, error) {

	typeName, err := p.typeName(file, field.Type)
	if err != nil {
		return nil, err
	}

	return &fieldInfo{
		name:     fieldName,
		path:     fieldPathDot(fieldPath, fieldName),
		typeName: typeName,
	}, nil
}

func fieldPathDot(prefix, fieldName string) string {
	if prefix == "" {
		return fieldName
	}

	return prefix + "." + fieldName
}

func strListContains(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}

	return false
}

func (p *astPkg) getStateFields(fieldPath string, fields []*ast.Field, file *ast.File) (
	[]*fieldInfo, error) {

	impVisitor := importVisitor{
		pkg:  p,
		file: file,
	}

	var sfs []*fieldInfo

	for _, f := range fields {
		fname := fieldName(f)
		if len(f.Names) == 0 {
			fstruct, sfile, pkg := p.structType(file, f.Type)
			if fstruct != nil {
				fpath := fieldPathDot(fieldPath, fname)
				stateFields, err := pkg.getStateFields(fpath, fstruct.Fields.List, sfile)
				if err != nil {
					return sfs, err
				}

				sfs = append(sfs, stateFields...)
			}
		}

		if f.Tag != nil {
			stag := f.Tag.Value[1 : len(f.Tag.Value)-1]
			if strListContains(strings.Split(stag, " "), "fstate") {
				ast.Walk(impVisitor, f)
				sf, err := p.getStateField(fieldPath, fname, f, file)
				if err != nil {
					return nil, err
				}

				sfs = append(sfs, sf)
			}
		}
	}

	return sfs, nil
}
