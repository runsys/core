// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"log/slog"
	"os"
	"reflect"
	"strings"

	"slices"

	"cogentcore.org/core/base/ordmap"
	"cogentcore.org/core/base/reflectx"
	"cogentcore.org/core/base/strcase"
)

// field represents a struct field in a configuration struct.
type field struct {
	// Field is the reflect struct field object for this field
	Field reflect.StructField
	// Value is the reflect value of the settable pointer to this field
	Value reflect.Value
	// Struct is the parent struct that contains this field
	Struct reflect.Value
	// Name is the fully qualified, nested name of this field (eg: A.B.C).
	// It is as it appears in code, and is NOT transformed something like kebab-case.
	Name string
	// Names contains all of the possible end-user names for this field as a flag.
	// It defaults to the name of the field, but custom names can be specified via
	// the cli struct tag.
	Names []string
}

// fields is a simple type alias for an ordered map of [field] objects.
type fields = ordmap.Map[string, *field]

// addAllFields, when passed as the command to [addFields], indicates
// to add all fields, regardless of their command association.
const addAllFields = "*"

// addFields adds to the given fields map all of the fields of the given
// object, in the context of the given command name. A value of [addAllFields]
// for cmd indicates to add all fields, regardless of their command association.
func addFields(obj any, allFields *fields, cmd string) {
	addFieldsImpl(obj, "", "", allFields, map[string]*field{}, cmd)
}

