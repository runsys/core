// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cosh defines convenient utility functions for
// use in the cosh shell, available with the cosh prefix.
package cosh

import (
	"os"
	"strings"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/base/slicesx"
	"cogentcore.org/core/base/stringsx"
)

// SplitLines returns a slice of given string split by lines
// with any extra whitespace trimmed for each line entry.
func SplitLines(s string) []string {
	sl := stringsx.SplitLines(s)
	for i, s := range sl {
		sl[i] = strings.TrimSpace(s)
	}
	return sl
}

// FileExists returns true if given file exists
func FileExists(path string) bool {
	ex := errors.Log1(fsx.FileExists(path))
	return ex
}

// WriteFile writes string to given file with standard permissions,
// logging any errors.
func WriteFile(filename, str string) error {
	err := os.WriteFile(filename, []byte(str), 0666)
	if err != nil {
		errors.Log(err)
	}
	return err
}

// ReadFile reads the string from the given file, logging any errors.
func ReadFile(filename string) string {
	str, err := os.ReadFile(filename)
	if err != nil {
		errors.Log(err)
	}
	return string(str)
}

// ReplaceInFile replaces all occurrences of given string with replacement
// in given file, rewriting the file.  Also returns the updated string.
func ReplaceInFile(filename, old, new string) string {
	str := ReadFile(filename)
	str = strings.ReplaceAll(str, old, new)
	WriteFile(filename, str)
	return str
}

// StringsToAnys converts a slice of strings to a slice of any,
// using slicesx.ToAny.  The interpreter cannot process generics
// yet, so this wrapper is needed.  Use for passing args to
// a command, for example.
func StringsToAnys(s []string) []any {
	return slicesx.ToAny(s)
}
