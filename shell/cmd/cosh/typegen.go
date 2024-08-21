// Code generated by "core generate -add-types -add-funcs"; DO NOT EDIT.

package main

import (
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "main.Config", IDName: "config", Doc: "Config is the configuration information for the cosh cli.", Directives: []types.Directive{{Tool: "go", Directive: "generate", Args: []string{"core", "generate", "-add-types", "-add-funcs"}}}, Fields: []types.Field{{Name: "Input", Doc: "Input is the input file to run/compile.\nIf this is provided as the first argument,\nthen the program will exit after running,\nunless the Interactive mode is flagged."}, {Name: "Expr", Doc: "Expr is an optional expression to evaluate, which can be used\nin addition to the Input file to run, to execute commands\ndefined within that file for example, or as a command to run\nprior to starting interactive mode if no Input is specified."}, {Name: "Args", Doc: "Args is an optional list of arguments to pass in the run command.\nThese arguments will be turned into an \"args\" local variable in the shell.\nThese are automatically processed from any leftover arguments passed, so\nyou should not need to specify this flag manually."}, {Name: "Interactive", Doc: "Interactive runs the interactive command line after processing any input file.\nInteractive mode is the default mode for the run command unless an input file\nis specified."}}})

var _ = types.AddFunc(&types.Func{Name: "main.Run", Doc: "Run runs the specified cosh file. If no file is specified,\nit runs an interactive shell that allows the user to input cosh.", Directives: []types.Directive{{Tool: "cli", Directive: "cmd", Args: []string{"-root"}}}, Args: []string{"c"}, Returns: []string{"error"}})

var _ = types.AddFunc(&types.Func{Name: "main.Interactive", Doc: "Interactive runs an interactive shell that allows the user to input cosh.", Args: []string{"c", "in"}, Returns: []string{"error"}})

var _ = types.AddFunc(&types.Func{Name: "main.Build", Doc: "Build builds the specified input cosh file, or all .cosh files in the current\ndirectory if no input is specified, to corresponding .go file name(s).\nIf the file does not already contain a \"package\" specification, then\n\"package main; func main()...\" wrappers are added, which allows the same\ncode to be used in interactive and Go compiled modes.", Args: []string{"c"}, Returns: []string{"error"}})
