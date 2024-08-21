// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filetree

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"cogentcore.org/core/base/fileinfo"
	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/base/vcs"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
)

// Filer is an interface for file tree file actions that all [Node]s satisfy.
type Filer interface { //types:add
	core.Treer

	// AsFileNode returns the [Node]
	AsFileNode() *Node

	// RenameFiles renames any selected files.
	RenameFiles()
}

var _ Filer = (*Node)(nil)

// OpenFilesDefault opens selected files with default app for that file type (os defined).
// runs open on Mac, xdg-open on Linux, and start on Windows
func (fn *Node) OpenFilesDefault() { //types:add
	fn.SelectedFunc(func(sn *Node) {
		sn.openFileDefault()
	})
}

// openFileDefault opens file with default app for that file type (os defined)
// runs open on Mac, xdg-open on Linux, and start on Windows
func (fn *Node) openFileDefault() error {
	core.TheApp.OpenURL("file://" + string(fn.Filepath))
	return nil
}

// duplicateFiles makes a copy of selected files
func (fn *Node) duplicateFiles() { //types:add
	fn.FileRoot.NeedsLayout()
	fn.SelectedFunc(func(sn *Node) {
		sn.duplicateFile()
	})
}

// duplicateFile creates a copy of given file -- only works for regular files, not
// directories
func (fn *Node) duplicateFile() error {
	_, err := fn.Info.Duplicate()
	if err == nil && fn.Parent != nil {
		fnp := AsNode(fn.Parent)
		fnp.Update()
	}
	return err
}

// deletes any selected files or directories. If any directory is selected,
// all files and subdirectories in that directory are also deleted.
func (fn *Node) deleteFiles() { //types:add
	d := core.NewBody().AddTitle("Delete Files?").
		AddText("Ok to delete file(s)?  This is not undoable and files are not moving to trash / recycle bin. If any selections are directories all files and subdirectories will also be deleted.")
	d.AddBottomBar(func(parent core.Widget) {
		d.AddCancel(parent)
		d.AddOK(parent).SetText("Delete Files").OnClick(func(e events.Event) {
			fn.deleteFilesImpl()
		})
	})
	d.RunDialog(fn)
}

// deleteFilesImpl does the actual deletion, no prompts
func (fn *Node) deleteFilesImpl() {
	fn.FileRoot.NeedsLayout()
	fn.SelectedFunc(func(sn *Node) {
		if !sn.Info.IsDir() {
			sn.deleteFile()
			return
		}
		var fns []string
		sn.Info.Filenames(&fns)
		ft := sn.FileRoot
		for _, filename := range fns {
			sn, ok := ft.FindFile(filename)
			if !ok {
				continue
			}
			if sn.Buffer != nil {
				sn.closeBuf()
			}
		}
		sn.deleteFile()
	})
}

// deleteFile deletes this file
func (fn *Node) deleteFile() error {
	if fn.isExternal() {
		return nil
	}
	pari := fn.Parent
	var parent *Node
	if pari != nil {
		parent = AsNode(pari)
	}
	fn.closeBuf()
	repo, _ := fn.Repo()
	var err error
	if !fn.Info.IsDir() && repo != nil && fn.Info.VCS >= vcs.Stored {
		// fmt.Printf("del repo: %v\n", fn.FPath)
		err = repo.Delete(string(fn.Filepath))
	} else {
		// fmt.Printf("del raw: %v\n", fn.FPath)
		err = fn.Info.Delete()
	}
	if err == nil {
		fn.Delete()
	}
	if parent != nil {
		parent.Update()
	}
	return err
}

// renames any selected files
func (fn *Node) RenameFiles() { //types:add
	fn.FileRoot.NeedsLayout()
	fn.SelectedFunc(func(sn *Node) {
		fb := core.NewSoloFuncButton(sn).SetFunc(sn.RenameFile)
		fb.Args[0].SetValue(sn.Name)
		fb.CallFunc()
	})
}

