// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package styles

import "strings"

// ToCSS converts the given [Style] object to a semicolon-separated CSS string.
// It is not guaranteed to be fully complete or accurate.
func ToCSS(s *Style) string {
	parts := []string{}
	add := func(key, value string) {
		parts = append(parts, key+":"+value)
	}

	add("font-size", s.Font.Size.StringCSS())

	return strings.Join(parts, ";")
}
