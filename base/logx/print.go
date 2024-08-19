// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logx

import (
	"fmt"
	"log/slog"
)

// Print is equivalent to [fmt.Print], but with color based on the given level.
// Also, if [UserLevel] is above the given level, it does not print anything.
func Print(level slog.Level, a ...any) (n int, err error) {
	if UserLevel > level {
		return 0, nil
	}
	return fmt.Print(LevelColor(level, fmt.Sprint(a...)))
}

// PrintDebug is equivalent to [Print] with level [slog.LevelDebug].
func PrintDebug(a ...any) (n int, err error) {
	return Print(slog.LevelDebug, a...)
}

// PrintInfo is equivalent to [Print] with level [slog.LevelInfo].
func PrintInfo(a ...any) (n int, err error) {
	return Print(slog.LevelInfo, a...)
}

// PrintWarn is equivalent to [Print] with level [slog.LevelWarn].
func PrintWarn(a ...any) (n int, err error) {
	return Print(slog.LevelWarn, a...)
}

// PrintError is equivalent to [Print] with level [slog.LevelError].
func PrintError(a ...any) (n int, err error) {
	return Print(slog.LevelError, a...)
}

// Println is equivalent to [fmt.Println], but with color based on the given level.
// Also, if [UserLevel] is above the given level, it does not print anything.
func Println(level slog.Level, a ...any) (n int, err error) {
	if UserLevel > level {
		return 0, nil
	}
	return fmt.Println(LevelColor(level, fmt.Sprint(a...)))
}

// PrintlnDebug is equivalent to [Println] with level [slog.LevelDebug].
func PrintlnDebug(a ...any) (n int, err error) {
	return Println(slog.LevelDebug, a...)
}

// PrintlnInfo is equivalent to [Println] with level [slog.LevelInfo].
func PrintlnInfo(a ...any) (n int, err error) {
	return Println(slog.LevelInfo, a...)
}

// PrintlnWarn is equivalent to [Println] with level [slog.LevelWarn].
func PrintlnWarn(a ...any) (n int, err error) {
	return Println(slog.LevelWarn, a...)
}

// PrintlnError is equivalent to [Println] with level [slog.LevelError].
func PrintlnError(a ...any) (n int, err error) {
	return Println(slog.LevelError, a...)
}

// Printf is equivalent to [fmt.Printf], but with color based on the given level.
// Also, if [UserLevel] is above the given level, it does not print anything.
func Printf(level slog.Level, format string, a ...any) (n int, err error) {
	if UserLevel > level {
		return 0, nil
	}
	return fmt.Println(LevelColor(level, fmt.Sprintf(format, a...)))
}

// PrintfDebug is equivalent to [Printf] with level [slog.LevelDebug].
func PrintfDebug(format string, a ...any) (n int, err error) {
	return Printf(slog.LevelDebug, format, a...)
}

// PrintfInfo is equivalent to [Printf] with level [slog.LevelInfo].
func PrintfInfo(format string, a ...any) (n int, err error) {
	return Printf(slog.LevelInfo, format, a...)
}

// PrintfWarn is equivalent to [Printf] with level [slog.LevelWarn].
func PrintfWarn(format string, a ...any) (n int, err error) {
	return Printf(slog.LevelWarn, format, a...)
}

// PrintfError is equivalent to [Printf] with level [slog.LevelError].
func PrintfError(format string, a ...any) (n int, err error) {
	return Printf(slog.LevelError, format, a...)
}
