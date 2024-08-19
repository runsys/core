// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logx

import (
	"image/color"
	"log/slog"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/colors/matcolor"
	"github.com/muesli/termenv"
)

var (
	// UseColor is whether to use color in log messages. It is on by default.
	UseColor = true

	// ColorSchemeIsDark is whether the color scheme of the current terminal is dark-themed.
	// Its primary use is in [ColorScheme], and it should typically only be accessed via that.
	ColorSchemeIsDark = true
)

// ColorScheme returns the appropriate appropriate color scheme
// for terminal colors. It should be used instead of [colors.Scheme]
// for terminal colors because the theme (dark vs light) of the terminal
// could be different than that of the main app.
func ColorScheme() *matcolor.Scheme {
	if ColorSchemeIsDark {
		return &colors.Schemes.Dark
	}
	return &colors.Schemes.Light
}

// colorProfile is the termenv color profile, stored globally for convenience.
// It is set by [SetDefaultLogger] to [termenv.ColorProfile] if [UseColor] is true.
var colorProfile termenv.Profile

// InitColor sets up the terminal environment for color output. It is called automatically
// in an init function if UseColor is set to true. However, if you call a system command
// (ls, cp, etc), you need to call this function again.
func InitColor() {
	restoreFunc, err := termenv.EnableVirtualTerminalProcessing(termenv.DefaultOutput())
	if err != nil {
		slog.Warn("error enabling virtual terminal processing for colored output on Windows: %w", err)
	}
	_ = restoreFunc // TODO: figure out how to call this at the end of the program
	colorProfile = termenv.ColorProfile()
	ColorSchemeIsDark = termenv.HasDarkBackground()
}

// ApplyColor applies the given color to the given string
// and returns the resulting string. If [UseColor] is set
// to false, it just returns the string it was passed.
func ApplyColor(clr color.Color, str string) string {
	if !UseColor {
		return str
	}
	return termenv.String(str).Foreground(colorProfile.FromColor(clr)).String()
}

// LevelColor applies the color associated with the given level to the
// given string and returns the resulting string. If [UseColor] is set
// to false, it just returns the string it was passed.
func LevelColor(level slog.Level, str string) string {
	var clr color.RGBA
	switch level {
	case slog.LevelDebug:
		return DebugColor(str)
	case slog.LevelInfo:
		return InfoColor(str)
	case slog.LevelWarn:
		return WarnColor(str)
	case slog.LevelError:
		return ErrorColor(str)
	}
	return ApplyColor(clr, str)
}

// DebugColor applies the color associated with the debug level to
// the given string and returns the resulting string. If [UseColor] is set
// to false, it just returns the string it was passed.
func DebugColor(str string) string {
	return ApplyColor(colors.ToUniform(ColorScheme().Tertiary.Base), str)
}

// InfoColor applies the color associated with the info level to
// the given string and returns the resulting string. Because the
// color associated with the info level is just white/black, it just
// returns the given string, but it exists for API consistency.
func InfoColor(str string) string {
	return str
}

// WarnColor applies the color associated with the warn level to
// the given string and returns the resulting string. If [UseColor] is set
// to false, it just returns the string it was passed.
func WarnColor(str string) string {
	return ApplyColor(colors.ToUniform(ColorScheme().Warn.Base), str)
}

// ErrorColor applies the color associated with the error level to
// the given string and returns the resulting string. If [UseColor] is set
// to false, it just returns the string it was passed.
func ErrorColor(str string) string {
	return ApplyColor(colors.ToUniform(ColorScheme().Error.Base), str)
}

// SuccessColor applies the color associated with success to the
// given string and returns the resulting string. If [UseColor] is set
// to false, it just returns the string it was passed.
func SuccessColor(str string) string {
	return ApplyColor(colors.ToUniform(ColorScheme().Success.Base), str)
}

// CmdColor applies the color associated with terminal commands and arguments
// to the given string and returns the resulting string. If [UseColor] is set
// to false, it just returns the string it was passed.
func CmdColor(str string) string {
	return ApplyColor(colors.ToUniform(ColorScheme().Primary.Base), str)
}

// TitleColor applies the color associated with titles and section headers
// to the given string and returns the resulting string. If [UseColor] is set
// to false, it just returns the string it was passed.
func TitleColor(str string) string {
	return ApplyColor(colors.ToUniform(ColorScheme().Warn.Base), str)
}
