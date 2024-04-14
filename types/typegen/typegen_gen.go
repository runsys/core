// Code generated by "typegen -output typegen_gen.go"; DO NOT EDIT.

package typegen

import (
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/types/typegen.Config", IDName: "config", Doc: "Config contains the configuration information\nused by typegen", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "Dir", Doc: "the source directory to run typegen on (can be set to multiple through paths like ./...)"}, {Name: "Output", Doc: "the output file location relative to the package on which typegen is being called"}, {Name: "AddTypes", Doc: "whether to add types to typegen by default"}, {Name: "AddMethods", Doc: "whether to add methods to typegen by default"}, {Name: "AddFuncs", Doc: "whether to add functions to typegen by default"}, {Name: "InterfaceConfigs", Doc: "An ordered map of configs keyed by fully-qualified interface type names; if a type implements the interface, the config will be applied to it.\nThe configs are applied in sequential ascending order, which means that\nthe last config overrides the other ones, so the most specific\ninterfaces should typically be put last.\nNote: the package typegen is run on must explicitly reference this interface at some point for this to work; adding a simple\n`var _ MyInterface = (*MyType)(nil)` statement to check for interface implementation is an easy way to accomplish that.\nNote: typegen will still succeed if it can not find one of the interfaces specified here in order to allow it to work generically across multiple directories; you can use the -v flag to get log warnings about this if you suspect that it is not finding interfaces when it should."}, {Name: "Instance", Doc: "whether to generate an instance of the type(s)"}, {Name: "TypeVar", Doc: "whether to generate a global type variable of the form 'TypeNameType'"}, {Name: "Setters", Doc: "Whether to generate chaining `Set*` methods for each exported field of each type (eg: \"SetText\" for field \"Text\").\nIf this is set to true, then you can add `set:\"-\"` struct tags to individual fields\nto prevent Set methods being generated for them."}, {Name: "Templates", Doc: "a slice of templates to execute on each type being added; the template data is of the type typegen.Type"}}})

var _ = types.AddFunc(&types.Func{Name: "cogentcore.org/core/types/typegen.Generate", Doc: "Generate generates typegen type info, using the\nconfiguration information, loading the packages from the\nconfiguration source directory, and writing the result\nto the configuration output file.\n\nIt is a simple entry point to typegen that does all\nof the steps; for more specific functionality, create\na new [Generator] with [NewGenerator] and call methods on it.", Directives: []types.Directive{{Tool: "cli", Directive: "cmd", Args: []string{"-root"}}, {Tool: "types", Directive: "add"}}, Args: []string{"cfg"}, Returns: []string{"error"}})