// RenameFile renames file to new name
func (fn *Node) RenameFile(newpath string) error { //types:add
	if fn.isExternal() {
		return nil
	}
	root := fn.FileRoot
	var err error
	fn.closeBuf() // invalid after this point
	orgpath := fn.Filepath
	newpath, err = fn.Info.Rename(newpath)
	if len(newpath) == 0 || err != nil {
		return err
	}
	if fn.IsDir() {
		if fn.FileRoot.isDirOpen(orgpath) {
			fn.FileRoot.setDirOpen(core.Filename(newpath))
		}
	}
	repo, _ := fn.Repo()
	stored := false
	if fn.IsDir() && !fn.HasChildren() {
		err = os.Rename(string(orgpath), newpath)
	} else if repo != nil && fn.Info.VCS >= vcs.Stored {
		stored = true
		err = repo.Move(string(orgpath), newpath)
	} else {
		err = os.Rename(string(orgpath), newpath)
		if err != nil && errors.Is(err, syscall.ENOENT) { // some kind of bogus error it seems?
			err = nil
		}
	}
	if err == nil {
		err = fn.Info.InitFile(newpath)
	}
	if err == nil {
		fn.Filepath = core.Filename(fn.Info.Path)
		fn.SetName(fn.Info.Name)
		fn.SetText(fn.Info.Name)
	}
	// todo: if you add orgpath here to git, then it will show the rename in status
	if stored {
		fn.AddToVCS()
	}
	if root != nil {
		root.UpdatePath(string(orgpath))
		root.UpdatePath(newpath)
	}
	return err
}

// newFiles makes a new file in selected directory
func (fn *Node) newFiles(filename string, addToVCS bool) { //types:add
	done := false
	fn.SelectedFunc(func(sn *Node) {
		if !done {
			sn.newFile(filename, addToVCS)
			done = true
		}
	})
}

// newFile makes a new file in this directory node
func (fn *Node) newFile(filename string, addToVCS bool) { //types:add
	if fn.isExternal() {
		return
	}
	ppath := string(fn.Filepath)
	if !fn.IsDir() {
		ppath, _ = filepath.Split(ppath)
	}
	np := filepath.Join(ppath, filename)
	_, err := os.Create(np)
	if err != nil {
		core.ErrorSnackbar(fn, err)
		return
	}
	if addToVCS {
		nfn, ok := fn.FileRoot.FindFile(np)
		if ok && nfn.This != fn.FileRoot.This && string(nfn.Filepath) == np {
			// todo: this is where it is erroneously adding too many files to vcs!
			fmt.Println("Adding new file to VCS:", nfn.Filepath)
			core.MessageSnackbar(fn, "Adding new file to VCS: "+fsx.DirAndFile(string(nfn.Filepath)))
			nfn.AddToVCS()
		}
	}
	fn.FileRoot.UpdatePath(np)
}

// makes a new folder in the given selected directory
func (fn *Node) newFolders(foldername string) { //types:add
	done := false
	fn.SelectedFunc(func(sn *Node) {
		if !done {
			sn.newFolder(foldername)
			done = true
		}
	})
}

// newFolder makes a new folder (directory) in this directory node
func (fn *Node) newFolder(foldername string) { //types:add
	if fn.isExternal() {
		return
	}
	ppath := string(fn.Filepath)
	if !fn.IsDir() {
		ppath, _ = filepath.Split(ppath)
	}
	np := filepath.Join(ppath, foldername)
	err := os.MkdirAll(np, 0775)
	if err != nil {
		core.ErrorSnackbar(fn, err)
		return
	}
	fn.FileRoot.UpdatePath(ppath)
}

// copyFileToDir copies given file path into node that is a directory.
// This does NOT check for overwriting -- that must be done at higher level!
func (fn *Node) copyFileToDir(filename string, perm os.FileMode) {
	if fn.isExternal() {
		return
	}
	ppath := string(fn.Filepath)
	sfn := filepath.Base(filename)
	tpath := filepath.Join(ppath, sfn)
	fileinfo.CopyFile(tpath, filename, perm)
	fn.FileRoot.UpdatePath(ppath)
	ofn, ok := fn.FileRoot.FindFile(filename)
	if ok && ofn.Info.VCS >= vcs.Stored {
		nfn, ok := fn.FileRoot.FindFile(tpath)
		if ok && nfn.This != fn.FileRoot.This {
			if string(nfn.Filepath) != tpath {
				fmt.Printf("error: nfn.FPath != tpath; %q != %q, see bug #453\n", nfn.Filepath, tpath)
			} else {
				nfn.AddToVCS() // todo: this sometimes is not just tpath!  See bug #453
			}
			nfn.Update()
		}
	}
}

// Shows file information about selected file(s)
func (fn *Node) showFileInfo() { //types:add
	fn.SelectedFunc(func(sn *Node) {
		d := core.NewBody().AddTitle("File info")
		core.NewForm(d).SetStruct(&sn.Info).SetReadOnly(true)
		d.AddOKOnly().RunFullDialog(sn)
	})
}
