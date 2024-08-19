// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strcase

import "testing"

func TestFormatList(t *testing.T) {
	type test struct {
		items []string
		want  string
	}
	tests := []test{
		{nil, ""},
		{[]string{"Go"}, "Go"},
		{[]string{"Go", "Python"}, "Go and Python"},
		{[]string{"Go", "Python", "JavaScript"}, "Go, Python, and JavaScript"},
		{[]string{"Go", "Python", "JavaScript", "C"}, "Go, Python, JavaScript, and C"},
	}
	for _, test := range tests {
		have := FormatList(test.items...)
		if have != test.want {
			t.Errorf("expected %q but got %q for %v", test.want, have, test.items)
		}
	}
}
