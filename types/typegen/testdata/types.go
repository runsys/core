// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testdata

import (
	"fmt"
	"image/color"

	"cogentcore.org/core/types"
)

// Person represents a person and their attributes.
// The zero value of a Person is not valid.
//
//tree:flagtype NodeFlags -field Flag
type Person struct { //core:embedder
	color.RGBA

	// Name is the name of the person
	Name string

	// Age is the age of the person
	Age int `json:"-"`

	// Type is the type of the person
	Type *types.Type

	unexportedField int

	// Nicknames are the nicknames of the person
	Nicknames []string
}

var _ = fmt.Stringer(&Person{})

//types:skip
func (p Person) String() string { return p.Name }

// Introduction returns an introduction for the person.
// It contains the name of the person and their age.
//
//core:toolbar -name ShowIntroduction -icon play -show-result -confirm
func (p *Person) Introduction() string { //types:add
	return fmt.Sprintf("%s is %d years old", p.Name, p.Age)
}

// Alert prints an alert with the given message
func Alert(msg string) {
	fmt.Println("Alert:", msg)
}

type (
	// BlockType is a type declared in a type block.
	BlockType struct{} //types:add

	// CommaFieldType is a type with inline comma fields.
	CommaFieldType struct { //types:add -setters
		A, B int
	}
)

// we test various type omitted arg combinations

func TypeOmittedArgs0(x, y float32)                {}
func TypeOmittedArgs1(x int, y, z struct{})        {}
func TypeOmittedArgs2(x, y, z int)                 {}
func TypeOmittedArgs3(x int, y, z bool, w float32) {}
func TypeOmittedArgs4(x, y, z string, w bool)      {}

// Setter is a type that can set a value.
type Setter interface { //types:add

	// Set sets the value.
	Set(value string) error

	// Clear clears the value.
	Clear() // A line comment.
}
