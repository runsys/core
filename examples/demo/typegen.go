// Code generated by "core generate"; DO NOT EDIT.

package main

import (
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "main.tableStruct", IDName: "table-struct", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "Icon", Doc: "an icon"}, {Name: "IntField", Doc: "an integer field"}, {Name: "FloatField", Doc: "a float field"}, {Name: "StrField", Doc: "a string field"}, {Name: "File", Doc: "a file"}}})

var _ = types.AddType(&types.Type{Name: "main.inlineStruct", IDName: "inline-struct", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "On", Doc: "click to show next"}, {Name: "ShowMe", Doc: "can u see me?"}, {Name: "Cond", Doc: "a conditional"}, {Name: "Cond1", Doc: "On and Cond=0"}, {Name: "Cond2", Doc: "if Cond=0"}, {Name: "Value", Doc: "a value"}}})

var _ = types.AddType(&types.Type{Name: "main.testStruct", IDName: "test-struct", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "Enum", Doc: "An enum value"}, {Name: "Name", Doc: "a string"}, {Name: "ShowNext", Doc: "click to show next"}, {Name: "ShowMe", Doc: "can u see me?"}, {Name: "Inline", Doc: "how about that"}, {Name: "Cond", Doc: "a conditional"}, {Name: "Cond1", Doc: "if Cond=0"}, {Name: "Cond2", Doc: "if Cond>=0"}, {Name: "Value", Doc: "a value"}, {Name: "Vec"}, {Name: "Things"}, {Name: "Stuff"}, {Name: "File", Doc: "a file"}}})

var _ = types.AddFunc(&types.Func{Name: "main.hello", Doc: "Hello displays a greeting message and an age in weeks based on the given information.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"firstName", "lastName", "age", "likesGo"}, Returns: []string{"greeting", "weeksOld"}})
