// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"

	"cogentcore.org/core/base/logx"
	"cogentcore.org/core/base/strcase"
	"cogentcore.org/core/types"
)

// indent is the value used for indentation in [usage].
var indent = "    "

// usage returns a usage string based on the given options,
// configuration struct, current command, and available commands.
// It contains [AppAbout], a list of commands and their descriptions,
// and a list of flags and their descriptions, scoped based on the
// current command and its associated commands and configuration.
// The resulting string contains color escape codes.
func usage[T any](opts *Options, cfg T, cmd string, cmds ...*Cmd[T]) string {
	var b strings.Builder
	if cmd == "" {
		if opts.AppAbout != "" {
			b.WriteString("\n" + opts.AppAbout + "\n\n")
		}
	} else {
		gotCmd := false
		for _, c := range cmds {
			if c.Name == cmd {
				if c.Doc != "" {
					b.WriteString("\n" + c.Doc + "\n\n")
				}
				gotCmd = true
				break
			}
		}
		if !gotCmd {
			fmt.Println(logx.CmdColor(cmdName()+" help") + logx.ErrorColor(fmt.Sprintf(" failed: command %q not found", cmd)))
			os.Exit(1)
		}
	}

	fs := &fields{}
	addFields(cfg, fs, cmd)

	cmdName := cmdName()
	if cmd != "" {
		cmdName += " " + cmd
	}
	b.WriteString(logx.TitleColor("Usage:\n") + indent + logx.CmdColor(cmdName+" "))

	posArgStrs := []string{}

	for _, kv := range fs.Order {
		v := kv.Value
		f := v.Field

		posArgTag, ok := f.Tag.Lookup("posarg")
		if ok {
			ui := uint64(0)
			if posArgTag == "all" || posArgTag == "leftover" {
				ui = uint64(len(posArgStrs))
			} else {
				var err error
				ui, err = strconv.ParseUint(posArgTag, 10, 64)
				if err != nil {
					slog.Error("programmer error: invalid value for posarg struct tag", "field", f.Name, "posArgTag", posArgTag, "err", err)
				}
			}
			// if the slice isn't big enough, grow it to fit this posarg
			if ui >= uint64(len(posArgStrs)) {
				posArgStrs = slices.Grow(posArgStrs, len(posArgStrs)-int(ui)+1) // increase capacity
				posArgStrs = posArgStrs[:ui+1]                                  // extend to capacity
			}
			nm := strcase.ToKebab(v.Names[0])
			req, has := f.Tag.Lookup("required")
			if req == "+" || req == "true" || !has { // default is required, so !has => required
				posArgStrs[ui] = logx.CmdColor("<" + nm + ">")
			} else {
				posArgStrs[ui] = logx.SuccessColor("[" + nm + "]")
			}
		}
	}
	b.WriteString(strings.Join(posArgStrs, " "))
	if len(posArgStrs) > 0 {
		b.WriteString(" ")
	}
	b.WriteString(logx.SuccessColor("[flags]\n"))

	commandUsage(&b, cmdName, cmd, cmds...)

	b.WriteString(logx.TitleColor("\nFlags:\n") + indent + logx.TitleColor("Flags are case-insensitive, can be in kebab-case, snake_case,\n"))
	b.WriteString(indent + logx.TitleColor("or CamelCase, and can have one or two leading dashes. Use a\n"))
	b.WriteString(indent + logx.TitleColor("\"no\" prefix to turn off a bool flag.\n\n"))

	// add meta ones (help, config, verbose, etc) first
	mcfields := &fields{}
	addMetaConfigFields(mcfields)
	flagUsage(mcfields, &b)

	flagUsage(fs, &b)
	return b.String()
}

// commandUsage adds the command usage info for the given commands to the
// given [strings.Builder].
// It also takes the full name of our command as it appears in the terminal (cmdName),
// (eg: "core build"), and the name of the command we are running (eg: "build").
//
// To be a command that is included in the usage, we must be one command
// nesting depth (subcommand) deeper than the current command (ie, if we
// are on "x", we can see usage for commands of the form "x y"), and all
// of our commands must be consistent with the current command. For example,
// "" could generate usage for "help", "build", and "run", and "mod" could
// generate usage for "mod init", "mod tidy", and "mod edit". This ensures
// that only relevant commands are shown in the usage.
func commandUsage[T any](b *strings.Builder, cmdName string, cmd string, cmds ...*Cmd[T]) {
	acmds := []*Cmd[T]{}           // actual commands we care about
	var rcmd *Cmd[T]               // root command
	cmdstrs := strings.Fields(cmd) // subcommand strings in passed command

	// need this label so that we can continue outer loop when we have non-matching cmdstr
outer:
	for _, c := range cmds {
		cstrs := strings.Fields(c.Name)   // subcommand strings in command we are checking
		if len(cstrs) != len(cmdstrs)+1 { // we must be one deeper
			continue
		}
		for i, cmdstr := range cmdstrs {
			if cmdstr != cstrs[i] { // every subcommand so far must match
				continue outer
			}
		}
		if c.Root {
			rcmd = c
		} else if c.Name != cmd { // if it is the same subcommand we are already on, we handle it above in main Usage
			acmds = append(acmds, c)
		}
	}

	if len(acmds) != 0 {
		b.WriteString(indent + logx.CmdColor(cmdName+" <subcommand> ") + logx.SuccessColor("[flags]\n"))
	}

	if rcmd != nil {
		b.WriteString(logx.TitleColor("\nDefault command:\n"))
		b.WriteString(indent + logx.CmdColor(rcmd.Name) + "\n" + indent + indent + strings.ReplaceAll(rcmd.Doc, "\n", "\n"+indent+indent) + "\n") // need to put two indents on every newline for formatting
	}

	if len(acmds) == 0 && cmd != "" { // nothing to do
		return
	}

	b.WriteString(logx.TitleColor("\nSubcommands:\n"))

	// if we are in root, we also add help
	if cmd == "" {
		b.WriteString(indent + logx.CmdColor("help") + "\n" + indent + indent + "Help shows usage information for a command\n")
	}

	for _, c := range acmds {
		b.WriteString(indent + logx.CmdColor(c.Name))
		if c.Doc != "" {
			// we only want the first paragraph of text for subcommand usage; after that is where more specific details can go
			doc, _, _ := strings.Cut(c.Doc, "\n\n")
			b.WriteString("\n" + indent + indent + strings.ReplaceAll(doc, "\n", "\n"+indent+indent)) // need to put two indents on every newline for formatting
		}
		b.WriteString("\n")
	}
}

// flagUsage adds the flag usage info for the given fields
// to the given [strings.Builder].
func flagUsage(fields *fields, b *strings.Builder) {
	for _, kv := range fields.Order {
		f := kv.Value
		b.WriteString(indent)
		for i, name := range f.Names {
			b.WriteString(logx.CmdColor("-" + strcase.ToKebab(name)))
			if i != len(f.Names)-1 {
				b.WriteString(", ")
			}
		}
		b.WriteString(" " + logx.SuccessColor(f.Field.Type.String()))
		b.WriteString("\n")
		field := types.GetField(f.Struct, f.Field.Name)
		if field != nil {
			b.WriteString(indent + indent + strings.ReplaceAll(field.Doc, "\n", "\n"+indent+indent)) // need to put two indents on every newline for formatting
		}
		def, ok := f.Field.Tag.Lookup("default")
		if ok && def != "" {
			b.WriteString(fmt.Sprintf(" (default: %s)", def))
		}
		b.WriteString("\n")
	}
}
