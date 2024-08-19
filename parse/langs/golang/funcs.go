// Copyright (c) 2020, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package golang

import (
	"fmt"
	"unicode"

	"cogentcore.org/core/parse"
	"cogentcore.org/core/parse/parser"
	"cogentcore.org/core/parse/syms"
	"cogentcore.org/core/parse/token"
)

// TypeMeths gathers method types from the type symbol's children
func (gl *GoLang) TypeMeths(fs *parse.FileState, pkg *syms.Symbol, ty *syms.Type) {
	_, tnm := SplitType(ty.Name)
	tsym, got := pkg.Children.FindNameScoped(tnm)
	if !got {
		if !unicode.IsLower(rune(tnm[0])) && TraceTypes {
			fmt.Printf("TypeMeths: error -- did NOT get type sym: %v in pkg: %v\n", tnm, pkg.Name)
		}
		return
	}
	for _, sy := range tsym.Children {
		if sy.Kind.SubCat() != token.NameFunction || sy.Ast == nil {
			continue
		}

		fty := gl.FuncTypeFromAst(fs, pkg, sy.Ast.(*parser.Ast), nil)
		if fty != nil {
			fty.Kind = syms.Method
			fty.Name = sy.Name
			fty.Filename = sy.Filename
			fty.Region = sy.Region
			ty.Meths.Add(fty)
			// if TraceTypes {
			// 	fmt.Printf("TypeMeths: Added method: %v\n", fty)
			// }
		} else {
			if TraceTypes {
				fmt.Printf("TypeMeths: method failed: %v\n", sy.Name)
			}
		}
	}
}

// NamesFromAst returns a slice of name(s) from namelist nodes
func (gl *GoLang) NamesFromAst(fs *parse.FileState, pkg *syms.Symbol, ast *parser.Ast, idx int) []string {
	sast := ast.ChildAst(idx)
	if sast == nil {
		if TraceTypes {
			fmt.Printf("TraceTypes: could not find child 0 on ast %v", ast)
		}
		return nil
	}
	var sary []string
	if sast.HasChildren() {
		for i := range sast.Children {
			sary = append(sary, gl.NamesFromAst(fs, pkg, sast, i)...)
		}
	} else {
		sary = append(sary, sast.Src)
	}
	return sary
}

// FuncTypeFromAst initializes a function type from ast -- type can either be anon
// or a named type -- if anon then the name is the full type signature without param names
func (gl *GoLang) FuncTypeFromAst(fs *parse.FileState, pkg *syms.Symbol, ast *parser.Ast, fty *syms.Type) *syms.Type {
	// ast.WriteTree(os.Stdout, 0)

	if ast == nil || !ast.HasChildren() {
		return nil
	}
	pars := ast.ChildAst(0)
	if pars == nil {
		if TraceTypes {
			fmt.Printf("TraceTypes: could not find child 0 on ast %v", ast)
		}
		return nil
	}
	if fty == nil {
		fty = &syms.Type{}
		fty.Kind = syms.Func
	}
	poff := 0
	isMeth := false
	if pars.Name == "MethRecvName" && len(ast.Children) > 2 {
		isMeth = true
		rcv := pars.Children[0].(*parser.Ast)
		rtyp := pars.Children[1].(*parser.Ast)
		fty.Els.Add(rcv.Src, rtyp.Src)
		poff = 2
		pars = ast.ChildAst(2)
	} else if pars.Name == "Name" && len(ast.Children) > 1 {
		poff = 1
		pars = ast.ChildAst(1)
	}
	npars := len(pars.Children)
	var sigpars *parser.Ast
	if npars > 0 && (pars.Name == "SigParams" || pars.Name == "SigParamsResult") {
		if ps := pars.ChildAst(0); ps == nil {
			sigpars = pars
			pars = ps
			npars = len(pars.Children)
		} else {
			npars = 0 // not really
		}
	}
	if npars > 0 {
		gl.ParamsFromAst(fs, pkg, pars, fty, "param")
		npars = len(fty.Els) // how many we added -- auto-includes receiver for method
	} else {
		if isMeth {
			npars = 1
		}
	}
	nrvals := 0
	if sigpars != nil && len(sigpars.Children) >= 2 {
		rvals := sigpars.ChildAst(1)
		gl.RvalsFromAst(fs, pkg, rvals, fty)
		nrvals = len(fty.Els) - npars // how many we added..
	} else if poff < 2 && (len(ast.Children) >= poff+2) {
		rvals := ast.ChildAst(poff + 1)
		gl.RvalsFromAst(fs, pkg, rvals, fty)
		nrvals = len(fty.Els) - npars // how many we added..
	}
	fty.Size = []int{npars, nrvals}
	return fty
}

// ParamsFromAst sets params as Els for given function type (also for return types)
func (gl *GoLang) ParamsFromAst(fs *parse.FileState, pkg *syms.Symbol, pars *parser.Ast, fty *syms.Type, name string) {
	npars := len(pars.Children)
	var pnames []string // param names that all share same type
	for i := 0; i < npars; i++ {
		par := pars.Children[i].(*parser.Ast)
		psz := len(par.Children)
		if par.Name == "ParType" && psz == 1 {
			ptypa := par.Children[0].(*parser.Ast)
			if ptypa.Name == "TypeNm" { // could be multiple args with same type or a separate type-only arg
				if ptl, _ := gl.FindTypeName(par.Src, fs, pkg); ptl != nil {
					fty.Els.Add(fmt.Sprintf("%s_%v", name, i), par.Src)
					continue
				}
				pnames = append(pnames, par.Src) // add to later type
			} else {
				ptyp, ok := gl.SubTypeFromAst(fs, pkg, par, 0)
				if ok {
					pnsz := len(pnames)
					if pnsz > 0 {
						for _, pn := range pnames {
							fty.Els.Add(pn, ptyp.Name)
						}
					}
					fty.Els.Add(fmt.Sprintf("%s_%v", name, i), ptyp.Name)
					continue
				}
				pnames = nil
			}
		} else if psz == 2 { // ParName
			pnm := par.Children[0].(*parser.Ast)
			ptyp, ok := gl.SubTypeFromAst(fs, pkg, par, 1)
			if ok {
				pnsz := len(pnames)
				if pnsz > 0 {
					for _, pn := range pnames {
						fty.Els.Add(pn, ptyp.Name)
					}
				}
				fty.Els.Add(pnm.Src, ptyp.Name)
				continue
			}
			pnames = nil
		}
	}
}

// RvalsFromAst sets return value(s) as Els for given function type
func (gl *GoLang) RvalsFromAst(fs *parse.FileState, pkg *syms.Symbol, rvals *parser.Ast, fty *syms.Type) {
	if rvals.Name == "Block" { // todo: maybe others
		return
	}
	nrvals := len(rvals.Children)
	if nrvals == 1 { // single rval, unnamed, has type directly..
		rval := rvals.ChildAst(0)
		if rval.Name != "ParName" {
			nrvals = 1
			rtyp, ok := gl.SubTypeFromAst(fs, pkg, rvals, 0)
			if ok {
				fty.Els.Add("rval", rtyp.Name)
				return
			}
		}
	}
	gl.ParamsFromAst(fs, pkg, rvals, fty, "rval")
}
