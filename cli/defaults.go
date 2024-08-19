// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import "cogentcore.org/core/base/reflectx"

// SetFromDefaults sets the values of the given config object
// from `default:` field tag values. Parsing errors are automatically logged.
func SetFromDefaults(cfg any) error {
	return reflectx.SetFromDefaultTags(cfg)
}
