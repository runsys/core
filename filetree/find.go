// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filetree

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"cogentcore.org/core/base/fileinfo"
	"cogentcore.org/core/core"
	"cogentcore.org/core/tree"
)

// findDirNode finds directory node by given path.
// Must be a relative path already rooted at tree, or absolute path within tree.
func (fn *Node) findDirNode(path string) (*Node, error) {
	rp := fn.RelativePathFrom(core.Filename(path))
	if rp == "" {
		return nil, fmt.Errorf("FindDirNode: path: %s is not relative to this node's path: %s", path, fn.Filepath)
	}
	if rp == "." {
		return fn, nil
	}
	dirs := filepath.SplitList(rp)
	dir := dirs[0]
	dni := fn.ChildByName(dir, 0)
	if dni == nil {
		return nil, fmt.Errorf("could not find child %q", dir)
	}
	dn := AsNode(dni)
	if len(dirs) == 1 {
		if dn.IsDir() {
			return dn, nil
		}
		return nil, fmt.Errorf("FindDirNode: item at path: %s is not a Directory", path)
	}
	return dn.findDirNode(filepath.Join(dirs[1:]...))
}

// FindFile finds first node representing given file (false if not found) --
// looks for full path names that have the given string as their suffix, so
// you can include as much of the path (including whole thing) as is relevant
// to disambiguate.  See FilesMatching for a list of files that match a given
// string.
func (fn *Node) FindFile(fnm string) (*Node, bool) {
	if fnm == "" {
		return nil, false
	}
	fneff := fnm
	if len(fneff) > 2 && fneff[:2] == ".." { // relative path -- get rid of it and just look for relative part
		dirs := strings.Split(fneff, string(filepath.Separator))
		for i, dr := range dirs {
			if dr != ".." {
				fneff = filepath.Join(dirs[i:]...)
				break
			}
		}
	}

	if efn, err := fn.FileRoot.externalNodeByPath(fnm); err == nil {
		return efn, true
	}

	if strings.HasPrefix(fneff, string(fn.Filepath)) { // full path
		ffn, err := fn.dirsTo(fneff)
		if err == nil {
			return ffn, true
		}
		return nil, false
	}

	var ffn *Node
	found := false
	fn.WidgetWalkDown(func(cw core.Widget, cwb *core.WidgetBase) bool {
		sfn := AsNode(cw)
		if sfn == nil {
			return tree.Continue
		}
		if strings.HasSuffix(string(sfn.Filepath), fneff) {
			ffn = sfn
			found = true
			return tree.Break
		}
		return tree.Continue
	})
	return ffn, found
}

// NodeNameCount is used to report counts of different string-based things
// in the file tree
type NodeNameCount struct {
	Name  string
	Count int
}

func NodeNameCountSort(ecs []NodeNameCount) {
	sort.Slice(ecs, func(i, j int) bool {
		return ecs[i].Count > ecs[j].Count
	})
}

// FileExtensionCounts returns a count of all the different file extensions, sorted
// from highest to lowest.
// If cat is != fileinfo.Unknown then it only uses files of that type
// (e.g., fileinfo.Code to find any code files)
func (fn *Node) FileExtensionCounts(cat fileinfo.Categories) []NodeNameCount {
	cmap := make(map[string]int)
	fn.WidgetWalkDown(func(cw core.Widget, cwb *core.WidgetBase) bool {
		sfn := AsNode(cw)
		if sfn == nil {
			return tree.Continue
		}
		if cat != fileinfo.UnknownCategory {
			if sfn.Info.Cat != cat {
				return tree.Continue
			}
		}
		ext := strings.ToLower(filepath.Ext(sfn.Name))
		if ec, has := cmap[ext]; has {
			cmap[ext] = ec + 1
		} else {
			cmap[ext] = 1
		}
		return tree.Continue
	})
	ecs := make([]NodeNameCount, len(cmap))
	idx := 0
	for key, val := range cmap {
		ecs[idx] = NodeNameCount{Name: key, Count: val}
		idx++
	}
	NodeNameCountSort(ecs)
	return ecs
}

// LatestFileMod returns the most recent mod time of files in the tree.
// If cat is != fileinfo.Unknown then it only uses files of that type
// (e.g., fileinfo.Code to find any code files)
func (fn *Node) LatestFileMod(cat fileinfo.Categories) time.Time {
	tmod := time.Time{}
	fn.WidgetWalkDown(func(cw core.Widget, cwb *core.WidgetBase) bool {
		sfn := AsNode(cw)
		if sfn == nil {
			return tree.Continue
		}
		if cat != fileinfo.UnknownCategory {
			if sfn.Info.Cat != cat {
				return tree.Continue
			}
		}
		ft := (time.Time)(sfn.Info.ModTime)
		if ft.After(tmod) {
			tmod = ft
		}
		return tree.Continue
	})
	return tmod
}
