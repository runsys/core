// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filetree

import "sync"

// DirFlags are flags on directories: Open, SortBy, etc.
// These flags are stored in the DirFlagMap for persistence.
// This map is saved to a file, so these flags must be stored
// as bit flags instead of a struct to ensure efficient serialization.
type DirFlags int64 //enums:bitflag -trim-prefix Dir

const (
	// DirIsOpen means directory is open -- else closed
	DirIsOpen DirFlags = iota

	// DirSortByName means sort the directory entries by name.
	// this is mutex with other sorts -- keeping option open for non-binary sort choices.
	DirSortByName

	// DirSortByModTime means sort the directory entries by modification time
	DirSortByModTime
)

// DirFlagMap is a map for encoding directories that are open in the file
// tree.  The strings are typically relative paths.  The bool value is used to
// mark active paths and inactive (unmarked) ones can be removed.
// Map access is protected by Mutex.
type DirFlagMap struct {

	// map of paths and associated flags
	Map map[string]DirFlags

	// mutex for accessing map
	Mu sync.Mutex `display:"-" json:"-" xml:"-" toml:"-"`
}

// Init initializes the map, and sets the Mutex lock -- must unlock manually
func (dm *DirFlagMap) Init() {
	dm.Mu.Lock()
	if dm.Map == nil {
		dm.Map = make(map[string]DirFlags)
	}
}

// IsOpen returns true if path has IsOpen bit flag set
func (dm *DirFlagMap) IsOpen(path string) bool {
	dm.Init()
	defer dm.Mu.Unlock()
	if df, ok := dm.Map[path]; ok {
		return df.HasFlag(DirIsOpen)
	}
	return false
}

// SetOpenState sets the given directory's open flag
func (dm *DirFlagMap) SetOpen(path string, open bool) {
	dm.Init()
	defer dm.Mu.Unlock()
	df := dm.Map[path]
	df.SetFlag(open, DirIsOpen)
	dm.Map[path] = df
}

// SortByName returns true if path is sorted by name (default if not in map)
func (dm *DirFlagMap) SortByName(path string) bool {
	dm.Init()
	defer dm.Mu.Unlock()
	if df, ok := dm.Map[path]; ok {
		return df.HasFlag(DirSortByName)
	}
	return true
}

// SortByModTime returns true if path is sorted by mod time
func (dm *DirFlagMap) SortByModTime(path string) bool {
	dm.Init()
	defer dm.Mu.Unlock()
	if df, ok := dm.Map[path]; ok {
		return df.HasFlag(DirSortByModTime)
	}
	return false
}

// SetSortBy sets the given directory's sort by option
func (dm *DirFlagMap) SetSortBy(path string, modTime bool) {
	dm.Init()
	defer dm.Mu.Unlock()
	df := dm.Map[path]
	if modTime {
		df.SetFlag(true, DirSortByModTime)
		df.SetFlag(false, DirSortByName)
	} else {
		df.SetFlag(false, DirSortByModTime)
		df.SetFlag(true, DirSortByName)
	}
	dm.Map[path] = df
}
