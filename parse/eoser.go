// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"github.com/goki/pi/lex"
	"github.com/goki/pi/token"
)

// Eoser is the end-of-statement (EOS) token finder, operating on the tokenized source,
// and breaking tokens into candidate chunks of "statements" terminated by the EOS
// This step bounds the outer-loop search for statement boundaries within which
// key tokens are searched for.
// Parentheses and square brackets are tracked to continue things across lines
// until properly terminated
type Eoser struct {
	Do        bool     `desc:"should we perform eos detection on this type of file?"`
	Eol       bool     `desc:"use end-of-line as a default EOS (python, Go)"`
	Semi      bool     `desc:"use semicolon (C)"`
	Backslash bool     `desc:"use backslash as a line continuer (python)"`
	State     EosState `view:"-" json:"-" xml:"-" desc:"state for eosing"`
}

// EosState is the state maintained for the Eosing process
type EosState struct {
	Ln       int            `desc:"line in lex tokens we're on"`
	Li       int            `desc:"index within line of lex tokens we're on"`
	Src      *lex.File      `desc:"file that we're operating on"`
	TokStack []token.Tokens `desc:"stack of lparen, lbrack tokens"`
	EosPos   []lex.Pos      `desc:"positions *in token coordinates* of the EOS markers generated"`
	Errs     lex.ErrorList  `desc:"any error messages accumulated during lexing specifically"`
}

// Init initializes state for a new pass of eosification
func (es *EosState) Init(src *lex.File) {
	es.Src = src
	es.Ln = 0
	es.Li = 0
	es.TokStack = es.TokStack[0:0]
	es.EosPos = es.EosPos[0:0]
}

// LexAt returns the Lex item at given position in current line
func (es *EosState) LexAt(idx int) lex.Lex {
	lx := es.Src.Lexs[es.Ln][idx]
	return lx
}

// Token returns the current token
func (es *EosState) Token() token.Tokens {
	lx := es.LexAt(es.Li)
	return lx.Token
}

// NextLine advances to next line
func (es *EosState) NextLine() {
	es.Ln++
	es.Li = 0
}

// InsertEOS inserts an EOS at the given token position
func (es *EosState) InsertEOS(at int) {
	plex := es.LexAt(at - 1)
	es.Src.Lexs[es.Ln].Insert(at, lex.Lex{token.EOS, plex.Ed, plex.Ed})
	es.EosPos = append(es.EosPos, lex.Pos{es.Ln, at})
}

// Error adds an eosing error at given position
func (eo *Eoser) Error(msg string) {
	clex := eo.State.LexAt(eo.State.Li - 1)
	eo.State.Errs.Add(lex.Pos{eo.State.Ln, clex.St}, eo.State.Src.Filename, "Eoser: "+msg)
}

// PushToken pushes a paren or bracket onto token stack
func (eo *Eoser) PushToken(tok token.Tokens) {
	eo.State.TokStack = append(eo.State.TokStack, tok)
}

// MismatchError reports a mismatch for given type of parentheses / bracket
func (eo *Eoser) MismatchError(tok token.Tokens) {
	switch tok {
	case token.PunctGpRParen:
		eo.Error("mismatching parentheses -- right paren ')' without matching left paren '('")
	case token.PunctGpRBrack:
		eo.Error("mismatching square brackets -- right bracket ']' without matching left bracket '['")
	}
}

// PopToken attempts to pop given token off of stack, generating error
// if it mismatches
func (eo *Eoser) PopToken(tok token.Tokens) {
	sz := len(eo.State.TokStack)
	if sz == 0 {
		eo.MismatchError(tok)
		return
	}
	cur := eo.State.TokStack[sz-1]
	eo.State.TokStack = eo.State.TokStack[:sz-1] // better to clear than keep even if err
	switch tok {
	case token.PunctGpRParen:
		if cur != token.PunctGpLParen {
			eo.MismatchError(tok)
		}
	case token.PunctGpRBrack:
		if cur != token.PunctGpLBrack {
			eo.MismatchError(tok)
		}
	}
}

// Perform EOS detection on given file
func (eo *Eoser) Eosify(src *lex.File) {
	eo.State.Init(src)
	nlines := src.NLines()
	for eo.State.Ln < nlines {
		sz := len(src.Lexs[eo.State.Ln])
		if sz == 0 {
			eo.State.NextLine()
			continue
		}
		if eo.State.Li >= sz {
			if eo.Eol && len(eo.State.TokStack) == 0 {
				eo.State.InsertEOS(eo.State.Li)
			}
			eo.State.NextLine()
			continue
		}
		tok := eo.State.Token()
		eo.State.Li++
		switch {
		case tok == token.PunctGpLParen || tok == token.PunctGpLBrack:
			eo.PushToken(tok)
			continue
		case tok == token.PunctGpRParen || tok == token.PunctGpRBrack:
			eo.PopToken(tok)
			continue
		case eo.Semi && tok == token.PunctSepSemicolon:
			eo.State.InsertEOS(eo.State.Li)
			continue
		case eo.Backslash && tok == token.PunctStrEsc && eo.State.Li >= sz:
			eo.State.NextLine() // prevent from marking as EOS
			continue
		}
	}
}
