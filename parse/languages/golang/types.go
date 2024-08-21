// Copyright (c) 2020, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package golang

import (
	"fmt"
	"strings"

	"cogentcore.org/core/parse"
	"cogentcore.org/core/parse/parser"
	"cogentcore.org/core/parse/syms"
	"cogentcore.org/core/parse/token"
)

var TraceTypes = false

// IsQualifiedType returns true if type is qualified by a package prefix
// is sensitive to [] or map[ prefix so it does NOT report as a qualified type in that
// case -- it is a compound local type defined in terms of a qualified type.
func IsQualifiedType(tnm string) bool {
	if strings.HasPrefix(tnm, "[]") || strings.HasPrefix(tnm, "map[") {
		return false
	}
	return strings.Index(tnm, ".") > 0
}

// QualifyType returns the type name tnm qualified by pkgnm if it is non-empty
// and only if tnm is not a basic type name
func QualifyType(pkgnm, tnm string) string {
	if pkgnm == "" || IsQualifiedType(tnm) {
		return tnm
	}
	if _, btyp := BuiltinTypes[tnm]; btyp {
		return tnm
	}
	return pkgnm + "." + tnm
}

// SplitType returns the package and type names from a potentially qualified
// type name -- if it is not qualified, package name is empty.
// is sensitive to [] prefix so it does NOT split in that case
func SplitType(nm string) (pkgnm, tnm string) {
	if !IsQualifiedType(nm) {
		return "", nm
	}
	sci := strings.Index(nm, ".")
	return nm[:sci], nm[sci+1:]
}

// PrefixType returns the type name prefixed with given prefix -- keeps any
// package name as the outer scope.
func PrefixType(pfx, nm string) string {
	pkgnm, tnm := SplitType(nm)
	return QualifyType(pkgnm, pfx+tnm)
}

// FindTypeName finds given type name in pkg and in broader context
// returns new package symbol if type name is in a different package
// else returns pkg arg.
func (gl *GoLang) FindTypeName(tynm string, fs *parse.FileState, pkg *syms.Symbol) (*syms.Type, *syms.Symbol) {
	if tynm == "" {
		return nil, nil
	}
	if tynm[0] == '*' {
		tynm = tynm[1:]
	}
	pnm, tnm := SplitType(tynm)
	if pnm == "" {
		if btyp, ok := BuiltinTypes[tnm]; ok {
			return btyp, pkg
		}
		if gtyp, ok := pkg.Types[tnm]; ok {
			return gtyp, pkg
		}
		// if TraceTypes {
		// 	fmt.Printf("FindTypeName: unqualified type name: %v not found in package: %v\n", tnm, pkg.Name)
		// }
		return nil, pkg
	}
	if npkg, ok := gl.PkgSyms(fs, pkg.Children, pnm); ok {
		if gtyp, ok := npkg.Types[tnm]; ok {
			return gtyp, npkg
		}
		if TraceTypes {
			fmt.Printf("FindTypeName: type name: %v not found in package: %v\n", tnm, pnm)
		}
	} else {
		if TraceTypes {
			fmt.Printf("FindTypeName: could not find package: %v\n", pnm)
		}
	}
	if TraceTypes {
		fmt.Printf("FindTypeName: type name: %v not found in package: %v\n", tynm, pkg.Name)
	}
	return nil, pkg
}

// ResolveTypes initializes all user-defined types from AST data
// and then resolves types of symbols.  The pkg must be a single
// package symbol i.e., the children there are all the elements of the
// package and the types are all the global types within the package.
// funInternal determines whether to include function-internal symbols
// (e.g., variables within function scope -- only for local files).
func (gl *GoLang) ResolveTypes(fs *parse.FileState, pkg *syms.Symbol, funInternal bool) {
	fs.SymsMu.Lock()
	gl.TypesFromAST(fs, pkg)
	gl.InferSymbolType(pkg, fs, pkg, funInternal)
	fs.SymsMu.Unlock()
}

// TypesFromAST initializes the types from their AST parse
func (gl *GoLang) TypesFromAST(fs *parse.FileState, pkg *syms.Symbol) {
	InstallBuiltinTypes()

	for _, ty := range pkg.Types {
		gl.InitTypeFromAST(fs, pkg, ty)
	}
}