// addFieldsImpl is the underlying implementation of [addFields].
// The path is the current path state, the cmdPath is the
// current path state without command-associated names,
// and usedNames is a map keyed by used CamelCase names with values
// of their associated fields, used to track naming conflicts. The
// [field.Name]s of the fields are set based on the path, whereas the
// names of the flags are set based on the command path. The difference
// between the two is that the path is always fully qualified, whereas the
// command path omits the names of structs associated with commands via
// the "cmd" struct tag, as the user already knows what command they are
// running, so they do not need that duplicated specificity for every flag.
func addFieldsImpl(obj any, path string, cmdPath string, allFields *fields, usedNames map[string]*field, cmd string) {
	if reflectx.AnyIsNil(obj) {
		return
	}
	ov := reflect.ValueOf(obj)
	if ov.Kind() == reflect.Pointer && ov.IsNil() {
		return
	}
	val := reflectx.NonPointerValue(ov)
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		fv := val.Field(i)
		pval := reflectx.PointerValue(fv)
		cmdTag, hct := f.Tag.Lookup("cmd")
		cmds := strings.Split(cmdTag, ",")
		if hct && !slices.Contains(cmds, cmd) && !slices.Contains(cmds, addAllFields) { // if we are associated with a different command, skip
			continue
		}
		if reflectx.NonPointerType(f.Type).Kind() == reflect.Struct {
			nwPath := f.Name
			if path != "" {
				nwPath = path + "." + nwPath
			}
			nwCmdPath := f.Name
			// if we have a command tag, we don't scope our command path with our name,
			// as we don't need it because we ran that command and know we are in it
			if hct {
				nwCmdPath = ""
			}
			if cmdPath != "" {
				nwCmdPath = cmdPath + "." + nwCmdPath
			}

			addFieldsImpl(reflectx.PointerValue(fv).Interface(), nwPath, nwCmdPath, allFields, usedNames, cmd)
			// we still add ourself if we are a struct, so we keep going,
			// unless we are associated with a command, in which case there
			// is no point in adding ourself
			if hct {
				continue
			}
		}
		// we first add our unqualified command name, which is the best case scenario
		name := f.Name
		names := []string{name}
		// then, we set our future [Field.Name] to the fully path scoped version (but we don't add it as a command name)
		if path != "" {
			name = path + "." + name
		}
		// then, we set add our command path scoped name as a command name
		if cmdPath != "" {
			names = append(names, cmdPath+"."+f.Name)
		}

		flagTag, ok := f.Tag.Lookup("flag")
		if ok {
			names = strings.Split(flagTag, ",")
			if len(names) == 0 {
				slog.Error("programmer error: expected at least one name in flag struct tag, but got none")
			}
		}

		nf := &field{
			Field:  f,
			Value:  pval,
			Struct: ov,
			Name:   name,
			Names:  names,
		}
		for i, name := range nf.Names {
			// duplicate deletion can cause us to get out of range
			if i >= len(nf.Names) {
				break
			}
			name := strcase.ToCamel(name)        // everybody is in camel for naming conflict check
			if of, has := usedNames[name]; has { // we have a conflict
				// if we have a naming conflict between two fields with the same base
				// (in the same parent struct), then there is no nesting and they have
				// been directly given conflicting names, so there is a simple programmer error
				nbase := ""
				nli := strings.LastIndex(nf.Name, ".")
				if nli >= 0 {
					nbase = nf.Name[:nli]
				}
				obase := ""
				oli := strings.LastIndex(of.Name, ".")
				if oli >= 0 {
					obase = of.Name[:oli]
				}
				if nbase == obase {
					slog.Error("programmer error: cli: two fields were assigned the same name", "name", name, "field0", of.Name, "field1", nf.Name)
					os.Exit(1)
				}

				// if that isn't the case, they are in different parent structs and
				// it is a nesting problem, so we use the nest tags to resolve the conflict.
				// the basic rule is that whoever specifies the nest:"-" tag gets to
				// be non-nested, and if no one specifies it, everyone is nested.
				// if both want to be non-nested, that is a programmer error.

				// nest field tag values for new and other
				nfns := nf.Field.Tag.Get("nest")
				ofns := of.Field.Tag.Get("nest")

				// whether new and other get to have non-nested version
				nfn := nfns == "-" || nfns == "false"
				ofn := ofns == "-" || ofns == "false"

				if nfn && ofn {
					slog.Error(`programmer error: cli: nest:"-" specified on two config fields with the same name; keep nest:"-" on the field you want to be able to access without nesting and remove it from the other one`, "name", name, "field0", of.Name, "field1", nf.Name, "exampleFlagWithoutNesting", "-"+name, "exampleFlagWithNesting", "-"+strcase.ToKebab(nf.Name))
					os.Exit(1)
				} else if !nfn && !ofn {
					// neither one gets it, so we replace both with fully qualified name
					applyShortestUniqueName(nf, i, usedNames)
					for i, on := range of.Names {
						if on == name {
							applyShortestUniqueName(of, i, usedNames)
						}
					}
				} else if nfn && !ofn {
					// we get it, so we keep ours as is and replace them with fully qualified name
					for i, on := range of.Names {
						if on == name {
							applyShortestUniqueName(of, i, usedNames)
						}
					}
					// we also need to update the field for our name to us
					usedNames[name] = nf
				} else if !nfn && ofn {
					// they get it, so we replace ours with fully qualified name
					applyShortestUniqueName(nf, i, usedNames)
				}
			} else {
				// if no conflict, we get the name
				usedNames[name] = nf
			}
		}
		allFields.Add(name, nf)
	}
}

// applyShortestUniqueName uses [shortestUniqueName] to apply the shortest
// unique name for the given field, in the context of the given
// used names, at the given index.
func applyShortestUniqueName(field *field, idx int, usedNames map[string]*field) {
	nm := shortestUniqueName(field.Name, usedNames)
	// if we already have this name, we don't need to add it, so we just delete this entry
	if slices.Contains(field.Names, nm) {
		field.Names = slices.Delete(field.Names, idx, idx+1)
	} else {
		field.Names[idx] = nm
		usedNames[nm] = field
	}
}

// shortestUniqueName returns the shortest unique camel-case name for
// the given fully qualified nest name of a field, using the given
// map of used names. It works backwards, so, for example, if given "A.B.C.D",
// it would check "D", then "C.D", then "B.C.D", and finally "A.B.C.D".
func shortestUniqueName(name string, usedNames map[string]*field) string {
	strs := strings.Split(name, ".")
	cur := ""
	for i := len(strs) - 1; i >= 0; i-- {
		if cur == "" {
			cur = strs[i]
		} else {
			cur = strs[i] + "." + cur
		}
		if _, has := usedNames[cur]; !has {
			return cur
		}
	}
	return cur // TODO: this should never happen, but if it does, we might want to print an error
}
