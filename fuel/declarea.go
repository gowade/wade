package main

import (
	"bytes"
)

func newDeclArea(parent *declArea) *declArea {
	return &declArea{
		nameIdx: map[string]int{},
		parent:  parent,
	}
}

type declCodeTD struct {
	VarName string
	Code    *bytes.Buffer
}

type declArea struct {
	parent  *declArea
	nameIdx map[string]int
	decls   []declCodeTD
}

// declare adds a declaration to the declaration area,
// add a new number suffix to it if the name is already taken,
// returns the new valid name and the buffer for the new variable's code content
func (z *declArea) declare(varName string) (string, *bytes.Buffer) {
	if z.parent != nil {
		z.nameIdx[varName] += z.parent.nameIdx[varName]
	}

	z.nameIdx[varName]++

	var buf bytes.Buffer
	newVarName := sfmt("%v%v", varName, z.nameIdx[varName])
	z.decls = append(z.decls, declCodeTD{
		VarName: newVarName,
		Code:    &buf,
	})

	return newVarName, &buf

}

// code returns the accumulated code content of all the variables
func (z *declArea) code() *bytes.Buffer {
	buf, err := execTplBuf(varDeclTpl, varDeclTD{
		Vars: z.decls,
	})

	if err != nil {
		panic(err)
	}

	return buf
}
