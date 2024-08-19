// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package lsp contains types for the Language Server Protocol
// LSP: https://microsoft.github.io/language-server-protocol/specification
// and mappings from these elements into the token.Tokens types
// which are used internally in parse.
package lsp

//go:generate core generate

import (
	"cogentcore.org/core/parse/token"
)

// SymbolKind is the Language Server Protocol (LSP) SymbolKind, which
// we map onto the token.Tokens that are used internally.
type SymbolKind int32 //enums:enum

// SymbolKind is the list of SymbolKind items from LSP
const (
	NoSymbolKind SymbolKind = iota
	File                    // 1 in LSP
	Module
	Namespace
	Package
	Class
	Method
	Property
	Field
	Constructor
	Enum
	Interface
	Function
	Variable
	Constant
	String
	Number
	Boolean
	Array
	Object
	Key
	Null
	EnumMember
	Struct
	Event
	Operator
	TypeParameter // 26 in LSP
)

// SymbolKindTokenMap maps between symbols and token.Tokens
var SymbolKindTokenMap = map[SymbolKind]token.Tokens{
	Module:        token.NameModule,
	Namespace:     token.NameNamespace,
	Package:       token.NamePackage,
	Class:         token.NameClass,
	Method:        token.NameMethod,
	Property:      token.NameProperty,
	Field:         token.NameField,
	Constructor:   token.NameConstructor,
	Enum:          token.NameEnum,
	Interface:     token.NameInterface,
	Function:      token.NameFunction,
	Variable:      token.NameVar,
	Constant:      token.NameConstant,
	String:        token.LitStr,
	Number:        token.LitNum,
	Boolean:       token.LiteralBool,
	Array:         token.NameArray,
	Object:        token.NameObject,
	Key:           token.NameTag,
	Null:          token.None,
	EnumMember:    token.NameEnumMember,
	Struct:        token.NameStruct,
	Event:         token.NameEvent,
	Operator:      token.Operator,
	TypeParameter: token.NameTypeParam,
}

// TokenSymbolKindMap maps from tokens to LSP SymbolKind
var TokenSymbolKindMap map[token.Tokens]SymbolKind

func init() {
	for s, t := range SymbolKindTokenMap {
		TokenSymbolKindMap[t] = s
	}
}
