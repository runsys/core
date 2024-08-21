// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// note: FindFileOnPaths adapted from viper package https://github.com/spf13/viper
// Copyright (c) 2014 Steve Francia

package cli

import (
	"errors"
	"reflect"

	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/base/iox/tomlx"
	"cogentcore.org/core/base/reflectx"
)

// TODO(kai): this seems bad

// includer is an interface that facilitates processing
// include files in configuration objects.
type includer interface {
	// IncludesPtr returns a pointer to the "Includes []string"
	// field containing file(s) to include before processing
	// the current config file.
	IncludesPtr() *[]string
}

// includeStack returns the stack of include files in the natural
// order in which they are encountered (nil if none).
// Files should then be read in reverse order of the slice.
// Returns an error if any of the include files cannot be found on IncludePath.
// Does not alter cfg.
func includeStack(opts *Options, cfg includer) ([]string, error) {
	clone := reflect.New(reflectx.NonPointerType(reflect.TypeOf(cfg))).Interface().(includer)
	*clone.IncludesPtr() = *cfg.IncludesPtr()
	return includeStackImpl(opts, clone, nil)
}

// includeStackImpl implements IncludeStack, operating on cloned cfg
// todo: could use a more efficient method to just extract the include field..
func includeStackImpl(opts *Options, clone includer, includes []string) ([]string, error) {
	incs := *clone.IncludesPtr()
	ni := len(incs)
	if ni == 0 {
		return includes, nil
	}
	for i := ni - 1; i >= 0; i-- {
		includes = append(includes, incs[i]) // reverse order so later overwrite earlier
	}
	var errs []error
	for _, inc := range incs {
		*clone.IncludesPtr() = nil
		err := tomlx.OpenFiles(clone, fsx.FindFilesOnPaths(opts.IncludePaths, inc)...)
		if err == nil {
			includes, err = includeStackImpl(opts, clone, includes)
			if err != nil {
				errs = append(errs, err)
			}
		} else {
			errs = append(errs, err)
		}
	}
	return includes, errors.Join(errs...)
}
