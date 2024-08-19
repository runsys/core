// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package iox provides boilerplate wrapper functions for the Go standard
// io functions to Read, Open, Write, and Save, with implementations for
// commonly used encoding formats.
package iox

import (
	"bufio"
	"bytes"
	"io"
	"io/fs"
	"os"
)

// Decoder is an interface for standard decoder types
type Decoder interface {
	// Decode decodes from io.Reader specified at creation
	Decode(v any) error
}

// DecoderFunc is a function that creates a new Decoder for given reader
type DecoderFunc func(r io.Reader) Decoder

// NewDecoderFunc returns a DecoderFunc for a specific Decoder type
func NewDecoderFunc[T Decoder](f func(r io.Reader) T) DecoderFunc {
	return func(r io.Reader) Decoder { return f(r) }
}

// Open reads the given object from the given filename using the given [DecoderFunc]
func Open(v any, filename string, f DecoderFunc) error {
	fp, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fp.Close()
	return Read(v, bufio.NewReader(fp), f)
}

// OpenFiles reads the given object from the given filenames using the given [DecoderFunc]
func OpenFiles(v any, filenames []string, f DecoderFunc) error {
	for _, file := range filenames {
		err := Open(v, file, f)
		if err != nil {
			return err
		}
	}
	return nil
}

// OpenFS reads the given object from the given filename using the given [DecoderFunc],
// using the given [fs.FS] filesystem (e.g., for embed files)
func OpenFS(v any, fsys fs.FS, filename string, f DecoderFunc) error {
	fp, err := fsys.Open(filename)
	if err != nil {
		return err
	}
	defer fp.Close()
	return Read(v, bufio.NewReader(fp), f)
}

// OpenFilesFS reads the given object from the given filenames using the given [DecoderFunc],
// using the given [fs.FS] filesystem (e.g., for embed files)
func OpenFilesFS(v any, fsys fs.FS, filenames []string, f DecoderFunc) error {
	for _, file := range filenames {
		err := OpenFS(v, fsys, file, f)
		if err != nil {
			return err
		}
	}
	return nil
}

// Read reads the given object from the given reader,
// using the given [DecoderFunc]
func Read(v any, reader io.Reader, f DecoderFunc) error {
	d := f(reader)
	return d.Decode(v)
}

// ReadBytes reads the given object from the given bytes,
// using the given [DecoderFunc]
func ReadBytes(v any, data []byte, f DecoderFunc) error {
	b := bytes.NewBuffer(data)
	return Read(v, b, f)
}
