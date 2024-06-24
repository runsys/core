// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflectx

import (
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/iox/jsonx"
)

// WalkFields calls the given walk function on all the exported primary fields of the
// given parent struct value, including those on anonymous embedded
// structs that this struct has. It effectively flattens all of the embedded fields
// of the struct.
//
// It passes the current parent struct, current [reflect.StructField],
// and current field value to the given should and walk functions.
//
// The given should function is called on every struct field (including
// on embedded structs themselves) to determine whether that field and any fields
// it has embedded should be handled (a return value of true indicates to continue
// down and a value of false indicates to not).
func WalkFields(parent reflect.Value, should func(parent reflect.Value, field reflect.StructField, value reflect.Value) bool, walk func(parent reflect.Value, parentField *reflect.StructField, field reflect.StructField, value reflect.Value)) {
	walkFields(parent, nil, should, walk)
}

func walkFields(parent reflect.Value, parentField *reflect.StructField, should func(parent reflect.Value, field reflect.StructField, value reflect.Value) bool, walk func(parent reflect.Value, parentField *reflect.StructField, field reflect.StructField, value reflect.Value)) {
	typ := parent.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		value := parent.Field(i)
		if !should(parent, field, value) {
			continue
		}
		if field.Type.Kind() == reflect.Struct && field.Anonymous {
			walkFields(value, &field, should, walk)
		} else {
			walk(parent, parentField, field, value)
		}
	}
}

// NumAllFields returns the number of elemental fields in the given struct type
// using [WalkFields].
func NumAllFields(parent reflect.Value) int {
	n := 0
	WalkFields(parent,
		func(parent reflect.Value, field reflect.StructField, value reflect.Value) bool {
			return true
		}, func(parent reflect.Value, parentField *reflect.StructField, field reflect.StructField, value reflect.Value) {
			n++
		})
	return n
}

// ValueIsDefault returns whether the given value is equivalent to the
// given string representation used in a field default tag.
func ValueIsDefault(fv reflect.Value, def string) bool {
	kind := fv.Kind()
	if kind >= reflect.Int && kind <= reflect.Complex128 && strings.Contains(def, ":") {
		dtags := strings.Split(def, ":")
		lo, _ := strconv.ParseFloat(dtags[0], 64)
		hi, _ := strconv.ParseFloat(dtags[1], 64)
		vf, err := ToFloat(fv.Interface())
		if err != nil {
			slog.Error("reflectx.ValueIsDefault: error parsing struct field numerical range def tag", "def", def, "err", err)
			return true
		}
		return lo <= vf && vf <= hi
	}
	dtags := strings.Split(def, ",")
	if strings.ContainsAny(def, "{[") { // complex type, so don't split on commas
		dtags = []string{def}
	}
	for _, df := range dtags {
		df = FormatDefault(df)
		if df == "" {
			return fv.IsZero()
		}
		dv := reflect.New(fv.Type())
		err := SetRobust(dv.Interface(), df)
		if err != nil {
			slog.Error("reflectx.ValueIsDefault: error getting value from default struct tag", "defaultStructTag", df, "value", fv, "err", err)
			return false
		}
		if reflect.DeepEqual(fv.Interface(), dv.Elem().Interface()) {
			return true
		}
	}
	return false
}

// SetFromDefaultTags sets the values of fields in the given struct based on
// `default:` default value struct field tags.
func SetFromDefaultTags(v any) error {
	if AnyIsNil(v) {
		return nil
	}
	ov := reflect.ValueOf(v)
	if ov.Kind() == reflect.Pointer && ov.IsNil() {
		return nil
	}
	val := NonPointerValue(ov)
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		fv := val.Field(i)
		def := f.Tag.Get("default")
		if NonPointerType(f.Type).Kind() == reflect.Struct && def == "" {
			SetFromDefaultTags(PointerValue(fv).Interface())
			continue
		}
		err := SetFromDefaultTag(fv, def)
		if err != nil {
			return fmt.Errorf("reflectx.SetFromDefaultTags: error setting field %q in object of type %q from val %q: %w", f.Name, typ.Name(), def, err)
		}
	}
	return nil
}

// SetFromDefaultTag sets the given value from the given default tag.
func SetFromDefaultTag(v reflect.Value, def string) error {
	def = FormatDefault(def)
	if def == "" {
		return nil
	}
	return SetRobust(UnderlyingPointer(v).Interface(), def)
}

// TODO: this needs to return an ordmap of the fields

// NonDefaultFields returns a map representing all of the fields of the given
// struct (or pointer to a struct) that have values different than their default
// values as specified by the `default:` struct tag. The resulting map is then typically
// saved using something like JSON or TOML. If a value has no default value, it
// checks whether its value is non-zero. If a field has a `save:"-"` tag, it wil
// not be included in the resulting map.
func NonDefaultFields(v any) map[string]any {
	res := map[string]any{}

	rv := NonPointerValue(reflect.ValueOf(v))
	if !rv.IsValid() {
		return nil
	}
	rt := rv.Type()
	nf := rt.NumField()
	for i := 0; i < nf; i++ {
		fv := rv.Field(i)
		ft := rt.Field(i)
		if ft.Tag.Get("save") == "-" {
			continue
		}
		def := ft.Tag.Get("default")
		if NonPointerType(ft.Type).Kind() == reflect.Struct && def == "" {
			sfm := NonDefaultFields(fv.Interface())
			if len(sfm) > 0 {
				res[ft.Name] = sfm
			}
			continue
		}
		if !ValueIsDefault(fv, def) {
			res[ft.Name] = fv.Interface()
		}
	}
	return res
}

// FormatDefault converts the given `default:` struct tag string into a format suitable
// for being used as a value in [SetRobust]. If it returns "", the default value
// should not be used.
func FormatDefault(def string) string {
	if def == "" {
		return ""
	}
	if strings.ContainsAny(def, "{[") { // complex type, so don't split on commas and colons
		return strings.ReplaceAll(def, `'`, `"`) // allow single quote to work as double quote for JSON format
	}
	// we split on commas and colons so we get the first item of lists and ranges
	def = strings.Split(def, ",")[0]
	def = strings.Split(def, ":")[0]
	return def
}

// StructTags returns a map[string]string of the tag string from a [reflect.StructTag] value.
func StructTags(tags reflect.StructTag) map[string]string {
	if len(tags) == 0 {
		return nil
	}
	flds := strings.Fields(string(tags))
	smap := make(map[string]string, len(flds))
	for _, fld := range flds {
		cli := strings.Index(fld, ":")
		if cli < 0 || len(fld) < cli+3 {
			continue
		}
		vl := strings.TrimSuffix(fld[cli+2:], `"`)
		smap[fld[:cli]] = vl
	}
	return smap
}

// StringJSON returns an indented JSON string representation
// of the given value for printing/debugging.
func StringJSON(v any) string {
	return string(errors.Log1(jsonx.WriteBytesIndent(v)))
}
