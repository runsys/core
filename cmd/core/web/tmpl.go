// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"text/template"

	"cogentcore.org/core/cmd/core/config"
)

// AppJSTmpl is the template used in [MakeAppJS] to build the app.js file
var AppJSTmpl = template.Must(template.New("app.js").Parse(AppJS))

// AppJSData is the data passed to AppJSTmpl
type AppJSData struct {
	Env                     string
	WasmContentLengthHeader string
	AutoUpdateInterval      int64
}

// MakeAppJS exectues [AppJSTmpl] based on the given configuration information.
func MakeAppJS(c *config.Config) ([]byte, error) {
	if c.Web.Env == nil {
		c.Web.Env = make(map[string]string)
	}
	c.Web.Env["GOAPP_STATIC_RESOURCES_URL"] = "/"
	c.Web.Env["GOAPP_ROOT_PREFIX"] = "."

	for k, v := range c.Web.Env {
		if err := os.Setenv(k, v); err != nil {
			slog.Error("setting app env variable failed", "name", k, "value", "err", err)
		}
	}

	wenv, err := json.Marshal(c.Web.Env)
	if err != nil {
		return nil, err
	}

	d := AppJSData{
		Env:                     string(wenv),
		WasmContentLengthHeader: c.Web.WasmContentLengthHeader,
		AutoUpdateInterval:      c.Web.AutoUpdateInterval.Milliseconds(),
	}
	b := &bytes.Buffer{}
	err = AppJSTmpl.Execute(b, d)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// AppWorkerJSData is the data passed to [config.Config.Web.ServiceWorkerTemplate]
type AppWorkerJSData struct {
	Version          string
	ResourcesToCache string
}

// MakeWorkerJS executes [config.Config.Web.ServiceWorkerTemplate]. If it empty, it
// sets it to [DefaultAppWorkerJS].
func MakeAppWorkerJS(c *config.Config) ([]byte, error) {
	resources := []string{
		"app.css",
		"app.js",
		"app.wasm",
		"manifest.webmanifest",
		"wasm_exec.js",
		"index.html",
	}

	tmpl, err := template.New("app-worker.js").Parse(DefaultAppWorkerJS)
	if err != nil {
		return nil, err
	}

	rstr, err := json.Marshal(resources)
	if err != nil {
		return nil, err
	}

	d := AppWorkerJSData{
		Version:          c.Version,
		ResourcesToCache: string(rstr),
	}

	b := &bytes.Buffer{}
	err = tmpl.Execute(b, d)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// ManifestJSONTmpl is the template used in [MakeManifestJSON] to build the mainfest.webmanifest file
var ManifestJSONTmpl = template.Must(template.New("manifest.webmanifest").Parse(ManifestJSON))

// ManifestJSONData is the data passed to [ManifestJSONTmpl]
type ManifestJSONData struct {
	ShortName   string
	Name        string
	Description string
}

// MakeManifestJSON exectues [ManifestJSONTmpl] based on the given configuration information.
func MakeManifestJSON(c *config.Config) ([]byte, error) {
	d := ManifestJSONData{
		ShortName:   c.Name,
		Name:        c.Name,
		Description: c.About,
	}

	b := &bytes.Buffer{}
	err := ManifestJSONTmpl.Execute(b, d)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// IndexHTMLTmpl is the template used in [MakeIndexHTML] to build the index.html file
var IndexHTMLTmpl = template.Must(template.New("index.html").Parse(IndexHTML))

// IndexHTMLData is the data passed to [IndexHTMLTmpl]
type IndexHTMLData struct {
	BasePath               string
	Author                 string
	Desc                   string
	Keywords               []string
	Title                  string
	SiteName               string
	Image                  string
	VanityURL              string
	GithubVanityRepository string
}

// MakeIndexHTML exectues [IndexHTMLTmpl] based on the given configuration information,
// base path for app resources (used in [MakePages]), and optional title (used in [MakePages],
// defaults to [config.Config.Name] otherwise).
func MakeIndexHTML(c *config.Config, basePath string, title string) ([]byte, error) {
	if title == "" {
		title = c.Name
	}
	d := IndexHTMLData{
		BasePath:               basePath,
		Author:                 c.Web.Author,
		Desc:                   c.About,
		Keywords:               c.Web.Keywords,
		Title:                  title,
		SiteName:               c.Name,
		Image:                  c.Web.Image,
		VanityURL:              c.Web.VanityURL,
		GithubVanityRepository: c.Web.GithubVanityRepository,
	}

	b := &bytes.Buffer{}
	err := IndexHTMLTmpl.Execute(b, d)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
