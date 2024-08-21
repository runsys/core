// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"cogentcore.org/core/base/fileinfo"
	"cogentcore.org/core/parse/lexer"
	"cogentcore.org/core/parse/parser"
	"cogentcore.org/core/parse/syms"
)

// FileState contains the full lexing and parsing state information for a given file.
// It is the master state record for everything that happens in parse.  One of these
// should be maintained for each file; texteditor.Buf has one as ParseState field.
//
// Separate State structs are maintained for each stage (Lexing, PassTwo, Parsing) and
// the final output of Parsing goes into the AST and Syms fields.
//
// The Src lexer.File field maintains all the info about the source file, and the basic
// tokenized version of the source produced initially by lexing and updated by the
// remaining passes.  It has everything that is maintained at a line-by-line level.
type FileState struct {

	// the source to be parsed -- also holds the full lexed tokens
	Src lexer.File `json:"-" xml:"-"`

	// state for lexing
	LexState lexer.State `json:"_" xml:"-"`

	// state for second pass nesting depth and EOS matching
	TwoState lexer.TwoState `json:"-" xml:"-"`

	// state for parsing
	ParseState parser.State `json:"-" xml:"-"`

	// ast output tree from parsing
	AST *parser.AST `json:"-" xml:"-"`

	// symbols contained within this file -- initialized at start of parsing and created by AddSymbol or PushNewScope actions.  These are then processed after parsing by the language-specific code, via Lang interface.
	Syms syms.SymMap `json:"-" xml:"-"`

	// External symbols that are entirely maintained in a language-specific way by the Lang interface code.  These are only here as a convenience and are not accessed in any way by the language-general parse code.
	ExtSyms syms.SymMap `json:"-" xml:"-"`

	// mutex protecting updates / reading of Syms symbols
	SymsMu sync.RWMutex `display:"-" json:"-" xml:"-"`

	// waitgroup for coordinating processing of other items
	WaitGp sync.WaitGroup `display:"-" json:"-" xml:"-"`

	// anonymous counter -- counts up
	AnonCtr int `display:"-" json:"-" xml:"-"`

	// path mapping cache -- for other files referred to by this file, this stores the full path associated with a logical path (e.g., in go, the logical import path -> local path with actual files) -- protected for access from any thread
	PathMap sync.Map `display:"-" json:"-" xml:"-"`
}

// Init initializes the file state
func (fs *FileState) Init() {
	// fmt.Println("fs init:", fs.Src.Filename)
	fs.AST = parser.NewAST()
	fs.LexState.Init()
	fs.TwoState.Init()
	fs.ParseState.Init(&fs.Src, fs.AST)
	fs.SymsMu.Lock()
	fs.Syms = make(syms.SymMap)
	fs.SymsMu.Unlock()
	fs.AnonCtr = 0
}

// NewFileState returns a new initialized file state
func NewFileState() *FileState {
	fs := &FileState{}
	fs.Init()
	return fs
}

func (fs *FileState) ClearAST() {
	fs.Syms.ClearAST()
	fs.ExtSyms.ClearAST()
}

func (fs *FileState) Destroy() {
	fs.ClearAST()
	fs.Syms = nil
	fs.ExtSyms = nil
	fs.AST.Destroy()
	fs.LexState.Init()
	fs.TwoState.Init()
	fs.ParseState.Destroy()
}

// SetSrc sets source to be parsed, and filename it came from, and also the
// base path for project for reporting filenames relative to
// (if empty, path to filename is used)
func (fs *FileState) SetSrc(src [][]rune, fname, basepath string, sup fileinfo.Known) {
	fs.Init()
	fs.Src.SetSrc(src, fname, basepath, sup)
	fs.LexState.Filename = fname
}

// LexAtEnd returns true if lexing state is now at end of source
func (fs *FileState) LexAtEnd() bool {
	return fs.LexState.Ln >= fs.Src.NLines()
}

// LexLine returns the lexing output for given line, combining comments and all other tokens
// and allocating new memory using clone
func (fs *FileState) LexLine(ln int) lexer.Line {
	return fs.Src.LexLine(ln)
}

// LexLineString returns a string rep of the current lexing output for the current line
func (fs *FileState) LexLineString() string {
	return fs.LexState.LineString()
}

// LexNextSrcLine returns the next line of source that the lexer is currently at
func (fs *FileState) LexNextSrcLine() string {
	return fs.LexState.NextSrcLine()
}

// LexHasErrs returns true if there were errors from lexing
func (fs *FileState) LexHasErrs() bool {
	return len(fs.LexState.Errs) > 0
}

// LexErrReport returns a report of all the lexing errors -- these should only
// occur during development of lexer so we use a detailed report format
func (fs *FileState) LexErrReport() string {
	return fs.LexState.Errs.Report(0, fs.Src.BasePath, true, true)
}

// PassTwoHasErrs returns true if there were errors from pass two processing
func (fs *FileState) PassTwoHasErrs() bool {
	return len(fs.TwoState.Errs) > 0
}

// PassTwoErrString returns all the pass two errors as a string -- these should
// only occur during development so we use a detailed report format
func (fs *FileState) PassTwoErrReport() string {
	return fs.TwoState.Errs.Report(0, fs.Src.BasePath, true, true)
}

// ParseAtEnd returns true if parsing state is now at end of source
func (fs *FileState) ParseAtEnd() bool {
	return fs.ParseState.AtEofNext()
}

