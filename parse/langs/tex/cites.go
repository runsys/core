// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tex

import (
	"log"
	"strings"

	"cogentcore.org/core/icons"
	"cogentcore.org/core/parse"
	"cogentcore.org/core/parse/complete"
	"cogentcore.org/core/parse/langs/bibtex"
	"cogentcore.org/core/parse/lexer"
)

// CompleteCite does completion on citation
func (tl *TexLang) CompleteCite(fss *parse.FileStates, origStr, str string, pos lexer.Pos) (md complete.Matches) {
	bfile, has := fss.MetaData("bibfile")
	if !has {
		return
	}
	bf, err := tl.Bibs.Open(bfile)
	if err != nil {
		return
	}
	md.Seed = str
	for _, be := range bf.BibTex.Entries {
		if strings.HasPrefix(be.CiteName, str) {
			c := complete.Completion{Text: be.CiteName, Label: be.CiteName, Icon: icons.Field}
			md.Matches = append(md.Matches, c)
		}
	}
	return md
}

// LookupCite does lookup on citation
func (tl *TexLang) LookupCite(fss *parse.FileStates, origStr, str string, pos lexer.Pos) (ld complete.Lookup) {
	bfile, has := fss.MetaData("bibfile")
	if !has {
		return
	}
	bf, err := tl.Bibs.Open(bfile)
	if err != nil {
		return
	}
	lkbib := bibtex.NewBibTex()
	for _, be := range bf.BibTex.Entries {
		if strings.HasPrefix(be.CiteName, str) {
			lkbib.Entries = append(lkbib.Entries, be)
		}
	}
	if len(lkbib.Entries) > 0 {
		ld.SetFile(fss.Filename, 0, 0)
		ld.Text = []byte(lkbib.PrettyString())
	}
	return ld
}

// OpenBibfile attempts to open the /bibliography file.
// Sets meta data "bibfile" to resulting file if found, and deletes it if not.
func (tl *TexLang) OpenBibfile(fss *parse.FileStates, pfs *parse.FileState) error {
	bfile := tl.FindBibliography(pfs)
	if bfile == "" {
		fss.DeleteMetaData("bibfile")
		return nil
	}
	_, err := tl.Bibs.Open(bfile)
	if err != nil {
		log.Println(err)
		fss.DeleteMetaData("bibfile")
		return err
	}
	fss.SetMetaData("bibfile", bfile)
	return nil
}

func (tl *TexLang) FindBibliography(pfs *parse.FileState) string {
	nlines := pfs.Src.NLines()
	trg := `\bibliography{`
	trgln := len(trg)
	for i := nlines - 1; i >= 0; i-- {
		sln := pfs.Src.Lines[i]
		lnln := len(sln)
		if lnln == 0 {
			continue
		}
		if sln[0] != '\\' {
			continue
		}
		if lnln > 100 {
			continue
		}
		lstr := string(sln)
		if strings.HasPrefix(lstr, trg) {
			return lstr[trgln:len(sln)-1] + ".bib"
		}
	}
	return ""
}
