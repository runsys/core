// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fileinfo

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
)

// MimeNoChar returns the mime string without any charset
// encoding information, or anything else after a ;
func MimeNoChar(mime string) string {
	if sidx := strings.Index(mime, ";"); sidx > 0 {
		return strings.TrimSpace(mime[:sidx])
	}
	return mime
}

// MimeTop returns the top-level main type category from mime type
// i.e., the thing before the /  -- returns empty if no /
func MimeTop(mime string) string {
	if sidx := strings.Index(mime, "/"); sidx > 0 {
		return mime[:sidx]
	}
	return ""
}

// MimeSub returns the sub-level subtype category from mime type
// i.e., the thing after the /  -- returns empty if no /
// also trims off any charset encoding stuff
func MimeSub(mime string) string {
	if sidx := strings.Index(MimeNoChar(mime), "/"); sidx > 0 {
		return mime[sidx+1:]
	}
	return ""
}

// MimeFromFile gets mime type from file, using Gabriel Vasile's mimetype
// package, mime.TypeByExtension, the chroma syntax highlighter,
// CustomExtMimeMap, and finally FileExtMimeMap.  Use the mimetype package's
// extension mechanism to add further content-based matchers as needed, and
// set CustomExtMimeMap to your own map or call AddCustomExtMime for
// extension-based ones.
func MimeFromFile(fname string) (mtype, ext string, err error) {
	ext = strings.ToLower(filepath.Ext(fname))
	if mtyp, has := ExtMimeMap[ext]; has { // use our map first: very fast!
		return mtyp, ext, nil
	}
	_, fn := filepath.Split(fname)
	fc := fn[0]
	lc := fn[len(fn)-1]
	if fc == '~' || fc == '%' || fc == '#' || lc == '~' || lc == '%' || lc == '#' {
		return MimeString(Trash), ext, nil
	}
	mtypt, err := filetype.MatchFile(fn) // h2non next -- has good coverage
	ptyp := ""
	isplain := false
	if err == nil {
		mtyp := mtypt.MIME.Value
		ext = mtypt.Extension
		if strings.HasPrefix(mtyp, "text/plain") {
			isplain = true
			ptyp = mtyp
		} else {
			return mtyp, ext, nil
		}
	}
	mtyp := mime.TypeByExtension(ext)
	if mtyp != "" {
		return mtyp, ext, nil
	}
	// TODO(kai/binsize): figure out how to do this without dragging in chroma dependency
	// lexer := lexers.Match(fn) // todo: could get start of file and pass to
	// // Analyze, but might be too slow..
	// if lexer != nil {
	// 	config := lexer.Config()
	// 	if len(config.MimeTypes) > 0 {
	// 		mtyp = config.MimeTypes[0]
	// 		return mtyp, ext, nil
	// 	}
	// 	mtyp := "application/" + strings.ToLower(config.Name)
	// 	return mtyp, ext, nil
	// }
	if isplain {
		return ptyp, ext, nil
	}
	if strings.ToLower(fn) == "makefile" {
		return MimeString(Makefile), ext, nil
	}
	return "", ext, fmt.Errorf("fileinfo.MimeFromFile could not find mime type for ext: %v file: %v", ext, fn)
}

// todo: use this to check against mime types!

// MimeToKindMapInit makes sure the MimeToKindMap is initialized from
// InitMimeToKindMap plus chroma lexer types.
// func MimeToKindMapInit() {
// 	if MimeToKindMap != nil {
// 		return
// 	}
// 	MimeToKindMap = InitMimeToKindMap
// 	for _, l := range lexers.Registry.Lexers {
// 		config := l.Config()
// 		nm := strings.ToLower(config.Name)
// 		if len(config.MimeTypes) > 0 {
// 			mtyp := config.MimeTypes[0]
// 			MimeToKindMap[mtyp] = nm
// 		} else {
// 			MimeToKindMap["application/"+nm] = nm
// 		}
// 	}
// }

//////////////////////////////////////////////////////////////////////////////
//    Mime types

// ExtMimeMap is the map from extension to mime type, built from AvailMimes
var ExtMimeMap = map[string]string{}

// MimeType contains all the information associated with a given mime type
type MimeType struct {

	// mimetype string: type/subtype
	Mime string

	// file extensions associated with this file type
	Exts []string

	// category of file
	Cat Categories

	// if known, the name of the known file type, else NoSupporUnknown
	Sup Known
}

// CustomMimes can be set by other apps to contain custom mime types that
// go beyond what is in the standard ones, and can also redefine and
// override the standard one
var CustomMimes []MimeType

// AvailableMimes is the full list (as a map from mimetype) of available defined mime types
// built from StdMimes (compiled in) and then CustomMimes can override
var AvailableMimes map[string]MimeType

// MimeKnown returns the known type for given mime key,
// or Unknown if not found or not a known file type
func MimeKnown(mime string) Known {
	mt, has := AvailableMimes[MimeNoChar(mime)]
	if !has {
		return Unknown
	}
	return mt.Sup
}

// ExtKnown returns the known type for given file extension,
// or Unknown if not found or not a known file type
func ExtKnown(ext string) Known {
	mime, has := ExtMimeMap[ext]
	if !has {
		return Unknown
	}
	mt, has := AvailableMimes[mime]
	if !has {
		return Unknown
	}
	return mt.Sup
}

// KnownFromFile returns the known type for given file,
// or Unknown if not found or not a known file type
func KnownFromFile(fname string) Known {
	mtyp, _, err := MimeFromFile(fname)
	if err != nil {
		return Unknown
	}
	return MimeKnown(mtyp)
}