// ParseNextSrcLine returns the next line of source that the parser is currently at
func (fs *FileState) ParseNextSrcLine() string {
	return fs.ParseState.NextSrcLine()
}

// ParseHasErrs returns true if there were errors from parsing
func (fs *FileState) ParseHasErrs() bool {
	return len(fs.ParseState.Errs) > 0
}

// ParseErrReport returns at most 10 parsing errors in end-user format, sorted
func (fs *FileState) ParseErrReport() string {
	fs.ParseState.Errs.Sort()
	return fs.ParseState.Errs.Report(10, fs.Src.BasePath, true, false)
}

// ParseErrReportAll returns all parsing errors in end-user format, sorted
func (fs *FileState) ParseErrReportAll() string {
	fs.ParseState.Errs.Sort()
	return fs.ParseState.Errs.Report(0, fs.Src.BasePath, true, false)
}

// ParseErrReportDetailed returns at most 10 parsing errors in detailed format, sorted
func (fs *FileState) ParseErrReportDetailed() string {
	fs.ParseState.Errs.Sort()
	return fs.ParseState.Errs.Report(10, fs.Src.BasePath, true, true)
}

// RuleString returns the rule info for entire source -- if full
// then it includes the full stack at each point -- otherwise just the top
// of stack
func (fs *FileState) ParseRuleString(full bool) string {
	return fs.ParseState.RuleString(full)
}

////////////////////////////////////////////////////////////////////////
//  Syms symbol processing support

// FindNameScoped looks for given symbol name within given map first
// (if non nil) and then in fs.Syms and ExtSyms maps,
// and any children on those global maps that are of subcategory
// token.NameScope (i.e., namespace, module, package, library)
func (fs *FileState) FindNameScoped(nm string, scope syms.SymMap) (*syms.Symbol, bool) {
	var sy *syms.Symbol
	has := false
	if scope != nil {
		sy, has = scope.FindName(nm)
		if has {
			return sy, true
		}
	}
	sy, has = fs.Syms.FindNameScoped(nm)
	if has {
		return sy, true
	}
	sy, has = fs.ExtSyms.FindNameScoped(nm)
	if has {
		return sy, true
	}
	return nil, false
}

// FindChildren fills out map with direct children of given symbol
// If seed is non-empty it is used as a prefix for filtering children names.
// Returns false if no children were found.
func (fs *FileState) FindChildren(sym *syms.Symbol, seed string, scope syms.SymMap, kids *syms.SymMap) bool {
	if len(sym.Children) == 0 {
		if sym.Type != "" {
			typ, got := fs.FindNameScoped(sym.NonPtrTypeName(), scope)
			if got {
				sym = typ
			} else {
				return false
			}
		}
	}
	if seed != "" {
		sym.Children.FindNamePrefix(seed, kids)
	} else {
		kids.CopyFrom(sym.Children, true) // src is newer
	}
	return len(*kids) > 0
}

// FindAnyChildren fills out map with either direct children of given symbol
// or those of the type of this symbol -- useful for completion.
// If seed is non-empty it is used as a prefix for filtering children names.
// Returns false if no children were found.
func (fs *FileState) FindAnyChildren(sym *syms.Symbol, seed string, scope syms.SymMap, kids *syms.SymMap) bool {
	if len(sym.Children) == 0 {
		if sym.Type != "" {
			typ, got := fs.FindNameScoped(sym.NonPtrTypeName(), scope)
			if got {
				sym = typ
			} else {
				return false
			}
		}
	}
	if seed != "" {
		sym.Children.FindNamePrefixRecursive(seed, kids)
	} else {
		kids.CopyFrom(sym.Children, true) // src is newer
	}
	return len(*kids) > 0
}

// FindNamePrefixScoped looks for given symbol name prefix within given map first
// (if non nil) and then in fs.Syms and ExtSyms maps,
// and any children on those global maps that are of subcategory
// token.NameScope (i.e., namespace, module, package, library)
// adds to given matches map (which can be nil), for more efficient recursive use
func (fs *FileState) FindNamePrefixScoped(seed string, scope syms.SymMap, matches *syms.SymMap) {
	lm := len(*matches)
	if scope != nil {
		scope.FindNamePrefixRecursive(seed, matches)
	}
	if len(*matches) != lm {
		return
	}
	fs.Syms.FindNamePrefixScoped(seed, matches)
	if len(*matches) != lm {
		return
	}
	fs.ExtSyms.FindNamePrefixScoped(seed, matches)
}

// NextAnonName returns the next anonymous name for this file, using counter here
// and given context name (e.g., package name)
func (fs *FileState) NextAnonName(ctxt string) string {
	fs.AnonCtr++
	fn := filepath.Base(fs.Src.Filename)
	ext := filepath.Ext(fn)
	if ext != "" {
		fn = strings.TrimSuffix(fn, ext)
	}
	return fmt.Sprintf("anon_%s_%d", ctxt, fs.AnonCtr)
}

// PathMapLoad does a mutex-protected load of PathMap for given string,
// returning value and true if found
func (fs *FileState) PathMapLoad(path string) (string, bool) {
	fabs, ok := fs.PathMap.Load(path)
	fs.PathMap.Load(path)
	if ok {
		return fabs.(string), ok
	}
	return "", ok
}

// PathMapStore does a mutex-protected store of abs path for given path key
func (fs *FileState) PathMapStore(path, abs string) {
	fs.PathMap.Store(path, abs)
}