// InitTypeFromAST initializes given type from ast
func (gl *GoLang) InitTypeFromAST(fs *parse.FileState, pkg *syms.Symbol, ty *syms.Type) {
	if ty.AST == nil || len(ty.AST.AsTree().Children) < 2 {
		// if TraceTypes {
		// 	fmt.Printf("TypesFromAST: Type has nil AST! %v\n", ty.String())
		// }
		return
	}
	tyast := ty.AST.(*parser.AST).ChildAST(1)
	if tyast == nil {
		if TraceTypes {
			fmt.Printf("TypesFromAST: Type has invalid AST! %v missing child 1\n", ty.String())
		}
		return
	}
	if ty.Name == "" {
		if TraceTypes {
			fmt.Printf("TypesFromAST: Type has no name! %v\n", ty.String())
		}
		return
	}
	if ty.Initialized {
		// if TraceTypes {
		// 	fmt.Printf("Type: %v already initialized\n", ty.Name)
		// }
		return
	}
	gl.TypeFromAST(fs, pkg, ty, tyast)
	gl.TypeMeths(fs, pkg, ty) // all top-level named types might have methods
	ty.Initialized = true
}

// SubTypeFromAST returns a subtype from child ast at given index, nil if failed
func (gl *GoLang) SubTypeFromAST(fs *parse.FileState, pkg *syms.Symbol, ast *parser.AST, idx int) (*syms.Type, bool) {
	sast := ast.ChildAST(idx)
	if sast == nil {
		if TraceTypes {
			fmt.Printf("TraceTypes: could not find child %d on ast %v", idx, ast)
		}
		return nil, false
	}
	return gl.TypeFromAST(fs, pkg, nil, sast)
}

// TypeToKindMap maps AST type names to syms.Kind basic categories for how we
// treat them for subsequent processing.  Basically: Primitive or Composite
var TypeToKindMap = map[string]syms.Kinds{
	"BasicType":     syms.Primitive,
	"TypeNm":        syms.Primitive,
	"QualType":      syms.Primitive,
	"PointerType":   syms.Primitive,
	"MapType":       syms.Composite,
	"SliceType":     syms.Composite,
	"ArrayType":     syms.Composite,
	"StructType":    syms.Composite,
	"InterfaceType": syms.Composite,
	"FuncType":      syms.Composite,
	"StringDbl":     syms.KindsN, // note: Lit is removed by ASTTypeName
	"StringTicks":   syms.KindsN,
	"Rune":          syms.KindsN,
	"NumInteger":    syms.KindsN,
	"NumFloat":      syms.KindsN,
	"NumImag":       syms.KindsN,
}

// ASTTypeName returns the effective type name from ast node
// dropping the "Lit" for example.
func (gl *GoLang) ASTTypeName(tyast *parser.AST) string {
	tnm := tyast.Name
	if strings.HasPrefix(tnm, "Lit") {
		tnm = tnm[3:]
	}
	return tnm
}

// TypeFromAST returns type from AST parse -- returns true if successful.
// This is used both for initialization of global types via TypesFromAST
// and also for online type processing in the course of tracking down
// other types while crawling the AST.  In the former case, ty is non-nil
// and the goal is to fill out the type information -- the ty will definitely
// have a name already.  In the latter case, the ty will be nil, but the
// tyast node may have a Src name that will first be looked up to determine
// if a previously processed type is already available.  The tyast.Name is
// the parser categorization of the type  (BasicType, StructType, etc).
func (gl *GoLang) TypeFromAST(fs *parse.FileState, pkg *syms.Symbol, ty *syms.Type, tyast *parser.AST) (*syms.Type, bool) {
	tnm := gl.ASTTypeName(tyast)
	bkind, ok := TypeToKindMap[tnm]
	if !ok { // must be some kind of expression
		sty, _, got := gl.TypeFromASTExprStart(fs, pkg, pkg, tyast)
		return sty, got
	}
	switch bkind {
	case syms.Primitive:
		return gl.TypeFromASTPrim(fs, pkg, ty, tyast)
	case syms.Composite:
		return gl.TypeFromASTComp(fs, pkg, ty, tyast)
	case syms.KindsN:
		return gl.TypeFromASTLit(fs, pkg, ty, tyast)
	}
	return nil, false
}

