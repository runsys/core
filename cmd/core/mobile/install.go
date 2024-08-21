// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mobile

import (
	"fmt"
	"path/filepath"

	"cogentcore.org/core/base/exec"
	"cogentcore.org/core/cmd/core/config"
)

// Install installs the app named by the import path on the attached mobile device.
// It assumes that it has already been built.
//
// On Android, the 'adb' tool must be on the PATH.
func Install(c *config.Config) error {
	if len(c.Build.Target) != 1 {
		return fmt.Errorf("expected 1 target platform, but got %d (%v)", len(c.Build.Target), c.Build.Target)
	}
	t := c.Build.Target[0]
	switch t.OS {
	case "android":
		return exec.Run("adb", "install", "-r", filepath.Join(c.Build.Output, c.Name+".apk"))
	case "ios":
		return exec.Major().SetBuffer(false).Run("ios-deploy", "-b", filepath.Join(c.Build.Output, c.Name+".app"))
	default:
		return fmt.Errorf("mobile.Install only supports target platforms android and ios, but got %q", t.OS)
	}
}
