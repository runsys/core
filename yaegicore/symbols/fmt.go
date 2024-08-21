// Code generated by 'yaegi extract fmt'. DO NOT EDIT.

//go:build go1.22
// +build go1.22

package symbols

import (
	"fmt"
	"reflect"
)

func init() {
	Symbols["fmt/fmt"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Append":       reflect.ValueOf(fmt.Append),
		"Appendf":      reflect.ValueOf(fmt.Appendf),
		"Appendln":     reflect.ValueOf(fmt.Appendln),
		"Errorf":       reflect.ValueOf(fmt.Errorf),
		"FormatString": reflect.ValueOf(fmt.FormatString),
		"Fprint":       reflect.ValueOf(fmt.Fprint),
		"Fprintf":      reflect.ValueOf(fmt.Fprintf),
		"Fprintln":     reflect.ValueOf(fmt.Fprintln),
		"Fscan":        reflect.ValueOf(fmt.Fscan),
		"Fscanf":       reflect.ValueOf(fmt.Fscanf),
		"Fscanln":      reflect.ValueOf(fmt.Fscanln),
		"Print":        reflect.ValueOf(fmt.Print),
		"Printf":       reflect.ValueOf(fmt.Printf),
		"Println":      reflect.ValueOf(fmt.Println),
		"Scan":         reflect.ValueOf(fmt.Scan),
		"Scanf":        reflect.ValueOf(fmt.Scanf),
		"Scanln":       reflect.ValueOf(fmt.Scanln),
		"Sprint":       reflect.ValueOf(fmt.Sprint),
		"Sprintf":      reflect.ValueOf(fmt.Sprintf),
		"Sprintln":     reflect.ValueOf(fmt.Sprintln),
		"Sscan":        reflect.ValueOf(fmt.Sscan),
		"Sscanf":       reflect.ValueOf(fmt.Sscanf),
		"Sscanln":      reflect.ValueOf(fmt.Sscanln),

		// type definitions
		"Formatter":  reflect.ValueOf((*fmt.Formatter)(nil)),
		"GoStringer": reflect.ValueOf((*fmt.GoStringer)(nil)),
		"ScanState":  reflect.ValueOf((*fmt.ScanState)(nil)),
		"Scanner":    reflect.ValueOf((*fmt.Scanner)(nil)),
		"State":      reflect.ValueOf((*fmt.State)(nil)),
		"Stringer":   reflect.ValueOf((*fmt.Stringer)(nil)),

		// interface wrapper definitions
		"_Formatter":  reflect.ValueOf((*_fmt_Formatter)(nil)),
		"_GoStringer": reflect.ValueOf((*_fmt_GoStringer)(nil)),
		"_ScanState":  reflect.ValueOf((*_fmt_ScanState)(nil)),
		"_Scanner":    reflect.ValueOf((*_fmt_Scanner)(nil)),
		"_State":      reflect.ValueOf((*_fmt_State)(nil)),
		"_Stringer":   reflect.ValueOf((*_fmt_Stringer)(nil)),
	}
}

// _fmt_Formatter is an interface wrapper for Formatter type
type _fmt_Formatter struct {
	IValue  interface{}
	WFormat func(f fmt.State, verb rune)
}

func (W _fmt_Formatter) Format(f fmt.State, verb rune) { W.WFormat(f, verb) }

// _fmt_GoStringer is an interface wrapper for GoStringer type
type _fmt_GoStringer struct {
	IValue    interface{}
	WGoString func() string
}

func (W _fmt_GoStringer) GoString() string { return W.WGoString() }

// _fmt_ScanState is an interface wrapper for ScanState type
type _fmt_ScanState struct {
	IValue      interface{}
	WRead       func(buf []byte) (n int, err error)
	WReadRune   func() (r rune, size int, err error)
	WSkipSpace  func()
	WToken      func(skipSpace bool, f func(rune) bool) (token []byte, err error)
	WUnreadRune func() error
	WWidth      func() (wid int, ok bool)
}

func (W _fmt_ScanState) Read(buf []byte) (n int, err error)      { return W.WRead(buf) }
func (W _fmt_ScanState) ReadRune() (r rune, size int, err error) { return W.WReadRune() }
func (W _fmt_ScanState) SkipSpace()                              { W.WSkipSpace() }
func (W _fmt_ScanState) Token(skipSpace bool, f func(rune) bool) (token []byte, err error) {
	return W.WToken(skipSpace, f)
}
func (W _fmt_ScanState) UnreadRune() error         { return W.WUnreadRune() }
func (W _fmt_ScanState) Width() (wid int, ok bool) { return W.WWidth() }

// _fmt_Scanner is an interface wrapper for Scanner type
type _fmt_Scanner struct {
	IValue interface{}
	WScan  func(state fmt.ScanState, verb rune) error
}

func (W _fmt_Scanner) Scan(state fmt.ScanState, verb rune) error { return W.WScan(state, verb) }

// _fmt_State is an interface wrapper for State type
type _fmt_State struct {
	IValue     interface{}
	WFlag      func(c int) bool
	WPrecision func() (prec int, ok bool)
	WWidth     func() (wid int, ok bool)
	WWrite     func(b []byte) (n int, err error)
}

func (W _fmt_State) Flag(c int) bool                   { return W.WFlag(c) }
func (W _fmt_State) Precision() (prec int, ok bool)    { return W.WPrecision() }
func (W _fmt_State) Width() (wid int, ok bool)         { return W.WWidth() }
func (W _fmt_State) Write(b []byte) (n int, err error) { return W.WWrite(b) }

// _fmt_Stringer is an interface wrapper for Stringer type
type _fmt_Stringer struct {
	IValue  interface{}
	WString func() string
}

func (W _fmt_Stringer) String() string {
	if W.WString == nil {
		return ""
	}
	return W.WString()
}