// TypeFromASTPrim handles primitive (non composite) type processing
func (gl *GoLang) TypeFromASTPrim(fs *parse.FileState, pkg *syms.Symbol, ty *syms.Type, tyast *parser.AST) (*syms.Type, bool) {
	tnm := gl.ASTTypeName(tyast)
	src := tyast.Src
	etyp, tpkg := gl.FindTypeName(src, fs, pkg)
	if etyp != nil {
		if ty == nil { // if we can find an existing type, and not filling in global, use it
			if tpkg != pkg {
				pkgnm := tpkg.Name
				qtnm := QualifyType(pkgnm, etyp.Name)
				if qtnm != etyp.Name {
					if letyp, ok := pkg.Types[qtnm]; ok {
						etyp = letyp
					} else {
						ntyp := &syms.Type{}
						*ntyp = *etyp
						ntyp.Name = qtnm
						pkg.Types.Add(ntyp)
						etyp = ntyp
					}
				}
			}
			return etyp, true
		}
	} else {
		if TraceTypes && src != "" {
			fmt.Printf("TypeFromAST: primitive type name: %v not found\n", src)
		}
	}
	switch tnm {
	case "BasicType":
		if etyp != nil {
			ty.Kind = etyp.Kind
			ty.Els.Add("par", etyp.Name) // parent type
			return ty, true
		} else {
			return nil, false
		}
	case "TypeNm", "QualType":
		if etyp != nil && etyp != ty {
			ty.Kind = etyp.Kind
			if ty.Name != etyp.Name {
				ty.Els.Add("par", etyp.Name) // parent type
				if TraceTypes {
					fmt.Printf("TypeFromAST: TypeNm %v defined from parent type: %v\n", ty.Name, etyp.Name)
				}
			}
			return ty, true
		} else {
			return nil, false
		}
	case "PointerType":
		if ty == nil {
			ty = &syms.Type{}
		}
		ty.Kind = syms.Ptr
		if sty, ok := gl.SubTypeFromAST(fs, pkg, tyast, 0); ok {
			ty.Els.Add("ptr", sty.Name)
			if ty.Name == "" {
				ty.Name = "*" + sty.Name
				pkg.Types.Add(ty) // add pointers so we don't have to keep redefining
				if TraceTypes {
					fmt.Printf("TypeFromAST: Adding PointerType: %v\n", ty.String())
				}
			}
			return ty, true
		} else {
			return nil, false
		}
	}
	return nil, false
}

// TypeFromASTComp handles composite type processing
func (gl *GoLang) TypeFromASTComp(fs *parse.FileState, pkg *syms.Symbol, ty *syms.Type, tyast *parser.AST) (*syms.Type, bool) {
	tnm := gl.ASTTypeName(tyast)
	newTy := false
	if ty == nil {
		newTy = true
		ty = &syms.Type{}
	}
	switch tnm {
	case "MapType":
		ty.Kind = syms.Map
		keyty, kok := gl.SubTypeFromAST(fs, pkg, tyast, 0)
		valty, vok := gl.SubTypeFromAST(fs, pkg, tyast, 1)
		if kok && vok {
			ty.Els.Add("key", SymTypeNameForPkg(keyty, pkg))
			ty.Els.Add("val", SymTypeNameForPkg(valty, pkg))
			if newTy {
				ty.Name = "map[" + keyty.Name + "]" + valty.Name
			}
		} else {
			return nil, false
		}
	case "SliceType":
		ty.Kind = syms.List
		valty, ok := gl.SubTypeFromAST(fs, pkg, tyast, 0)
		if ok {
			ty.Els.Add("val", SymTypeNameForPkg(valty, pkg))
			if newTy {
				ty.Name = "[]" + valty.Name
			}
		} else {
			return nil, false
		}
	case "ArrayType":
		ty.Kind = syms.Array
		valty, ok := gl.SubTypeFromAST(fs, pkg, tyast, 1)
		if ok {
			ty.Els.Add("val", SymTypeNameForPkg(valty, pkg))
			if newTy {
				ty.Name = "[]" + valty.Name // todo: get size from child0, set to Size
			}
		} else {
			return nil, false
		}
	case "StructType":
		ty.Kind = syms.Struct
		nfld := len(tyast.Children)
		if nfld == 0 {
			return BuiltinTypes["struct{}"], true
		}
		ty.Size = []int{nfld}
		for i := 0; i < nfld; i++ {
			fld := tyast.Children[i].(*parser.AST)
			fsrc := fld.Src
			switch fld.Name {
			case "NamedField":
				if len(fld.Children) <= 1 { // anonymous, non-qualified
					ty.Els.Add(fsrc, fsrc)
					gl.StructInheritEls(fs, pkg, ty, fsrc)
					continue
				}
				fldty, ok := gl.SubTypeFromAST(fs, pkg, fld, 1)
				if ok {
					nms := gl.NamesFromAST(fs, pkg, fld, 0)
					for _, nm := range nms {
						ty.Els.Add(nm, SymTypeNameForPkg(fldty, pkg))
					}
				}
			case "AnonQualField":
				ty.Els.Add(fsrc, fsrc) // anon two are same
				gl.StructInheritEls(fs, pkg, ty, fsrc)
			}
		}
		if newTy {
			ty.Name = fs.NextAnonName(pkg.Name + "_struct")
		}
		// if TraceTypes {
		// 	fmt.Printf("TypeFromAST: New struct type defined: %v\n", ty.Name)
		// 	ty.WriteDoc(os.Stdout, 0)
		// }
	case "InterfaceType":
		ty.Kind = syms.Interface
		nmth := len(tyast.Children)
		if nmth == 0 {
			return BuiltinTypes["interface{}"], true
		}
		ty.Size = []int{nmth}
		for i := 0; i < nmth; i++ {
			fld := tyast.Children[i].(*parser.AST)
			fsrc := fld.Src
			switch fld.Name {
			case "MethSpecAnonLocal":
				fallthrough
			case "MethSpecAnonQual":
				ty.Els.Add(fsrc, fsrc) // anon two are same
			case "MethSpecName":
				if nm := fld.ChildAST(0); nm != nil {
					mty := syms.NewType(ty.Name+":"+nm.Src, syms.Method)
					pkg.Types.Add(mty)                    // add interface methods as new types..
					gl.FuncTypeFromAST(fs, pkg, fld, mty) // todo: this is not working -- debug
					ty.Els.Add(nm.Src, mty.Name)
				}
			}
		}
		if newTy {
			if nmth == 0 {
				ty.Name = "interface{}"
			} else {
				ty.Name = fs.NextAnonName(pkg.Name + "_interface")
			}
		}
	case "FuncType":
		ty.Kind = syms.Func
		gl.FuncTypeFromAST(fs, pkg, tyast, ty)
		if newTy {
			if len(ty.Els) == 0 {
				ty.Name = "func()"
			} else {
				ty.Name = fs.NextAnonName(pkg.Name + "_func")
			}
		}
	}
	if newTy {
		etyp, has := pkg.Types[ty.Name]
		if has {
			return etyp, true
		}
		pkg.Types.Add(ty) // add anon composite types
		// if TraceTypes {
		// 	fmt.Printf("TypeFromASTComp: Created new anon composite type: %v %s\n", ty.Name, ty.String())
		// }
	}
	return ty, true // fallthrough is true..
}

