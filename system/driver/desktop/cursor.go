// Copyright 2023 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package desktop

import (
	"sync"

	"cogentcore.org/core/cursors/cursorimg"
	"cogentcore.org/core/enums"
	"cogentcore.org/core/system"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// TheCursor is the single [system.Cursor] for the desktop platform
var TheCursor = &Cursor{CursorBase: system.CursorBase{Vis: true, Size: 32}, Cursors: map[enums.Enum]map[int]*glfw.Cursor{}}

// Cursor is the [cursor.Cursor] implementation for the desktop platform
type Cursor struct {
	system.CursorBase

	// Cursors are the cached glfw cursors
	Cursors map[enums.Enum]map[int]*glfw.Cursor

	// Mu is a mutex protecting access to the cursors
	Mu sync.Mutex

	// PrevSize is the cached previous size
	PrevSize int
}

func (cu *Cursor) Set(cursor enums.Enum) error {
	cu.Mu.Lock()
	defer cu.Mu.Unlock()
	if cursor == cu.Cur && cu.Size == cu.PrevSize { // we already have, so we don't need to set again
		return nil
	}
	sm := cu.Cursors[cursor]
	if sm == nil {
		sm = map[int]*glfw.Cursor{}
		cu.Cursors[cursor] = sm
	}
	if cur, ok := sm[cu.Size]; ok {
		TheApp.CtxWindow.Glw.SetCursor(cur)
		cu.PrevSize = cu.Size
		cu.Cur = cursor
		return nil
	}

	ci, err := cursorimg.Get(cursor, cu.Size)
	if err != nil {
		return err
	}
	h := ci.Hotspot
	gc := glfw.CreateCursor(ci.Image, h.X, h.Y)
	sm[cu.Size] = gc
	TheApp.CtxWindow.Glw.SetCursor(gc)
	cu.PrevSize = cu.Size
	cu.Cur = cursor
	return nil
}
