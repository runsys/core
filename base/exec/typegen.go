// Code generated by "core generate"; DO NOT EDIT.

package exec

import (
	"io"

	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/base/exec.Config", IDName: "config", Doc: "Config contains the configuration information that\ncontrols the behavior of exec. It is passed to most\nhigh-level functions, and a default version of it\ncan be easily constructed using [DefaultConfig].", Directives: []types.Directive{{Tool: "types", Directive: "add", Args: []string{"-setters"}}}, Fields: []types.Field{{Name: "Buffer", Doc: "Buffer is whether to buffer the output of Stdout and Stderr,\nwhich is necessary for the correct printing of commands and output\nwhen there is an error with a command, and for correct coloring\non Windows. Therefore, it should be kept at the default value of\ntrue in most cases, except for when a command will run for a log\ntime and print output throughout (eg: a log command)."}, {Name: "PrintOnly", Doc: "PrintOnly is whether to only print commands that would be run and\nnot actually run them. It can be used, for example, for safely testing\nan app."}, {Name: "Dir", Doc: "The directory to execute commands in. If it is unset,\ncommands are run in the current directory."}, {Name: "Env", Doc: "Env contains any additional environment variables specified.\nThe current environment variables will also be passed to the\ncommand, but they will be overridden by any variables here\nif there are conflicts."}, {Name: "Echo", Doc: "Echo is the writer for echoing the command string to.\nIt can be set to nil to disable echoing."}, {Name: "StdIO", Doc: "Standard Input / Output management"}}})

// SetBuffer sets the [Config.Buffer]:
// Buffer is whether to buffer the output of Stdout and Stderr,
// which is necessary for the correct printing of commands and output
// when there is an error with a command, and for correct coloring
// on Windows. Therefore, it should be kept at the default value of
// true in most cases, except for when a command will run for a log
// time and print output throughout (eg: a log command).
func (t *Config) SetBuffer(v bool) *Config { t.Buffer = v; return t }

// SetPrintOnly sets the [Config.PrintOnly]:
// PrintOnly is whether to only print commands that would be run and
// not actually run them. It can be used, for example, for safely testing
// an app.
func (t *Config) SetPrintOnly(v bool) *Config { t.PrintOnly = v; return t }

// SetDir sets the [Config.Dir]:
// The directory to execute commands in. If it is unset,
// commands are run in the current directory.
func (t *Config) SetDir(v string) *Config { t.Dir = v; return t }

// SetEcho sets the [Config.Echo]:
// Echo is the writer for echoing the command string to.
// It can be set to nil to disable echoing.
func (t *Config) SetEcho(v io.Writer) *Config { t.Echo = v; return t }

// SetStdIO sets the [Config.StdIO]:
// Standard Input / Output management
func (t *Config) SetStdIO(v StdIO) *Config { t.StdIO = v; return t }
