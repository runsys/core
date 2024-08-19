// Copyright (c) 2022, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ordmap

import (
	"testing"
)

func TestMap(t *testing.T) {
	om := New[string, int]()
	om.Add("key0", 0)
	om.Add("key1", 1)
	om.Add("key2", 2)

	if v, ok := om.ValueByKeyTry("key1"); !ok || v != 1 {
		t.Error("ValByKey")
	}

	if i, ok := om.IndexByKeyTry("key2"); !ok || i != 2 {
		t.Error("IndexByKey")
	}

	if om.KeyByIndex(0) != "key0" {
		t.Error("KeyByIndex")
	}

	if om.ValueByIndex(1) != 1 {
		t.Error("ValByIndex")
	}

	if om.Len() != 3 {
		t.Error("Len")
	}

	om.DeleteIndex(1, 2)
	// for i, kv := range om.Order {
	// 	fmt.Printf("i: %d\tkv: %v\n", i, kv)
	// }
	if om.ValueByIndex(1) != 2 {
		t.Error("DeleteIndex")
	}
	if i, ok := om.IndexByKeyTry("key2"); !ok || i != 1 {
		t.Error("Delete IndexByKey")
	}

	// fmt.Printf("%v\n", om.Map)
	om.InsertAtIndex(0, "new0", 3)
	// fmt.Printf("%v\n", om.Map)
	if om.ValueByIndex(0) != 3 {
		t.Error("InsertAtIndex ValByIndex 0")
	}
	if om.ValueByIndex(1) != 0 {
		t.Error("InsertAtIndex ValByIndex 1")
	}
	if i, ok := om.IndexByKeyTry("key2"); !ok || i != 2 {
		t.Errorf("InsertAtIndex IndexByKey: %d != 2", i)
	}

	// constr

	nm := Make([]KeyValue[string, int]{{"one", 1}, {"two", 2}, {"three", 3}})

	if nm.ValueByIndex(2) != 3 {
		t.Error("Make ValByIndex 2")
	}
	if nm.ValueByIndex(1) != 2 {
		t.Error("Make ValByIndex 1")
	}
	if val, ok := nm.ValueByKeyTry("three"); !ok || val != 3 {
		t.Error("Make ValByKey 3")
	}

}
