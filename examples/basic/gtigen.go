// Code generated by "gtigen.exe"; DO NOT EDIT.

package main

import (
	"goki.dev/gti"
	"goki.dev/ordmap"
)

var _ = gti.AddFunc(&gti.Func{
	Name: "main.Build",
	Doc:  "Build builds the app for the given platform.",
	Directives: gti.Directives{
		&gti.Directive{Tool: "gti", Directive: "add", Args: []string{}},
	},
	Args: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"c", &gti.Field{Name: "c", Doc: "", Directives: gti.Directives{}}},
	}),
	Returns: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"error", &gti.Field{Name: "error", Doc: "", Directives: gti.Directives{}}},
	}),
})

var _ = gti.AddFunc(&gti.Func{
	Name: "main.Run",
	Doc:  "Run runs the app for the given user.",
	Directives: gti.Directives{
		&gti.Directive{Tool: "gti", Directive: "add", Args: []string{}},
		&gti.Directive{Tool: "grease", Directive: "cmd", Args: []string{"-root"}},
	},
	Args: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"c", &gti.Field{Name: "c", Doc: "", Directives: gti.Directives{}}},
	}),
	Returns: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"error", &gti.Field{Name: "error", Doc: "", Directives: gti.Directives{}}},
	}),
})

var _ = gti.AddFunc(&gti.Func{
	Name: "main.ModTidy",
	Doc:  "",
	Directives: gti.Directives{
		&gti.Directive{Tool: "gti", Directive: "add", Args: []string{}},
		&gti.Directive{Tool: "grease", Directive: "cmd", Args: []string{"-name", "mod tidy"}},
	},
	Args: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"c", &gti.Field{Name: "c", Doc: "", Directives: gti.Directives{}}},
	}),
	Returns: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"error", &gti.Field{Name: "error", Doc: "", Directives: gti.Directives{}}},
	}),
})