// TypeFromASTLit gets type from literals
func (gl *GoLang) TypeFromASTLit(fs *parse.FileState, pkg *syms.Symbol, ty *syms.Type, tyast *parser.AST) (*syms.Type, bool) {
	tnm := tyast.Name
	var bty *syms.Type
	switch tnm {
	case "LitStringDbl":
		bty = BuiltinTypes["string"]
	case "LitStringTicks":
		bty = BuiltinTypes["string"]
	case "LitRune":
		bty = BuiltinTypes["rune"]
	case "LitNumInteger":
		bty = BuiltinTypes["int"]
	case "LitNumFloat":
		bty = BuiltinTypes["float64"]
	case "LitNumImag":
		bty = BuiltinTypes["complex128"]
	}
	if bty == nil {
		return nil, false
	}
	if ty == nil {
		return bty, true
	}
	ty.Kind = bty.Kind
	ty.Els.Add("par", bty.Name) // parent type
	return ty, true
}

// StructInheritEls inherits struct fields and meths from given embedded type.
// Ensures that copied values are properly qualified if from another package.
func (gl *GoLang) StructInheritEls(fs *parse.FileState, pkg *syms.Symbol, ty *syms.Type, etynm string) {
	ety, _ := gl.FindTypeName(etynm, fs, pkg)
	if ety == nil {
		if TraceTypes {
			fmt.Printf("Embedded struct type not found: %v for type: %v\n", etynm, ty.Name)
		}
		return
	}
	if !ety.Initialized {
		// if TraceTypes {
		// 	fmt.Printf("Embedded struct type not yet initialized, initializing: %v for type: %v\n", ety.Name, ty.Name)
		// }
		gl.InitTypeFromAST(fs, pkg, ety)
	}
	pkgnm := pkg.Name
	diffPkg := false
	epkg, has := ety.Scopes[token.NamePackage]
	if has && epkg != pkgnm {
		diffPkg = true
	}
	if diffPkg {
		for i := range ety.Els {
			nt := ety.Els[i].Clone()
			tnm := nt.Type
			_, isb := BuiltinTypes[tnm]
			if !isb && !IsQualifiedType(tnm) {
				tnm = QualifyType(epkg, tnm)
				// fmt.Printf("Fixed type: %v to %v\n", ety.Els[i].Type, tnm)
			}
			nt.Type = tnm
			ty.Els = append(ty.Els, *nt)
		}
		nmt := len(ety.Meths)
		if nmt > 0 {
			ty.Meths = make(syms.TypeMap, nmt)
			for mn, mt := range ety.Meths {
				nmt := mt.Clone()
				for i := range nmt.Els {
					t := &nmt.Els[i]
					tnm := t.Type
					_, isb := BuiltinTypes[tnm]
					if !isb && !IsQualifiedType(tnm) {
						tnm = QualifyType(epkg, tnm)
					}
					t.Type = tnm
				}
				ty.Meths[mn] = nmt
			}
		}
	} else {
		ty.Els.CopyFrom(ety.Els)
		ty.Meths.CopyFrom(ety.Meths, false) // dest is newer
	}
	ty.Size[0] += len(ety.Els)
	// if TraceTypes {
	// 	fmt.Printf("Struct Type: %v inheriting from: %v\n", ty.Name, ety.Name)
	// }
}
