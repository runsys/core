// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"cogentcore.org/core/base/exec"
	"cogentcore.org/core/cmd/core/config"
)

// Log prints the logs from your app running on Android to the terminal.
// Android is the only supported platform for log; use the -debug flag on
// run for other platforms.
func Log(c *config.Config) error { //types:add
	if c.Log.Target != "android" {
		return fmt.Errorf("only android is supported for log; use the -debug flag on run for other platforms")
	}
	if !c.Log.Keep {
		err := exec.Run("adb", "logcat", "-c")
		if err != nil {
			return fmt.Errorf("error clearing logs: %w", err)
		}
	}
	// we are logging continiously so we can't buffer, and we must be verbose
	err := exec.Verbose().SetBuffer(false).Run("adb", "logcat", "*:"+c.Log.All, "Go:D", "GoLog:D")
	if err != nil {
		return fmt.Errorf("erroring getting logs: %w", err)
	}
	return nil
}
