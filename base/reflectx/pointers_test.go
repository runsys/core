// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflectx

import (
	"reflect"
	"testing"
	"unsafe"
)

type PointerTestSub struct {
	Mbr1 string
	Mbr2 int
}

type PointerTest struct {
	Mbr1     string
	Mbr2     int
	SubField PointerTestSub
}

var pt = PointerTest{}

func InitPointerTest() {
	pt.Mbr1 = "mbr1 string"
	pt.Mbr2 = 2
}

func FieldValue(obj any, fld reflect.StructField) reflect.Value {
	ov := reflect.ValueOf(obj)
	f := unsafe.Pointer(ov.Pointer() + fld.Offset)
	nw := reflect.NewAt(fld.Type, f)
	return nw
}

func SubFieldValue(obj any, fld reflect.StructField, sub reflect.StructField) reflect.Value {
	ov := reflect.ValueOf(obj)
	f := unsafe.Pointer(ov.Pointer() + fld.Offset + sub.Offset)
	nw := reflect.NewAt(sub.Type, f)
	return nw
}

// test ability to create an addressable pointer value to fields of a struct
func TestNewAt(t *testing.T) {
	InitPointerTest()
	typ := reflect.TypeOf(pt)
	fld, _ := typ.FieldByName("Mbr2")
	vf := FieldValue(&pt, fld)

	// fmt.Printf("Fld: %v Typ: %v vf: %v vfi: %v vfT: %v vfp: %v canaddr: %v canset: %v caninterface: %v\n", fld.Name, vf.Type().String(), vf.String(), vf.Interface(), vf.Interface(), vf.Interface(), vf.CanAddr(), vf.CanSet(), vf.CanInterface())

	vf.Elem().Set(reflect.ValueOf(int(10)))

	if pt.Mbr2 != 10 {
		t.Errorf("Mbr2 should be 10, is: %v\n", pt.Mbr2)
	}

	fld, _ = typ.FieldByName("Mbr1")
	vf = FieldValue(&pt, fld)

	// fmt.Printf("Fld: %v Typ: %v vf: %v vfi: %v vfT: %v vfp: %v canaddr: %v canset: %v caninterface: %v\n", fld.Name, vf.Type().String(), vf.String(), vf.Interface(), vf.Interface(), vf.Interface(), vf.CanAddr(), vf.CanSet(), vf.CanInterface())

	vf.Elem().Set(reflect.ValueOf("this is a new string"))

	if pt.Mbr1 != "this is a new string" {
		t.Errorf("Mbr1 should be 'this is a new string': %v\n", pt.Mbr1)
	}
}