// MergeAvailableMimes merges the StdMimes and CustomMimes into AvailMimes
// if CustomMimes is updated, then this should be called -- initially
// it just has StdMimes.
// It also builds the ExtMimeMap to map from extension to mime type
// and KnownMimes map of known file types onto their full
// mime type entry
func MergeAvailableMimes() {
	AvailableMimes = make(map[string]MimeType, len(StandardMimes)+len(CustomMimes))
	for _, mt := range StandardMimes {
		AvailableMimes[mt.Mime] = mt
	}
	for _, mt := range CustomMimes {
		AvailableMimes[mt.Mime] = mt // overwrite automatically
	}
	ExtMimeMap = make(map[string]string) // start over
	KnownMimes = make(map[Known]MimeType)
	for _, mt := range AvailableMimes {
		if len(mt.Exts) > 0 { // first pass add only ext guys to support
			for _, ex := range mt.Exts {
				if ex[0] != '.' {
					fmt.Printf("fileinfo.MergeAvailMimes: ext: %v does not start with a . in type: %v\n", ex, mt.Mime)
				}
				if pmt, has := ExtMimeMap[ex]; has {
					fmt.Printf("fileinfo.MergeAvailMimes: non-unique ext: %v assigned to mime type: %v AND %v\n", ex, pmt, mt.Mime)
				} else {
					ExtMimeMap[ex] = mt.Mime
				}
			}
			if mt.Sup != Unknown {
				if hsp, has := KnownMimes[mt.Sup]; has {
					fmt.Printf("fileinfo.MergeAvailMimes: more-than-one mimetype has extensions for same known file type: %v -- one: %v other %v\n", mt.Sup, hsp.Mime, mt.Mime)
				} else {
					KnownMimes[mt.Sup] = mt
				}
			}
		}
	}
	// second pass to get any known guys that don't have exts
	for _, mt := range AvailableMimes {
		if mt.Sup != Unknown {
			if _, has := KnownMimes[mt.Sup]; !has {
				KnownMimes[mt.Sup] = mt
			}
		}
	}
}

func init() {
	MergeAvailableMimes()
}

// http://www.iana.org/assignments/media-types/media-types.xhtml
// https://github.com/mirage/ocaml-magic-mime/blob/master/x-mime.types
// https://www.apt-browse.org/browse/debian/stretch/main/all/mime-support/3.60/file/etc/mime.types
// https://developer.apple.com/library/archive/documentation/Miscellaneous/Reference/UTIRef/Articles/System-DeclaredUniformTypeIdentifiers.html
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Complete_list_of_MIME_types

