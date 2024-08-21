// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"
	"strings"

	"cogentcore.org/core/base/strcase"
	"cogentcore.org/core/types"
)

// Cmd represents a runnable command with configuration options.
// The type constraint is the type of the configuration
// information passed to the command.
type Cmd[T any] struct {
	// Func is the actual function that runs the command.
	// It takes configuration information and returns an error.
	Func func(T) error
	// Name is the name of the command.
	Name string
	// Doc is the documentation for the command.
	Doc string
	// Root is whether the command is the root command
	// (what is called when no subcommands are passed)
	Root bool
	// Icon is the icon of the command in the tool bar
	// when running in the GUI via clicore
	Icon string
	// SepBefore is whether to add a separator before the
	// command in the tool bar when running in the GUI via clicore
	SepBefore bool
	// SepAfter is whether to add a separator after the
	// command in the tool bar when running in the GUI via clicore
	SepAfter bool
}

// CmdOrFunc is a generic type constraint that represents either
// a [*Cmd] with the given config type or a command function that
// takes the given config type and returns an error.
type CmdOrFunc[T any] interface {
	*Cmd[T] | func(T) error
}

// cmdFromFunc returns a new [Cmd] object from the given function
// and any information specified on it using comment directives,
// which requires the use of [types].
func cmdFromFunc[T any](fun func(T) error) (*Cmd[T], error) {
	cmd := &Cmd[T]{
		Func: fun,
	}

	fn := types.FuncName(fun)

	// we need to get rid of package name and then convert to kebab
	strs := strings.Split(fn, ".")
	cfn := strs[len(strs)-1] // camel function name
	cmd.Name = strcase.ToKebab(cfn)

	if f := types.FuncByName(fn); f != nil {
		cmd.Doc = f.Doc
		for _, dir := range f.Directives {
			if dir.Tool != "cli" {
				continue
			}
			if dir.Directive != "cmd" {
				return cmd, fmt.Errorf("unrecognized comment directive %q (from comment %q)", dir.Directive, dir.String())
			}
			_, err := SetFromArgs(cmd, dir.Args, ErrNotFound)
			if err != nil {
				return cmd, fmt.Errorf("error setting command from directive arguments (from comment %q): %w", dir.String(), err)
			}
		}
		// we format the doc after the directives so that we have the up-to-date documentation and name
		cmd.Doc = types.FormatDoc(cmd.Doc, cfn, strcase.ToSentence(cmd.Name))
	}
	return cmd, nil
}

// cmdFromCmdOrFunc returns a new [Cmd] object from the given
// [CmdOrFunc] object, using [cmdFromFunc] if it is a function.
func cmdFromCmdOrFunc[T any, C CmdOrFunc[T]](cmd C) (*Cmd[T], error) {
	switch c := any(cmd).(type) {
	case *Cmd[T]:
		return c, nil
	case func(T) error:
		return cmdFromFunc(c)
	default:
		panic(fmt.Errorf("internal/programmer error: cli.CmdFromCmdOrFunc: impossible type %T for command %v", cmd, cmd))
	}
}

// CmdsFromCmdOrFuncs is a helper function that returns a slice
// of command objects from the given slice of [CmdOrFunc] objects,
// using [cmdFromCmdOrFunc].
func CmdsFromCmdOrFuncs[T any, C CmdOrFunc[T]](cmds []C) ([]*Cmd[T], error) {
	res := make([]*Cmd[T], len(cmds))
	for i, cmd := range cmds {
		cmd, err := cmdFromCmdOrFunc[T, C](cmd)
		if err != nil {
			return nil, err
		}
		res[i] = cmd
	}
	return res, nil
}

// AddCmd adds the given command to the given set of commands
// if there is not already a command with the same name in the
// set of commands. Also, if [Cmd.Root] is set to true on the
// passed command, and there are no other root commands in the
// given set of commands, the passed command will be made the
// root command; otherwise, it will be made not the root command.
func AddCmd[T any](cmds []*Cmd[T], cmd *Cmd[T]) []*Cmd[T] {
	hasCmd := false
	hasRoot := false
	for _, c := range cmds {
		if c.Name == cmd.Name {
			hasCmd = true
		}
		if c.Root {
			hasRoot = true
		}
	}
	if hasCmd {
		return cmds
	}
	cmd.Root = cmd.Root && !hasRoot // we must both want root and be able to take root
	cmds = append(cmds, cmd)
	return cmds
}
