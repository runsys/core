// Code generated by "typegen.test -test.testlogfile=/var/folders/x1/r8shprmj7j71zbw3qvgl9dqc0000gq/T/go-build4151448840/b905/testlog.txt -test.paniconexit0 -test.timeout=10m0s -test.v=true"; DO NOT EDIT.

package testdata

import (
	"cogentcore.org/core/types"
)

// PersonType is the [types.Type] for [Person]
var PersonType = types.AddType(&types.Type{Name: "cogentcore.org/core/types/typegen/testdata.Person", IDName: "person", Doc: "Person represents a person and their attributes.\nThe zero value of a Person is not valid.", Directives: []types.Directive{{Tool: "ki", Directive: "flagtype", Args: []string{"NodeFlags", "-field", "Flag"}}, {Tool: "core", Directive: "embedder"}}, Methods: []types.Method{{Name: "Introduction", Doc: "Introduction returns an introduction for the person.\nIt contains the name of the person and their age.", Directives: []types.Directive{{Tool: "gi", Directive: "toolbar", Args: []string{"-name", "ShowIntroduction", "-icon", "play", "-show-result", "-confirm"}}, {Tool: "types", Directive: "add"}}, Returns: []string{"string"}}}, Embeds: []types.Field{{Name: "RGBA"}}, Fields: []types.Field{{Name: "Name", Doc: "Name is the name of the person"}, {Name: "Age", Doc: "Age is the age of the person"}, {Name: "Type", Doc: "Type is the type of the person"}, {Name: "unexportedField"}, {Name: "Nicknames", Doc: "Nicknames are the nicknames of the person"}}, Instance: &Person{}})

func (t *Person) MyCustomFuncForStringers(a any) error {
	return nil
}

// SetName sets the [Person.Name]:
// Name is the name of the person
func (t *Person) SetName(v string) *Person { t.Name = v; return t }

// SetAge sets the [Person.Age]:
// Age is the age of the person
func (t *Person) SetAge(v int) *Person { t.Age = v; return t }

// SetType sets the [Person.Type]:
// Type is the type of the person
func (t *Person) SetType(v *types.Type) *Person { t.Type = v; return t }

// SetNicknames sets the [Person.Nicknames]:
// Nicknames are the nicknames of the person
func (t *Person) SetNicknames(v ...string) *Person { t.Nicknames = v; return t }

// SetR sets the [Person.R]
func (t *Person) SetR(v uint8) *Person { t.R = v; return t }

// SetG sets the [Person.G]
func (t *Person) SetG(v uint8) *Person { t.G = v; return t }

// SetB sets the [Person.B]
func (t *Person) SetB(v uint8) *Person { t.B = v; return t }

// SetA sets the [Person.A]
func (t *Person) SetA(v uint8) *Person { t.A = v; return t }

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/types/typegen/testdata.BlockType", IDName: "block-type", Doc: "BlockType is a type declared in a type block.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}})

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/types/typegen/testdata.CommaFieldType", IDName: "comma-field-type", Doc: "CommaFieldType is a type with inline comma fields.", Directives: []types.Directive{{Tool: "types", Directive: "add", Args: []string{"-setters"}}}, Fields: []types.Field{{Name: "A"}, {Name: "B"}}})

// SetA sets the [CommaFieldType.A]
func (t *CommaFieldType) SetA(v int) *CommaFieldType { t.A = v; return t }

// SetB sets the [CommaFieldType.B]
func (t *CommaFieldType) SetB(v int) *CommaFieldType { t.B = v; return t }

var _ = types.AddFunc(&types.Func{Name: "cogentcore.org/core/types/typegen/testdata.Alert", Doc: "Alert prints an alert with the given message", Args: []string{"msg"}})

var _ = types.AddFunc(&types.Func{Name: "cogentcore.org/core/types/typegen/testdata.TypeOmittedArgs0", Args: []string{"x", "y"}})

var _ = types.AddFunc(&types.Func{Name: "cogentcore.org/core/types/typegen/testdata.TypeOmittedArgs1", Args: []string{"x", "y", "z"}})

var _ = types.AddFunc(&types.Func{Name: "cogentcore.org/core/types/typegen/testdata.TypeOmittedArgs2", Args: []string{"x", "y", "z"}})

var _ = types.AddFunc(&types.Func{Name: "cogentcore.org/core/types/typegen/testdata.TypeOmittedArgs3", Args: []string{"x", "y", "z", "w"}})

var _ = types.AddFunc(&types.Func{Name: "cogentcore.org/core/types/typegen/testdata.TypeOmittedArgs4", Args: []string{"x", "y", "z", "w"}})