// StandardMimes is the full list of standard mime types compiled into our code
// various other maps etc are constructed from it.
// When there are multiple types associated with the same real type, pick one
// to be the canonical one and give it, and *only* it, the extensions!
var StandardMimes = []MimeType{
	// Folder
	{"text/directory", nil, Folder, Unknown},

	// Archive
	{"multipart/mixed", nil, Archive, Multipart},
	{"application/tar", []string{".tar", ".tar.gz", ".tgz", ".taz", ".taZ", ".tar.bz2", ".tz2", ".tbz2", ".tbz", ".tar.lz", ".tar.lzma", ".tlz", ".tar.lzop", ".tar.xz"}, Archive, Tar},
	{"application/x-gtar", nil, Archive, Tar},
	{"application/x-gtar-compressed", nil, Archive, Tar},
	{"application/x-tar", nil, Archive, Tar},

	{"application/zip", []string{".zip"}, Archive, Zip},
	{"application/gzip", []string{".gz"}, Archive, GZip},
	{"application/x-7z-compressed", []string{".7z"}, Archive, SevenZ},
	{"application/x-xz", []string{".xz"}, Archive, Xz},
	{"application/x-bzip", []string{".bz", ".bz2"}, Archive, BZip},
	{"application/x-bzip2", nil, Archive, BZip},

	{"application/x-apple-diskimage", []string{".dmg"}, Archive, Dmg},
	{"application/x-shar", []string{".shar"}, Archive, Shar},

	{"application/x-bittorrent", []string{".torrent"}, Archive, Unknown},
	{"application/rar", []string{".rar"}, Archive, Unknown},
	{"application/x-stuffit", []string{".sit", ".sitx"}, Archive, Unknown},

	{"application/vnd.android.package-archive", []string{".apk"}, Archive, Unknown},
	{"application/vnd.debian.binary-package", []string{".deb", ".ddeb", ".udeb"}, Archive, Unknown},
	{"application/x-debian-package", nil, Archive, Unknown},
	{"application/x-redhat-package-manager", []string{".rpm"}, Archive, Unknown},
	{"text/x-rpm-spec", nil, Archive, Unknown},

	// Backup
	{"application/x-trash", []string{".bak", ".old", ".sik"}, Backup, Trash}, // also "~", "%", "#",

	// Code -- use text/ as main instead of application as there are more text
	{"text/x-ada", []string{".adb", ".ads", ".ada"}, Code, Ada},
	{"text/x-asp", []string{".aspx", ".asax", ".ascx", ".ashx", ".asmx", ".axd"}, Code, Unknown},

	{"text/x-sh", []string{".bash", ".sh"}, Code, Bash},
	{"application/x-sh", nil, Code, Bash},

	{"text/x-csrc", []string{".c", ".C", ".c++", ".cpp", ".cxx", ".cc", ".h", ".h++", ".hpp", ".hxx", ".hh", ".hlsl", ".gsl", ".frag", ".vert", ".mm"}, Code, C}, // this is apparently the main one now
	{"text/x-chdr", nil, Code, C},
	{"text/x-c", nil, Code, C},
	{"text/x-c++hdr", nil, Code, C},
	{"text/x-c++src", nil, Code, C},
	{"text/x-chdr", nil, Code, C},
	{"text/x-cpp", nil, Code, C},

	{"text/x-csh", []string{".csh"}, Code, Csh},
	{"application/x-csh", nil, Code, Csh},

	{"text/x-csharp", []string{".cs"}, Code, CSharp},
	{"text/x-dsrc", []string{".d"}, Code, D},
	{"text/x-diff", []string{".diff", ".patch"}, Code, Diff},
	{"text/x-eiffel", []string{".e"}, Code, Eiffel},
	{"text/x-erlang", []string{".erl", ".hrl", ".escript"}, Code, Erlang}, // note: ".es" conflicts with ecmascript
	{"text/x-forth", []string{".frt"}, Code, Forth},                       // note: ".fs" conflicts with fsharp
	{"text/x-fortran", []string{".f", ".F"}, Code, Fortran},
	{"text/x-fsharp", []string{".fs", ".fsi"}, Code, FSharp},
	{"text/x-gosrc", []string{".go", ".mod", ".work", ".cosh"}, Code, Go},
	{"text/x-haskell", []string{".hs", ".lhs"}, Code, Haskell},
	{"text/x-literate-haskell", nil, Code, Haskell}, // todo: not sure if same or not

	{"text/x-java", []string{".java", ".jar"}, Code, Java},
	{"application/java-archive", nil, Code, Java},
	{"application/javascript", []string{".js"}, Code, JavaScript},
	{"application/ecmascript", []string{".es"}, Code, Unknown},

	{"text/x-common-lisp", []string{".lisp", ".cl", ".el"}, Code, Lisp},
	{"text/elisp", nil, Code, Lisp},
	{"text/x-elisp", nil, Code, Lisp},
	{"application/emacs-lisp", nil, Code, Lisp},

	{"text/x-lua", []string{".lua", ".wlua"}, Code, Lua},

	{"text/x-makefile", nil, Code, Makefile},
	{"text/x-autoconf", nil, Code, Makefile},

	{"text/x-moc", []string{".moc"}, Code, Unknown},

	{"application/mathematica", []string{".nb", ".nbp"}, Code, Mathematica},

	{"text/x-matlab", []string{".m"}, Code, Matlab},
	{"text/matlab", nil, Code, Matlab},
	{"text/octave", nil, Code, Matlab},
	{"text/scilab", []string{".sci", ".sce", ".tst"}, Code, Unknown},

	{"text/x-modelica", []string{".mo"}, Code, Unknown},
	{"text/x-nemerle", []string{".n"}, Code, Unknown},

	{"text/x-objcsrc", nil, Code, ObjC}, // doesn't have chroma support -- use C instead
	{"text/x-objective-j", nil, Code, Unknown},

	{"text/x-ocaml", []string{".ml", ".mli", ".mll", ".mly"}, Code, OCaml},
	{"text/x-pascal", []string{".p", ".pas"}, Code, Pascal},
	{"text/x-perl", []string{".pl", ".pm"}, Code, Perl},
	{"text/x-php", []string{".php", ".php3", ".php4", ".php5", ".inc"}, Code, Php},
	{"text/x-prolog", []string{".ecl", ".prolog", ".pro"}, Code, Prolog}, // note: ".pl" conflicts

	{"text/x-python", []string{".py", ".pyc", ".pyo", ".pyw"}, Code, Python},
	{"application/x-python-code", nil, Code, Python},

	{"text/x-rust", []string{".rs"}, Code, Rust},
	{"text/rust", nil, Code, Rust},

	{"text/x-r", []string{".r", ".S", ".R", ".Rhistory", ".Rprofile", ".Renviron"}, Code, R},
	{"text/x-R", nil, Code, R},
	{"text/S-Plus", nil, Code, R},
	{"text/S", nil, Code, R},
	{"text/x-r-source", nil, Code, R},
	{"text/x-r-history", nil, Code, R},
	{"text/x-r-profile", nil, Code, R},

	{"text/x-ruby", []string{".rb"}, Code, Ruby},
	{"application/x-ruby", nil, Code, Ruby},
	{"text/x-scala", []string{".scala"}, Code, Scala},
	{"text/x-tcl", []string{".tcl", ".tk"}, Code, Tcl},
	{"application/x-tcl", nil, Code, Tcl},

	// Doc
	{"text/x-bibtex", []string{".bib"}, Doc, BibTeX},
	{"text/x-tex", []string{".tex", ".ltx", ".sty", ".cls", ".latex"}, Doc, TeX},
	{"application/x-latex", nil, Doc, TeX},

	{"application/x-texinfo", []string{".texinfo", ".texi"}, Doc, Texinfo},

	{"application/x-troff", []string{".t", ".tr", ".roff", ".man", ".me", ".ms"}, Doc, Troff},
	{"application/x-troff-man", nil, Doc, Troff},
	{"application/x-troff-me", nil, Doc, Troff},
	{"application/x-troff-ms", nil, Doc, Troff},

	{"text/html", []string{".html", ".htm", ".shtml", ".xhtml", ".xht"}, Doc, Html},
	{"application/xhtml+xml", nil, Doc, Html},
	{"text/mathml", []string{".mml"}, Doc, Unknown},
	{"text/css", []string{".css"}, Doc, Css},

	{"text/markdown", []string{".md", ".markdown"}, Doc, Markdown},
	{"text/x-markdown", nil, Doc, Markdown},

	{"application/rtf", []string{".rtf"}, Doc, Rtf},
	{"text/richtext", []string{".rtx"}, Doc, Unknown},

	{"application/mbox", []string{".mbox"}, Doc, Unknown},
	{"application/x-rss+xml", []string{".rss"}, Doc, Unknown},

	{"application/msword", []string{".doc", ".dot", ".docx", ".dotx"}, Doc, MSWord},
	{"application/vnd.ms-word", nil, Doc, MSWord},
	{"application/vnd.openxmlformats-officedocument.wordprocessingml.document", nil, Doc, MSWord},
	{"application/vnd.openxmlformats-officedocument.wordprocessingml.template", nil, Doc, MSWord},

	{"application/vnd.oasis.opendocument.text", []string{".odt", ".odm", ".ott", ".oth", ".sxw", ".sxg", ".stw", ".sxm"}, Doc, OpenText},
	{"application/vnd.oasis.opendocument.text-master", nil, Doc, OpenText},
	{"application/vnd.oasis.opendocument.text-template", nil, Doc, OpenText},
	{"application/vnd.oasis.opendocument.text-web", nil, Doc, OpenText},
	{"application/vnd.sun.xml.writer", nil, Doc, OpenText},
	{"application/vnd.sun.xml.writer.global", nil, Doc, OpenText},
	{"application/vnd.sun.xml.writer.template", nil, Doc, OpenText},
	{"application/vnd.sun.xml.math", nil, Doc, OpenText},

	{"application/vnd.oasis.opendocument.presentation", []string{".odp", ".otp", ".sxi", ".sti"}, Doc, OpenPres},
	{"application/vnd.oasis.opendocument.presentation-template", nil, Doc, OpenPres},
	{"application/vnd.sun.xml.impress", nil, Doc, OpenPres},
	{"application/vnd.sun.xml.impress.template", nil, Doc, OpenPres},

	{"application/vnd.ms-powerpoint", []string{".ppt", ".pps", ".pptx", ".sldx", ".ppsx", ".potx"}, Doc, MSPowerpoint},
	{"application/vnd.openxmlformats-officedocument.presentationml.presentation", nil, Doc, MSPowerpoint},
	{"application/vnd.openxmlformats-officedocument.presentationml.slide", nil, Doc, MSPowerpoint},
	{"application/vnd.openxmlformats-officedocument.presentationml.slideshow", nil, Doc, MSPowerpoint},
	{"application/vnd.openxmlformats-officedocument.presentationml.template", nil, Doc, MSPowerpoint},

	{"application/ms-tnef", nil, Doc, Unknown},
	{"application/vnd.ms-tnef", nil, Doc, Unknown},

	{"application/onenote", []string{".one", ".onetoc2", ".onetmp", ".onepkg"}, Doc, Unknown},

	{"application/pgp-encrypted", []string{".pgp"}, Doc, Unknown},
	{"application/pgp-keys", []string{".key"}, Doc, Unknown},
	{"application/pgp-signature", []string{".sig"}, Doc, Unknown},

	{"application/vnd.amazon.ebook", []string{".azw"}, Doc, EBook},
	{"application/epub+zip", []string{".epub"}, Doc, EPub},

	// Sheet
	{"application/vnd.ms-excel", []string{".xls", ".xlb", ".xlt", ".xlsx", ".xltx"}, Sheet, MSExcel},
	{"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil, Sheet, MSExcel},
	{"application/vnd.openxmlformats-officedocument.spreadsheetml.template", nil, Sheet, MSExcel},

	{"application/vnd.oasis.opendocument.spreadsheet", []string{".ods", ".ots", ".sxc", ".stc", ".odf"}, Sheet, OpenSheet},
	{"application/vnd.oasis.opendocument.spreadsheet-template", nil, Sheet, OpenSheet},
	{"application/vnd.oasis.opendocument.formula", nil, Sheet, OpenSheet}, // todo: could be separate
	{"application/vnd.sun.xml.calc", nil, Sheet, OpenSheet},
	{"application/vnd.sun.xml.calc.template", nil, Sheet, OpenSheet},

	// Data
	{"text/csv", []string{".csv"}, Data, Csv},
	{"application/json", []string{".json"}, Data, Json},
	{"application/xml", []string{".xml", ".xsd"}, Data, Xml},
	{"text/xml", nil, Data, Xml},
	{"text/x-protobuf", []string{".proto"}, Data, Protobuf},
	{"text/x-ini", []string{".ini", ".cfg", ".inf"}, Data, Ini},
	{"text/x-ini-file", nil, Data, Ini},
	{"text/uri-list", nil, Data, Uri},
	{"application/x-color", nil, Data, Color},

	{"application/rdf+xml", []string{".rdf"}, Data, Unknown},
	{"application/msaccess", []string{".mdb"}, Data, Unknown},
	{"application/vnd.oasis.opendocument.database", []string{".odb"}, Data, Unknown},
	{"text/tab-separated-values", []string{".tsv"}, Data, Tsv},
	{"application/vnd.google-earth.kml+xml", []string{".kml", ".kmz"}, Data, Unknown},
	{"application/vnd.google-earth.kmz", nil, Data, Unknown},
	{"application/x-sql", []string{".sql"}, Data, Unknown},

	// Text
	{"text/plain", []string{".asc", ".txt", ".text", ".pot", ".brf", ".srt"}, Text, PlainText},
	{"text/cache-manifest", []string{".appcache"}, Text, Unknown},
	{"text/calendar", []string{".ics", ".icz"}, Text, ICal},
	{"text/x-vcalendar", []string{".vcs"}, Text, VCal},
	{"text/vcard", []string{".vcf", ".vcard"}, Text, VCard},

	// Image
	{"application/pdf", []string{".pdf"}, Image, Pdf},
	{"application/postscript", []string{".ps", ".ai", ".eps", ".epsi", ".epsf", ".eps2", ".eps3"}, Image, Postscript},
	{"application/vnd.oasis.opendocument.graphics", []string{".odc", ".odg", ".otg", ".odi", ".sxd", ".std"}, Image, Unknown},
	{"application/vnd.oasis.opendocument.chart", nil, Image, Unknown},
	{"application/vnd.oasis.opendocument.graphics-template", nil, Image, Unknown},
	{"application/vnd.oasis.opendocument.image", nil, Image, Unknown},
	{"application/vnd.sun.xml.draw", nil, Image, Unknown},
	{"application/vnd.sun.xml.draw.template", nil, Image, Unknown},
	{"application/x-xfig", []string{".fig"}, Image, Unknown},
	{"application/x-xcf", []string{".xcf"}, Image, Gimp},
	{"text/vnd.graphviz", []string{".gv"}, Image, GraphVis},

	{"image/gif", []string{".gif"}, Image, Gif},
	{"image/ief", []string{".ief"}, Image, Unknown},
	{"image/jp2", []string{".jp2", ".jpg2"}, Image, Unknown},
	{"image/jpeg", []string{".jpeg", ".jpg", ".jpe"}, Image, Jpeg},
	{"image/jpm", []string{".jpm"}, Image, Unknown},
	{"image/jpx", []string{".jpx", ".jpf"}, Image, Unknown},
	{"image/pcx", []string{".pcx"}, Image, Unknown},
	{"image/png", []string{".png"}, Image, Png},
	{"image/heic", []string{".heic"}, Image, Heic},
	{"image/heif", []string{".heif"}, Image, Heif},
	{"image/svg+xml", []string{".svg", ".svgz"}, Image, Svg},
	{"image/tiff", []string{".tiff", ".tif"}, Image, Tiff},
	{"image/vnd.djvu", []string{".djvu", ".djv"}, Image, Unknown},
	{"image/vnd.microsoft.icon", []string{".ico"}, Image, Unknown},
	{"image/vnd.wap.wbmp", []string{".wbmp"}, Image, Unknown},
	{"image/x-canon-cr2", []string{".cr2"}, Image, Unknown},
	{"image/x-canon-crw", []string{".crw"}, Image, Unknown},
	{"image/x-cmu-raster", []string{".ras"}, Image, Unknown},
	{"image/x-coreldraw", []string{".cdr", ".pat", ".cdt", ".cpt"}, Image, Unknown},
	{"image/x-coreldrawpattern", nil, Image, Unknown},
	{"image/x-coreldrawtemplate", nil, Image, Unknown},
	{"image/x-corelphotopaint", nil, Image, Unknown},
	{"image/x-epson-erf", []string{".erf"}, Image, Unknown},
	{"image/x-jg", []string{".art"}, Image, Unknown},
	{"image/x-jng", []string{".jng"}, Image, Unknown},
	{"image/x-ms-bmp", []string{".bmp"}, Image, Bmp},
	{"image/x-nikon-nef", []string{".nef"}, Image, Unknown},
	{"image/x-olympus-orf", []string{".orf"}, Image, Unknown},
	{"image/x-photoshop", []string{".psd"}, Image, Unknown},
	{"image/x-portable-anymap", []string{".pnm"}, Image, Pnm},
	{"image/x-portable-bitmap", []string{".pbm"}, Image, Pbm},
	{"image/x-portable-graymap", []string{".pgm"}, Image, Pgm},
	{"image/x-portable-pixmap", []string{".ppm"}, Image, Ppm},
	{"image/x-rgb", []string{".rgb"}, Image, Unknown},
	{"image/x-xbitmap", []string{".xbm"}, Image, Xbm},
	{"image/x-xpixmap", []string{".xpm"}, Image, Xpm},
	{"image/x-xwindowdump", []string{".xwd"}, Image, Unknown},

	// Model
	{"model/iges", []string{".igs", ".iges"}, Model, Unknown},
	{"model/mesh", []string{".msh", ".mesh", ".silo"}, Model, Unknown},
	{"model/vrml", []string{".wrl", ".vrml", ".vrm"}, Model, Vrml},
	{"x-world/x-vrml", nil, Model, Vrml},
	{"model/x3d+xml", []string{".x3dv", ".x3d", ".x3db"}, Model, X3d},
	{"model/x3d+vrml", nil, Model, X3d},
	{"model/x3d+binary", nil, Model, X3d},

	// Audio
	{"audio/aac", []string{".aac"}, Audio, Aac},
	{"audio/flac", []string{".flac"}, Audio, Flac},
	{"audio/mpeg", []string{".mpga", ".mpega", ".mp2", ".mp3", ".m4a"}, Audio, Mp3},
	{"audio/mpegurl", []string{".m3u"}, Audio, Unknown},
	{"audio/x-mpegurl", nil, Audio, Unknown},
	{"audio/ogg", []string{".oga", ".ogg", ".opus", ".spx"}, Audio, Ogg},
	{"audio/amr", []string{".amr"}, Audio, Unknown},
	{"audio/amr-wb", []string{".awb"}, Audio, Unknown},
	{"audio/annodex", []string{".axa"}, Audio, Unknown},
	{"audio/basic", []string{".au", ".snd"}, Audio, Unknown},
	{"audio/csound", []string{".csd", ".orc", ".sco"}, Audio, Unknown},
	{"audio/midi", []string{".mid", ".midi", ".kar"}, Audio, Midi},
	{"audio/prs.sid", []string{".sid"}, Audio, Unknown},
	{"audio/x-aiff", []string{".aif", ".aiff", ".aifc"}, Audio, Unknown},
	{"audio/x-gsm", []string{".gsm"}, Audio, Unknown},
	{"audio/x-ms-wma", []string{".wma"}, Audio, Unknown},
	{"audio/x-ms-wax", []string{".wax"}, Audio, Unknown},
	{"audio/x-pn-realaudio", []string{".ra", ".rm", ".ram"}, Audio, Unknown},
	{"audio/x-realaudio", nil, Audio, Unknown},
	{"audio/x-scpls", []string{".pls"}, Audio, Unknown},
	{"audio/x-sd2", []string{".sd2"}, Audio, Unknown},
	{"audio/x-wav", []string{".wav"}, Audio, Wav},

	// Video
	{"video/3gpp", []string{".3gp"}, Video, Unknown},
	{"video/annodex", []string{".axv"}, Video, Unknown},
	{"video/dl", []string{".dl"}, Video, Unknown},
	{"video/dv", []string{".dif", ".dv"}, Video, Unknown},
	{"video/fli", []string{".fli"}, Video, Unknown},
	{"video/gl", []string{".gl"}, Video, Unknown},
	{"video/h264", nil, Video, Unknown},
	{"video/mpeg", []string{".mpeg", ".mpg", ".mpe"}, Video, Mpeg},
	{"video/MP2T", []string{".ts"}, Video, Unknown},
	{"video/mp4", []string{".mp4"}, Video, Mp4},
	{"video/quicktime", []string{".qt", ".mov"}, Video, Mov},
	{"video/ogg", []string{".ogv"}, Video, Ogv},
	{"video/webm", []string{".webm"}, Video, Unknown},
	{"video/vnd.mpegurl", []string{".mxu"}, Video, Unknown},
	{"video/x-flv", []string{".flv"}, Video, Unknown},
	{"video/x-la-asf", []string{".lsf", ".lsx"}, Video, Unknown},
	{"video/x-mng", []string{".mng"}, Video, Unknown},
	{"video/x-ms-asf", []string{".asf", ".asx"}, Video, Unknown},
	{"video/x-ms-wm", []string{".wm"}, Video, Unknown},
	{"video/x-ms-wmv", []string{".wmv"}, Video, Wmv},
	{"video/x-ms-wmx", []string{".wmx"}, Video, Unknown},
	{"video/x-ms-wvx", []string{".wvx"}, Video, Unknown},
	{"video/x-msvideo", []string{".avi"}, Video, Avi},
	{"video/x-sgi-movie", []string{".movie"}, Video, Unknown},
	{"video/x-matroska", []string{".mpv", ".mkv"}, Video, Unknown},
	{"application/x-shockwave-flash", []string{".swf"}, Video, Unknown},

	// Font
	{"font/ttf", []string{".otf", ".ttf", ".ttc"}, Font, TrueType},
	{"font/otf", nil, Font, TrueType},
	{"application/font-sfnt", nil, Font, TrueType},
	{"application/x-font-ttf", nil, Font, TrueType},

	{"application/x-font", []string{".pfa", ".pfb", ".gsf", ".pcf", ".pcf.Z"}, Font, Unknown},
	{"application/x-font-pcf", nil, Font, Unknown},
	{"application/vnd.ms-fontobject", []string{".eot"}, Font, Unknown},

	{"font/woff", []string{".woff", ".woff2"}, Font, WebOpenFont},
	{"font/woff2", nil, Font, WebOpenFont},
	{"application/font-woff", nil, Font, WebOpenFont},

	// Exe
	{"application/x-executable", nil, Exe, Unknown},
	{"application/x-msdos-program", []string{".com", ".exe", ".bat", ".dll"}, Exe, Unknown},

	// Binary
	{"application/octet-stream", []string{".bin"}, Bin, Unknown},
	{"application/x-object", []string{".o"}, Bin, Unknown},
	{"text/x-libtool", nil, Bin, Unknown},
}

// below are entries from official /etc/mime.types that we don't recognize
// or consider to be old / obsolete / not relevant -- please file an issue
// or a pull-request to add to main list or add yourself in your own app

// application/activemessage
// application/andrew-insetez
// application/annodexanx
// application/applefile
// application/atom+xmlatom
// application/atomcat+xmlatomcat
// application/atomicmail
// application/atomserv+xmlatomsrv
// application/batch-SMTP
// application/bbolinlin
// application/beep+xml
// application/cals-1840
// application/commonground
// application/cu-seemecu
// application/cybercash
// application/davmount+xmldavmount
// application/dca-rft
// application/dec-dx
// application/dicomdcm
// application/docbook+xml
// application/dsptypetsp
// application/dvcs
// application/edi-consent
// application/edi-x12
// application/edifact
// application/eshop
// application/font-tdpfrpfr
// application/futuresplashspl
// application/ghostview
// application/htahta
// application/http
// application/hyperstudio
// application/iges
// application/index
// application/index.cmd
// application/index.obj
// application/index.response
// application/index.vnd
// application/iotp
// application/ipp
// application/isup
// application/java-serialized-objectser
// application/java-vmclass
// application/m3gm3g
// application/mac-binhex40hqx
// application/mac-compactprocpt
// application/macwriteii
// application/marc
// application/mxfmxf
// application/news-message-id
// application/news-transmission
// application/ocsp-request
// application/ocsp-response
// application/octet-streambin deploy msu msp
// application/odaoda
// application/oebps-package+xmlopf
// application/oggogx
// application/parityfec
// application/pics-rulesprf
// application/pkcs10
// application/pkcs7-mime
// application/pkcs7-signature
// application/pkix-cert
// application/pkix-crl
// application/pkixcmp
// application/prs.alvestrand.titrax-sheet
// application/prs.cww
// application/prs.nprend
// application/qsig
// application/remote-printing
// application/riscos
// application/sdp
// application/set-payment
// application/set-payment-initiation
// application/set-registration
// application/set-registration-initiation
// application/sgml
// application/sgml-open-catalog
// application/sieve
// application/slastl
// application/slate
// application/smil+xmlsmi smil
// application/timestamp-query
// application/timestamp-reply
// application/vemmi
// application/whoispp-query
// application/whoispp-response
// application/wita
// application/x400-bp
// application/xml-dtd
// application/xml-external-parsed-entity
// application/xslt+xmlxsl xslt
// application/xspf+xmlxspf
// application/vnd.3M.Post-it-Notes
// application/vnd.accpac.simply.aso
// application/vnd.accpac.simply.imp
// application/vnd.acucobol
// application/vnd.aether.imp
// application/vnd.anser-web-certificate-issue-initiation
// application/vnd.anser-web-funds-transfer-initiation
// application/vnd.audiograph
// application/vnd.bmi
// application/vnd.businessobjects
// application/vnd.canon-cpdl
// application/vnd.canon-lips
// application/vnd.cinderellacdy
// application/vnd.claymore
// application/vnd.commerce-battelle
// application/vnd.commonspace
// application/vnd.comsocaller
// application/vnd.contact.cmsg
// application/vnd.cosmocaller
// application/vnd.ctc-posml
// application/vnd.cups-postscript
// application/vnd.cups-raster
// application/vnd.cups-raw
// application/vnd.cybank
// application/vnd.dna
// application/vnd.dpgraph
// application/vnd.dxr
// application/vnd.ecdis-update
// application/vnd.ecowin.chart
// application/vnd.ecowin.filerequest
// application/vnd.ecowin.fileupdate
// application/vnd.ecowin.series
// application/vnd.ecowin.seriesrequest
// application/vnd.ecowin.seriesupdate
// application/vnd.enliven
// application/vnd.epson.esf
// application/vnd.epson.msf
// application/vnd.epson.quickanime
// application/vnd.epson.salt
// application/vnd.epson.ssf
// application/vnd.ericsson.quickcall
// application/vnd.eudora.data
// application/vnd.fdf
// application/vnd.ffsns
// application/vnd.flographit
// application/vnd.font-fontforge-sfdsfd
// application/vnd.framemaker
// application/vnd.fsc.weblaunch
// application/vnd.fujitsu.oasys
// application/vnd.fujitsu.oasys2
// application/vnd.fujitsu.oasys3
// application/vnd.fujitsu.oasysgp
// application/vnd.fujitsu.oasysprs
// application/vnd.fujixerox.ddd
// application/vnd.fujixerox.docuworks
// application/vnd.fujixerox.docuworks.binder
// application/vnd.fut-misnet
// application/vnd.grafeq
// application/vnd.groove-account
// application/vnd.groove-identity-message
// application/vnd.groove-injector
// application/vnd.groove-tool-message
// application/vnd.groove-tool-template
// application/vnd.groove-vcard
// application/vnd.hhe.lesson-player
// application/vnd.hp-HPGL
// application/vnd.hp-PCL
// application/vnd.hp-PCLXL
// application/vnd.hp-hpid
// application/vnd.hp-hps
// application/vnd.httphone
// application/vnd.hzn-3d-crossword
// application/vnd.ibm.MiniPay
// application/vnd.ibm.afplinedata
// application/vnd.ibm.modcap
// application/vnd.informix-visionary
// application/vnd.intercon.formnet
// application/vnd.intertrust.digibox
// application/vnd.intertrust.nncp
// application/vnd.intu.qbo
// application/vnd.intu.qfx
// application/vnd.irepository.package+xml
// application/vnd.is-xpr
// application/vnd.japannet-directory-service
// application/vnd.japannet-jpnstore-wakeup
// application/vnd.japannet-payment-wakeup
// application/vnd.japannet-registration
// application/vnd.japannet-registration-wakeup
// application/vnd.japannet-setstore-wakeup
// application/vnd.japannet-verification
// application/vnd.japannet-verification-wakeup
// application/vnd.koan
// application/vnd.lotus-1-2-3
// application/vnd.lotus-approach
// application/vnd.lotus-freelance
// application/vnd.lotus-notes
// application/vnd.lotus-organizer
// application/vnd.lotus-screencam
// application/vnd.lotus-wordpro
// application/vnd.mcd
// application/vnd.mediastation.cdkey
// application/vnd.meridian-slingshot
// application/vnd.mif
// application/vnd.minisoft-hp3000-save
// application/vnd.mitsubishi.misty-guard.trustweb
// application/vnd.mobius.daf
// application/vnd.mobius.dis
// application/vnd.mobius.msl
// application/vnd.mobius.plc
// application/vnd.mobius.txf
// application/vnd.motorola.flexsuite
// application/vnd.motorola.flexsuite.adsi
// application/vnd.motorola.flexsuite.fis
// application/vnd.motorola.flexsuite.gotap
// application/vnd.motorola.flexsuite.kmr
// application/vnd.motorola.flexsuite.ttc
// application/vnd.motorola.flexsuite.wem
// application/vnd.mozilla.xul+xmlxul
// application/vnd.ms-artgalry
// application/vnd.ms-asf
// application/vnd.ms-excel.addin.macroEnabled.12xlam
// application/vnd.ms-excel.sheet.binary.macroEnabled.12xlsb
// application/vnd.ms-excel.sheet.macroEnabled.12xlsm
// application/vnd.ms-excel.template.macroEnabled.12xltm
// application/vnd.ms-fontobjecteot
// application/vnd.ms-lrm
// application/vnd.ms-officethemethmx
// application/vnd.ms-pki.seccatcat
// #application/vnd.ms-pki.stlstl
// application/vnd.ms-powerpoint.addin.macroEnabled.12ppam
// application/vnd.ms-powerpoint.presentation.macroEnabled.12pptm
// application/vnd.ms-powerpoint.slide.macroEnabled.12sldm
// application/vnd.ms-powerpoint.slideshow.macroEnabled.12ppsm
// application/vnd.ms-powerpoint.template.macroEnabled.12potm
// application/vnd.ms-project
// application/vnd.ms-word.document.macroEnabled.12docm
// application/vnd.ms-word.template.macroEnabled.12dotm
// application/vnd.ms-works
// application/vnd.mseq
// application/vnd.msign
// application/vnd.music-niff
// application/vnd.musician
// application/vnd.netfpx
// application/vnd.noblenet-directory
// application/vnd.noblenet-sealer
// application/vnd.noblenet-web
// application/vnd.novadigm.EDM
// application/vnd.novadigm.EDX
// application/vnd.novadigm.EXT
// application/vnd.osa.netdeploy
// application/vnd.palm
// application/vnd.pg.format
// application/vnd.pg.osasli
// application/vnd.powerbuilder6
// application/vnd.powerbuilder6-s
// application/vnd.powerbuilder7
// application/vnd.powerbuilder7-s
// application/vnd.powerbuilder75
// application/vnd.powerbuilder75-s
// application/vnd.previewsystems.box
// application/vnd.publishare-delta-tree
// application/vnd.pvi.ptid1
// application/vnd.pwg-xhtml-print+xml
// application/vnd.rapid
// application/vnd.rim.codcod
// application/vnd.s3sms
// application/vnd.seemail
// application/vnd.shana.informed.formdata
// application/vnd.shana.informed.formtemplate
// application/vnd.shana.informed.interchange
// application/vnd.shana.informed.package
// application/vnd.smafmmf
// application/vnd.sss-cod
// application/vnd.sss-dtf
// application/vnd.sss-ntf
// application/vnd.stardivision.calcsdc
// application/vnd.stardivision.chartsds
// application/vnd.stardivision.drawsda
// application/vnd.stardivision.impresssdd
// application/vnd.stardivision.mathsdf
// application/vnd.stardivision.writersdw
// application/vnd.stardivision.writer-globalsgl
// application/vnd.street-stream
// application/vnd.svd
// application/vnd.swiftview-ics
// application/vnd.symbian.installsis
// application/vnd.tcpdump.pcapcap pcap
// application/vnd.triscape.mxs
// application/vnd.trueapp
// application/vnd.truedoc
// application/vnd.tve-trigger
// application/vnd.ufdl
// application/vnd.uplanet.alert
// application/vnd.uplanet.alert-wbxml
// application/vnd.uplanet.bearer-choice
// application/vnd.uplanet.bearer-choice-wbxml
// application/vnd.uplanet.cacheop
// application/vnd.uplanet.cacheop-wbxml
// application/vnd.uplanet.channel
// application/vnd.uplanet.channel-wbxml
// application/vnd.uplanet.list
// application/vnd.uplanet.list-wbxml
// application/vnd.uplanet.listcmd
// application/vnd.uplanet.listcmd-wbxml
// application/vnd.uplanet.signal
// application/vnd.vcx
// application/vnd.vectorworks
// application/vnd.vidsoft.vidconference
// application/vnd.visiovsd vst vsw vss
// application/vnd.vividence.scriptfile
// application/vnd.wap.sic
// application/vnd.wap.slc
// application/vnd.wap.wbxmlwbxml
// application/vnd.wap.wmlcwmlc
// application/vnd.wap.wmlscriptcwmlsc
// application/vnd.webturbo
// application/vnd.wordperfectwpd
// application/vnd.wordperfect5.1wp5
// application/vnd.wrq-hp3000-labelled
// application/vnd.wt.stf
// application/vnd.xara
// application/vnd.xfdl
// application/vnd.yellowriver-custom-menu
// application/zlib
// application/x-123wk
// application/x-abiwordabw
// application/x-bcpiobcpio
// application/x-cabcab
// application/x-cbrcbr
// application/x-cbzcbz
// application/x-cdfcdf cda
// application/x-cdlinkvcd
// application/x-chess-pgnpgn
// application/x-comsolmph
// application/x-core
// application/x-cpiocpio
// application/x-directordcr dir dxr
// application/x-dmsdms
// application/x-doomwad
// application/x-dvidvi
// application/x-freemindmm
// application/x-futuresplashspl
// application/x-ganttprojectgan
// application/x-gnumericgnumeric
// application/x-go-sgfsgf
// application/x-graphing-calculatorgcf
// application/x-hdfhdf
// #application/x-httpd-erubyrhtml
// #application/x-httpd-phpphtml pht php
// #application/x-httpd-php-sourcephps
// #application/x-httpd-php3php3
// #application/x-httpd-php3-preprocessedphp3p
// #application/x-httpd-php4php4
// #application/x-httpd-php5php5
// application/x-hwphwp
// application/x-icaica
// application/x-infoinfo
// application/x-internet-signupins isp
// application/x-iphoneiii
// application/x-iso9660-imageiso
// application/x-jamjam
// application/x-java-applet
// application/x-java-bean
// application/x-java-jnlp-filejnlp
// application/x-jmoljmz
// application/x-kchartchrt
// application/x-kdelnk
// application/x-killustratorkil
// application/x-koanskp skd skt skm
// application/x-kpresenterkpr kpt
// application/x-kspreadksp
// application/x-kwordkwd kwt
// application/x-lhalha
// application/x-lyxlyx
// application/x-lzhlzh
// application/x-lzxlzx
// application/x-makerfrm maker frame fm fb book fbdoc
// application/x-mifmif
// application/x-mpegURLm3u8
// application/x-ms-applicationapplication
// application/x-ms-manifestmanifest
// application/x-ms-wmdwmd
// application/x-ms-wmzwmz
// application/x-msimsi
// application/x-netcdfnc
// application/x-ns-proxy-autoconfigpac
// application/x-nwcnwc
// application/x-oz-applicationoza
// application/x-pkcs7-certreqrespp7r
// application/x-pkcs7-crlcrl
// application/x-qgisqgs shp shx
// application/x-quicktimeplayerqtl
// application/x-rdprdp
// application/x-rx
// application/x-scilabsci sce
// application/x-scilab-xcosxcos
// application/x-shellscript
// application/x-shockwave-flashswf swfl
// application/x-silverlightscr
// application/x-sv4cpiosv4cpio
// application/x-sv4crcsv4crc

// application/x-tex-gfgf
// application/x-tex-pkpk
// application/x-ustarustar
// application/x-videolan
// application/x-wais-sourcesrc
// application/x-wingzwz
// application/x-x509-ca-certcrt
// application/x-xpinstallxpi

// audio/32kadpcm
// audio/3gpp
// audio/g.722.1
// audio/l16
// audio/mp4a-latm
// audio/mpa-robust
// audio/parityfec
// audio/telephone-event
// audio/tone
// audio/vnd.cisco.nse
// audio/vnd.cns.anp1
// audio/vnd.cns.inf1
// audio/vnd.digital-winds
// audio/vnd.everad.plj
// audio/vnd.lucent.voice
// audio/vnd.nortel.vbk
// audio/vnd.nuera.ecelp4800
// audio/vnd.nuera.ecelp7470
// audio/vnd.nuera.ecelp9600
// audio/vnd.octel.sbc
// audio/vnd.qcelp
// audio/vnd.rhetorex.32kadpcm
// audio/vnd.vmx.cvsd

// chemical/x-alchemyalc
// chemical/x-cachecac cache
// chemical/x-cache-csfcsf
// chemical/x-cactvs-binarycbin cascii ctab
// chemical/x-cdxcdx
// chemical/x-ceriuscer
// chemical/x-chem3dc3d
// chemical/x-chemdrawchm
// chemical/x-cifcif
// chemical/x-cmdfcmdf
// chemical/x-cmlcml
// chemical/x-compasscpa
// chemical/x-crossfirebsd
// chemical/x-csmlcsml csm
// chemical/x-ctxctx
// chemical/x-cxfcxf cef
// #chemical/x-daylight-smilessmi
// chemical/x-embl-dl-nucleotideemb embl
// chemical/x-galactic-spcspc
// chemical/x-gamess-inputinp gam gamin
// chemical/x-gaussian-checkpointfch fchk
// chemical/x-gaussian-cubecub
// chemical/x-gaussian-inputgau gjc gjf
// chemical/x-gaussian-loggal
// chemical/x-gcg8-sequencegcg
// chemical/x-genbankgen
// chemical/x-hinhin
// chemical/x-isostaristr ist
// chemical/x-jcamp-dxjdx dx
// chemical/x-kinemagekin
// chemical/x-macmoleculemcm
// chemical/x-macromodel-inputmmd mmod
// chemical/x-mdl-molfilemol
// chemical/x-mdl-rdfilerd
// chemical/x-mdl-rxnfilerxn
// chemical/x-mdl-sdfilesd sdf
// chemical/x-mdl-tgftgf
// #chemical/x-mifmif
// chemical/x-mmcifmcif
// chemical/x-mol2mol2
// chemical/x-molconn-Zb
// chemical/x-mopac-graphgpt
// chemical/x-mopac-inputmop mopcrt mpc zmt
// chemical/x-mopac-outmoo
// chemical/x-mopac-vibmvb
// chemical/x-ncbi-asn1asn
// chemical/x-ncbi-asn1-asciiprt ent
// chemical/x-ncbi-asn1-binaryval aso
// chemical/x-ncbi-asn1-specasn
// chemical/x-pdbpdb ent
// chemical/x-rosdalros
// chemical/x-swissprotsw
// chemical/x-vamas-iso14976vms
// chemical/x-vmdvmd
// chemical/x-xtelxtel
// chemical/x-xyzxyz

// image/cgm
// image/g3fax
// image/naplps
// image/prs.btif
// image/prs.pti
// image/vnd.cns.inf2
// image/vnd.dwg
// image/vnd.dxf
// image/vnd.fastbidsheet
// image/vnd.fpx
// image/vnd.fst
// image/vnd.fujixerox.edmics-mmr
// image/vnd.fujixerox.edmics-rlc
// image/vnd.mix
// image/vnd.net-fpx
// image/vnd.svf
// image/vnd.xiff
// image/x-icon

// inode/chardevice
// inode/blockdevice
// inode/directory-locked
// inode/directory
// inode/fifo
// inode/socket

// message/delivery-status
// message/disposition-notification
// message/external-body
// message/http
// message/s-http
// message/news
// message/partial
// message/rfc822eml

// model/vnd.dwf
// model/vnd.flatland.3dml
// model/vnd.gdl
// model/vnd.gs-gdl
// model/vnd.gtw
// model/vnd.mts
// model/vnd.vtu

// multipart/alternative
// multipart/appledouble
// multipart/byteranges
// multipart/digest
// multipart/encrypted
// multipart/form-data
// multipart/header-set
// multipart/mixed
// multipart/parallel
// multipart/related
// multipart/report
// multipart/signed
// multipart/voice-message

// text/english
// text/enriched
// {"text/x-gap",
// {"text/x-gtkrc",
// text/h323323
// text/iulsuls
//{"text/x-idl",
//{"text/x-netrexx",
//{"text/x-ocl",
//{"text/x-dtd",
// {"text/x-gettext-translation",
// {"text/x-gettext-translation-template",
// text/parityfec
// text/prs.lines.tag
// text/rfc822-headers
// text/scriptletsct wsc
// text/t140
// text/texmacstm
// text/turtlettl
// text/vnd.abc
// text/vnd.curl
// text/vnd.debian.copyright
// text/vnd.DMClientScript
// text/vnd.flatland.3dml
// text/vnd.fly
// text/vnd.fmi.flexstor
// text/vnd.in3d.3dml
// text/vnd.in3d.spot
// text/vnd.IPTC.NewsML
// text/vnd.IPTC.NITF
// text/vnd.latex-z
// text/vnd.motorola.reflex
// text/vnd.ms-mediapackage
// text/vnd.sun.j2me.app-descriptorjad
// text/vnd.wap.si
// text/vnd.wap.sl
// text/vnd.wap.wmlwml
// text/vnd.wap.wmlscriptwmls
// text/x-booboo
// text/x-componenthtc
// text/x-crontab
// text/x-lilypondly
// text/x-pcs-gcdgcd
// text/x-setextetx
// text/x-sfvsfv

// video/mp4v-es
// video/parityfec
// video/pointer
// video/vnd.fvt
// video/vnd.motorola.video
// video/vnd.motorola.videop
// video/vnd.mts
// video/vnd.nokia.interleaved-multimedia
// video/vnd.vivo

// x-conference/x-cooltalkice
//
// x-epoc/x-sisx-appsisx
