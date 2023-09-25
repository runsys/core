// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grog

import (
	"image/color"
	"log/slog"

	"github.com/muesli/termenv"
	"goki.dev/matcolor"
)

// UseColor is whether to use color in log messages.
// It is on by default.
var UseColor = true

// colorProfile is the termenv color profile, stored globally for convenience.
// It is set by [SetDefaultLogger] to [termenv.ColorProfile] if [UseColor] is true.
var colorProfile termenv.Profile

// ApplyColor applies the given color to the given string
// and returns the resulting string. If [UseColor] is set
// to false, it just returns the string it was passed.
func ApplyColor(clr color.RGBA, str string) string {
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
		clr = matcolor.TheScheme.Secondary
	case slog.LevelInfo:
		clr = matcolor.TheScheme.Primary
	case slog.LevelWarn:
		clr = matcolor.TheScheme.Tertiary
	case slog.LevelError:
		clr = matcolor.TheScheme.Error
	}
	return ApplyColor(clr, str)
}

// DebugColor applies the color associated with the debug level to
// the given string and returns the resulting string. If [UseColor] is set
// to false, it just returns the string it was passed.
func DebugColor(str string) string {
	return ApplyColor(matcolor.TheScheme.Secondary, str)
}

// InfoColor applies the color associated with the info level to
// the given string and returns the resulting string. If [UseColor] is set
// to false, it just returns the string it was passed.
func InfoColor(str string) string {
	return ApplyColor(matcolor.TheScheme.Primary, str)
}

// WarnColor applies the color associated with the warn level to
// the given string and returns the resulting string. If [UseColor] is set
// to false, it just returns the string it was passed.
func WarnColor(str string) string {
	return ApplyColor(matcolor.TheScheme.Tertiary, str)
}

// ErrorColor applies the color associated with the error level to
// the given string and returns the resulting string. If [UseColor] is set
// to false, it just returns the string it was passed.
func ErrorColor(str string) string {
	return ApplyColor(matcolor.TheScheme.Error, str)
}
