package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

// this code is the worst code

func setVariable(dirName, varName string, content []byte) (filename string, newFile []byte, err error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dirName, nil, parser.ParseComments)
	if err != nil {
		return "", nil, err
	}

	visitObj := func(obj *ast.Object) (bool, error) {
		if obj.Name != varName {
			return false, nil
		}

		switch obj.Kind {
		case ast.Con, ast.Var:
			// all good
		case ast.Bad, ast.Pkg, ast.Typ, ast.Fun, ast.Lbl:
			return true, fmt.Errorf("can only work on const or var: found a %q definition", obj.Kind)
		default:
			return true, fmt.Errorf("not a const or a var, unknown kind")
		}

		spec, ok := obj.Decl.(*ast.ValueSpec)
		if !ok {
			return true, fmt.Errorf("need a ValueSpec, found a '%T'", obj.Decl)
		}

		if len(spec.Values) > 1 {
			return true, fmt.Errorf("can't set more than one value, got %d", len(spec.Values))
		}
		if len(spec.Values) == 1 {
			switch value := spec.Values[0].(type) {
			case *ast.BasicLit, *ast.CompositeLit:
			default:
				return true, fmt.Errorf("can only overwrite a literals, found a %T", value)
			}
		}

		var isString bool
		var isByteSlice bool

		switch t := spec.Type.(type) {
		case *ast.Ident:
			if t.Name == "string" {
				isString = true
			}
		case *ast.ArrayType:
			tElt, ok := t.Elt.(*ast.Ident)
			if !ok {
				return true, fmt.Errorf("need a '*ast.Ident' or '*ast.ArrayType', found a '%T'", spec.Type)
			}
			if tElt.Name == "byte" {
				isByteSlice = true
			}
		case nil:
			switch value := spec.Values[0].(type) {
			case *ast.BasicLit:
				if value.Kind == token.STRING {
					isString = true
				}
			case *ast.CompositeLit:
				arrayType, ok := value.Type.(*ast.ArrayType)
				if !ok {
					return true, fmt.Errorf("need a composite of '*ast.ArrayType', found a '%T'", value.Type)
				}
				eltIdent, ok := arrayType.Elt.(*ast.Ident)
				if !ok {
					return true, fmt.Errorf("array element must have an identity, found a '%T'", arrayType.Elt)
				}
				if eltIdent.Name == "byte" {
					isByteSlice = true
				}
			}
		default:
			return true, fmt.Errorf("need a '*ast.Ident' or '*ast.ArrayType', found a '%T'", spec.Type)
		}

		valueExpr := &ast.BasicLit{}

		switch {
		case isString:
			valueExpr.Value = fmt.Sprintf("%#v", string(content))
		case isByteSlice:
			valueExpr.Value = fmt.Sprintf("%#v", content)
		default:
			return false, fmt.Errorf("unsupported type: %v", spec.Type)
		}
		spec.Values = []ast.Expr{valueExpr}

		spec.Type = nil
		return true, nil
	}

	var node *ast.File
loop:
	for _, pkg := range pkgs {
		for nodeFilename, file := range pkg.Files {
			obj := file.Scope.Lookup(varName)
			if obj == nil {
				continue
			}
			changed, err := visitObj(obj)
			if !changed || err != nil {
				return "", nil, err
			}
			filename = nodeFilename
			node = file
			break loop
		}
	}

	var buf bytes.Buffer
	err = format.Node(&buf, fset, node)
	return filename, buf.Bytes(), err
}
