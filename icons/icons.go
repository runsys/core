// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package icons provides Go constant names for Material Design Symbols as SVG files.
package icons

import (
	"embed"
	"io/fs"
	"strings"

	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/base/mergefs"
)

//go:generate go run gen.go

// Icons contains all of the embedded svg icons. It is initialized
// to contain of the default icons located in the svg directory
// (https://github.com/cogentcore/core/tree/main/icons/svg), but it can
// be extended using [AddFS]. All icons should be stored in the root directory
// of the fs, which can be accomplished using [fsx.Sub] if you have icons in a
// subdirectory.
var Icons = fsx.Sub(defaults, "svg")

// defaults contains the default icons.
//
//go:embed svg/*.svg
var defaults embed.FS

const (
	// None is an icon that indicates to not use an icon.
	// It completely prevents the rendering of an icon,
	// whereas [Blank] renders a blank icon.
	None Icon = "none"

	// Blank is a blank icon that can be used as a
	// placeholder when no other icon is appropriate.
	// It still renders an icon, just a blank one,
	// whereas [None] indicates to not render one at all.
	Blank Icon = "blank"
)

// AddFS adds the given [fs.FS] of icons to the global [Icons] library.
// All icons should be stored in the root directory of the fs, which can be
// accomplished using [fsx.Sub] if you have icons in a subdirectory.
func AddFS(fs fs.FS) {
	Icons = mergefs.Merge(Icons, fs)
}

// Icon contains the name of an icon
type Icon string

// Fill returns the icon as a filled icon.
// It returns the icon unchanged if it is already filled.
func (i Icon) Fill() Icon {
	if i.IsFilled() {
		return i
	}
	return i + "-fill"
}

// Unfill returns the icon as an unfilled icon.
// It returns the icon unchanged if it is already unfilled.
// Icons are unfilled by default, so you only
// need to call this to reverse a prior [Icon.Fill] call
func (i Icon) Unfill() Icon {
	return Icon(strings.TrimSuffix(string(i), "-fill"))
}

// IsFilled returns whether the icon
// is a filled icon.
func (i Icon) IsFilled() bool {
	return strings.HasSuffix(string(i), "-fill")
}

// IsNil returns whether the icon name is "" or
// [None], which indicates not to use an icon.
func (i Icon) IsNil() bool {
	return i == "" || i == None
}

// IsSet returns whether the icon is set to a value other than "" or [None].
// It is the opposite of [Icon.IsNil].
func (i Icon) IsSet() bool {
	return !i.IsNil()
}

// IsValid returns whether the icon name corresponds to
// a valid existing icon.
func (i Icon) IsValid() bool {
	if i.IsNil() {
		return false
	}
	ex, _ := fsx.FileExistsFS(Icons, i.Filename())
	return ex
}

// Filename returns the filename of the icon in [Icons]
func (i Icon) Filename() string {
	return string(i) + ".svg"
}

// AllIcons is a list of all icons
var AllIcons []Icon

// All returns a list of all the Icons (excluding "fill" versions)
func All() []Icon {
	if AllIcons != nil {
		return AllIcons
	}
	files, err := fs.ReadDir(Icons, ".")
	if err != nil {
		return nil
	}
	ics := make([]Icon, 0, len(files)/2) // no fill
	for _, fi := range files {
		nm := fi.Name()
		if strings.HasSuffix(nm, "-fill.svg") {
			continue
		}
		ic := Icon(strings.TrimSuffix(fi.Name(), ".svg"))
		ics = append(ics, ic)
	}
	AllIcons = ics
	return AllIcons
}

// DefaultAppIcon is the default icon used for apps during packaging and in the app
// if no icon is specified in the icon.svg file. It defaults to a Google Blue version
// of [Toolbar].
const DefaultAppIcon = `<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 -960 960 960"><path fill="#4285f4" d="M180-120q-24 0-42-18t-18-42v-600q0-24 18-42t42-18h600q24 0 42 18t18 42v600q0 24-18 42t-42 18H180Zm0-513h600v-147H180v147Zm600 60H180v393h600v-393Zm-600-60v60-60Zm0 0v-147 147Zm0 60v393-393Z"/></svg>`
