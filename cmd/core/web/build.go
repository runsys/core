// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

//go:generate go run gen/scripts.go

import (
	"crypto/sha1"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cogentcore.org/core/base/exec"
	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/core/cmd/core/config"
	"cogentcore.org/core/cmd/core/rendericon"
	"cogentcore.org/core/pages/wpath"
	strip "github.com/grokify/html-strip-tags-go"
)

// Build builds an app for web using the given configuration information.
func Build(c *config.Config) error {
	output := filepath.Join("bin", "web", "app.wasm")
	opath := output
	if c.Web.Gzip {
		opath += ".orig"
	}
	err := exec.Major().SetEnv("GOOS", "js").SetEnv("GOARCH", "wasm").Run("go", "build", "-o", opath, "-ldflags", config.LinkerFlags(c))
	if err != nil {
		return err
	}
	if c.Web.Gzip {
		err = exec.RemoveAll(output + ".orig.gz")
		if err != nil {
			return err
		}
		err = exec.Run("gzip", output+".orig")
		if err != nil {
			return err
		}
		err = os.Rename(output+".orig.gz", output)
		if err != nil {
			return err
		}
	}
	return MakeFiles(c)
}

// MakeFiles makes the necessary static web files based on the given configuration information.
func MakeFiles(c *config.Config) error {
	odir := filepath.Join("bin", "web")

	if c.Web.RandomVersion {
		t := time.Now().UTC().String()
		c.Version = fmt.Sprintf(`%x`, sha1.Sum([]byte(t)))
	}

	// The about text may contain HTML, which we need to get rid of.
	// It is trusted, so we do not need a more advanced sanitizer.
	c.About = strip.StripTags(c.About)

	wej := []byte(WASMExecJS())
	err := os.WriteFile(filepath.Join(odir, "wasm_exec.js"), wej, 0666)
	if err != nil {
		return err
	}

	ajs, err := MakeAppJS(c)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(odir, "app.js"), ajs, 0666)
	if err != nil {
		return err
	}

	awjs, err := MakeAppWorkerJS(c)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(odir, "app-worker.js"), awjs, 0666)
	if err != nil {
		return err
	}

	man, err := MakeManifestJSON(c)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(odir, "manifest.webmanifest"), man, 0666)
	if err != nil {
		return err
	}

	acs := []byte(AppCSS)
	err = os.WriteFile(filepath.Join(odir, "app.css"), acs, 0666)
	if err != nil {
		return err
	}

	iht, err := MakeIndexHTML(c, "", "")
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(odir, "index.html"), iht, 0666)
	if err != nil {
		return err
	}

	if c.Pages != "" {
		err := MakePages(c)
		if err != nil {
			return err
		}
	}

	err = os.MkdirAll(filepath.Join(odir, "icons"), 0777)
	if err != nil {
		return err
	}
	sizes := []int{32, 192, 512}
	for _, size := range sizes {
		ic, err := rendericon.Render(size)
		if err != nil {
			return err
		}
		err = imagex.Save(ic, filepath.Join(odir, "icons", strconv.Itoa(size)+".png"))
		if err != nil {
			return err
		}
	}
	err = exec.Run("cp", "icon.svg", filepath.Join(odir, "icons", "svg.svg"))
	if err != nil {
		return err
	}

	return nil
}

// MakePages makes a directory structure of pages for
// the core pages located at [config.Config.Pages].
func MakePages(c *config.Config) error {
	return filepath.WalkDir(c.Pages, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		path = strings.ReplaceAll(path, `\`, "/")
		path = strings.TrimSuffix(path, "index.md")
		path = strings.TrimSuffix(path, ".md")
		path = strings.TrimPrefix(path, c.Pages)
		path = strings.TrimPrefix(path, "/")
		path = wpath.Format(path)
		if path == "" { // exclude root index
			return nil
		}
		opath := filepath.Join("bin", "web", path)
		err = os.MkdirAll(opath, 0777)
		if err != nil {
			return err
		}
		numNested := strings.Count(path, "/") + 1
		basePath := ""
		for range numNested {
			basePath += "../"
		}
		title := wpath.Label(path, c.Name)
		if title != c.Name {
			title += " • " + c.Name
		}
		b, err := MakeIndexHTML(c, basePath, title)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(opath, "index.html"), b, 0666)
	})
}
