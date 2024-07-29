// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package styles

import (
	"fmt"
	"strconv"
	"strings"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
)

// ToCSS converts the given [Style] object to a semicolon-separated CSS string.
// It is not guaranteed to be fully complete or accurate.
func ToCSS(s *Style) string {
	parts := []string{}
	add := func(key, value string) {
		if value == "0" || value == "0px" || value == "0dot" {
			return
		}
		parts = append(parts, key+":"+value)
	}

	if s.Color != nil {
		add("color", colors.AsHex(colors.ToUniform(s.Color)))
	}
	if s.Background != nil {
		add("background", colors.AsHex(colors.ToUniform(s.Background)))
	}
	if s.Is(states.Invisible) {
		add("display", "none")
	} else {
		add("display", s.Display.String())
	}
	add("flex-direction", s.Direction.String())
	add("flex-grow", fmt.Sprintf("%g", s.Grow.Y))
	add("justify-content", s.Justify.Content.String())
	add("align-items", s.Align.Items.String())
	add("columns", strconv.Itoa(s.Columns))
	add("gap", s.Gap.X.StringCSS())
	add("width", s.Min.X.StringCSS())
	add("height", s.Min.Y.StringCSS())
	add("max-width", s.Max.X.StringCSS())
	add("max-height", s.Max.Y.StringCSS())
	add("padding", s.Padding.Top.StringCSS())
	add("margin", s.Margin.Top.StringCSS())
	if s.Font.Size.Value != 16 || s.Font.Size.Unit != units.UnitDp {
		add("font-size", s.Font.Size.StringCSS())
	}
	if s.Font.Family != "" && s.Font.Family != "Roboto" {
		ff := s.Font.Family
		if strings.HasSuffix(ff, "Mono") {
			ff += ", monospace"
		} else {
			ff += ", sans-serif"
		}
		add("font-family", ff)
	}
	if s.Font.Weight == WeightMedium {
		add("font-weight", "500")
	} else {
		add("font-weight", s.Font.Weight.String())
	}
	add("line-height", s.Text.LineHeight.StringCSS())
	if s.Border.Width.Top.Value > 0 {
		add("border-style", s.Border.Style.Top.String())
		add("border-width", s.Border.Width.Top.StringCSS())
		add("border-color", colors.AsHex(colors.ToUniform(s.Border.Color.Top)))
	}
	add("border-radius", s.Border.Radius.Top.StringCSS())

	return strings.Join(parts, ";")
}
