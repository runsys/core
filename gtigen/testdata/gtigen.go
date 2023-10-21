// Code generated by "gtigen.test -test.paniconexit0 -test.timeout=10m0s"; DO NOT EDIT.

package testdata

import (
	"goki.dev/gti"
	"goki.dev/ordmap"
)

// PersonType is the [gti.Type] for [Person]
var PersonType = gti.AddType(&gti.Type{
	Name:      "goki.dev/gti/gtigen/testdata.Person",
	ShortName: "testdata.Person",
	IDName:    "person",
	Doc:       "Person represents a person and their attributes.\nThe zero value of a Person is not valid.",
	Directives: gti.Directives{
		&gti.Directive{Tool: "ki", Directive: "flagtype", Args: []string{"NodeFlags", "-field", "Flag"}},
	},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Name", &gti.Field{Name: "Name", Type: "string", LocalType: "string", Doc: "Name is the name of the person", Directives: gti.Directives{
			&gti.Directive{Tool: "gi", Directive: "toolbar", Args: []string{"-hide"}},
		}, Tag: ""}},
		{"Age", &gti.Field{Name: "Age", Type: "int", LocalType: "int", Doc: "Age is the age of the person", Directives: gti.Directives{
			&gti.Directive{Tool: "gi", Directive: "view", Args: []string{"inline"}},
		}, Tag: "json:\"-\""}},
		{"Type", &gti.Field{Name: "Type", Type: "*goki.dev/gti.Type", LocalType: "*gti.Type", Doc: "Type is the type of the person", Directives: gti.Directives{}, Tag: ""}},
	}),
	Embeds: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{
		{"Introduction", &gti.Method{Name: "Introduction", Doc: "Introduction returns an introduction for the person.\nIt contains the name of the person and their age.", Directives: gti.Directives{
			&gti.Directive{Tool: "gi", Directive: "toolbar", Args: []string{"-name", "ShowIntroduction", "-icon", "play", "-show-result", "-confirm"}},
		}, Args: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}), Returns: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
			{"string", &gti.Field{Name: "string", Type: "string", LocalType: "string", Doc: "", Directives: gti.Directives{}, Tag: ""}},
		})}},
	}),
	Instance: &Person{},
})

func (t *Person) MyCustomFuncForStringers(a any) error {
	return nil
}

var _ = gti.AddFunc(&gti.Func{
	Name:       "goki.dev/gti/gtigen/testdata.Alert",
	Doc:        "Alert prints an alert with the given message",
	Directives: gti.Directives{},
	Args: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"msg", &gti.Field{Name: "msg", Type: "string", LocalType: "string", Doc: "", Directives: gti.Directives{}, Tag: ""}},
	}),
	Returns: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
})
