// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package svg

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unsafe"

	"cogentcore.org/core/base/slicesx"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint"
	"cogentcore.org/core/tree"
)

// Path renders SVG data sequences that can render just about anything
type Path struct {
	NodeBase

	// the path data to render -- path commands and numbers are serialized, with each command specifying the number of floating-point coord data points that follow
	Data []PathData `xml:"-" set:"-"`

	// string version of the path data
	DataStr string `xml:"d"`
}

func (g *Path) SVGName() string { return "path" }

func (g *Path) SetPos(pos math32.Vector2) {
	// todo: set first point
}

func (g *Path) SetSize(sz math32.Vector2) {
	// todo: scale bbox
}

// SetData sets the path data to given string, parsing it into an optimized
// form used for rendering
func (g *Path) SetData(data string) error {
	g.DataStr = data
	var err error
	g.Data, err = PathDataParse(data)
	if err != nil {
		return err
	}
	err = PathDataValidate(&g.Data, g.Path())
	return err
}

func (g *Path) LocalBBox() math32.Box2 {
	bb := PathDataBBox(g.Data)
	hlw := 0.5 * g.LocalLineWidth()
	bb.Min.SetSubScalar(hlw)
	bb.Max.SetAddScalar(hlw)
	return bb
}

func (g *Path) Render(sv *SVG) {
	sz := len(g.Data)
	if sz < 2 {
		return
	}
	vis, pc := g.PushTransform(sv)
	if !vis {
		return
	}
	PathDataRender(g.Data, pc)
	pc.FillStrokeClear()

	g.BBoxes(sv)

	if mrk := sv.MarkerByName(g, "marker-start"); mrk != nil {
		// todo: could look for close-path at end and find angle from there..
		stv, ang := PathDataStart(g.Data)
		mrk.RenderMarker(sv, stv, ang, g.Paint.StrokeStyle.Width.Dots)
	}
	if mrk := sv.MarkerByName(g, "marker-end"); mrk != nil {
		env, ang := PathDataEnd(g.Data)
		mrk.RenderMarker(sv, env, ang, g.Paint.StrokeStyle.Width.Dots)
	}
	if mrk := sv.MarkerByName(g, "marker-mid"); mrk != nil {
		var ptm2, ptm1, pt math32.Vector2
		gotidx := 0
		PathDataIterFunc(g.Data, func(idx int, cmd PathCmds, ptIndex int, cp math32.Vector2, ctrls []math32.Vector2) bool {
			ptm2 = ptm1
			ptm1 = pt
			pt = cp
			if gotidx < 2 {
				gotidx++
				return true
			}
			if idx >= sz-3 { // todo: this is approximate...
				return false
			}
			ang := 0.5 * (math32.Atan2(pt.Y-ptm1.Y, pt.X-ptm1.X) + math32.Atan2(ptm1.Y-ptm2.Y, ptm1.X-ptm2.X))
			mrk.RenderMarker(sv, ptm1, ang, g.Paint.StrokeStyle.Width.Dots)
			gotidx++
			return true
		})
	}

	g.RenderChildren(sv)
	pc.PopTransform()
}

// AddPath adds given path command to the PathData
func (g *Path) AddPath(cmd PathCmds, args ...float32) {
	na := len(args)
	cd := cmd.EncCmd(na)
	g.Data = append(g.Data, cd)
	if na > 0 {
		ad := unsafe.Slice((*PathData)(unsafe.Pointer(&args[0])), na)
		g.Data = append(g.Data, ad...)
	}
}

// AddPathArc adds an arc command using the simpler Paint.DrawArc parameters
// with center at the current position, and the given radius
// and angles in degrees. Because the y axis points down, angles are clockwise,
// and the rendering draws segments progressing from angle1 to angle2.
func (g *Path) AddPathArc(r, angle1, angle2 float32) {
	ra1 := math32.DegToRad(angle1)
	ra2 := math32.DegToRad(angle2)
	xs := r * math32.Cos(ra1)
	ys := r * math32.Sin(ra1)
	xe := r * math32.Cos(ra2)
	ye := r * math32.Sin(ra2)
	longArc := float32(0)
	if math32.Abs(angle2-angle1) >= 180 {
		longArc = 1
	}
	sweep := float32(1)
	if angle2-angle1 < 0 {
		sweep = 0
	}
	g.AddPath(Pcm, xs, ys)
	g.AddPath(Pca, r, r, 0, longArc, sweep, xe-xs, ye-ys)
}

// UpdatePathString sets the path string from the Data
func (g *Path) UpdatePathString() {
	g.DataStr = PathDataString(g.Data)
}

// PathCmds are the commands within the path SVG drawing data type
type PathCmds byte //enum: enum

const (
	// move pen, abs coords
	PcM PathCmds = iota
	// move pen, rel coords
	Pcm
	// lineto, abs
	PcL
	// lineto, rel
	Pcl
	// horizontal lineto, abs
	PcH
	// relative lineto, rel
	Pch
	// vertical lineto, abs
	PcV
	// vertical lineto, rel
	Pcv
	// Bezier curveto, abs
	PcC
	// Bezier curveto, rel
	Pcc
	// smooth Bezier curveto, abs
	PcS
	// smooth Bezier curveto, rel
	Pcs
	// quadratic Bezier curveto, abs
	PcQ
	// quadratic Bezier curveto, rel
	Pcq
	// smooth quadratic Bezier curveto, abs
	PcT
	// smooth quadratic Bezier curveto, rel
	Pct
	// elliptical arc, abs
	PcA
	// elliptical arc, rel
	Pca
	// close path
	PcZ
	// close path
	Pcz
	// error -- invalid command
	PcErr
)

// PathData encodes the svg path data, using 32-bit floats which are converted
// into uint32 for path commands, and contain the command as the first 5
// bits, and the remaining 27 bits are the number of data points following the
// path command to interpret as numbers.
type PathData float32

// Cmd decodes path data as a command and a number of subsequent values for that command
func (pd PathData) Cmd() (PathCmds, int) {
	iv := uint32(pd)
	cmd := PathCmds(iv & 0x1F)       // only the lowest 5 bits (31 values) for command
	n := int((iv & 0xFFFFFFE0) >> 5) // extract the n from remainder of bits
	return cmd, n
}

// EncCmd encodes command and n into PathData
func (pc PathCmds) EncCmd(n int) PathData {
	nb := int32(n << 5) // n up-shifted
	pd := PathData(int32(pc) | nb)
	return pd
}

// PathDataNext gets the next path data point, incrementing the index
func PathDataNext(data []PathData, i *int) float32 {
	pd := data[*i]
	(*i)++
	return float32(pd)
}

// PathDataNextVector gets the next 2 path data points as a vector
func PathDataNextVector(data []PathData, i *int) math32.Vector2 {
	v := math32.Vector2{}
	v.X = float32(data[*i])
	(*i)++
	v.Y = float32(data[*i])
	(*i)++
	return v
}

// PathDataNextRel gets the next 2 path data points as a relative vector
// and returns that relative vector added to current point
func PathDataNextRel(data []PathData, i *int, cp math32.Vector2) math32.Vector2 {
	v := math32.Vector2{}
	v.X = float32(data[*i])
	(*i)++
	v.Y = float32(data[*i])
	(*i)++
	return v.Add(cp)
}

// PathDataNextCmd gets the next path data command, incrementing the index -- ++
// not an expression so its clunky
func PathDataNextCmd(data []PathData, i *int) (PathCmds, int) {
	pd := data[*i]
	(*i)++
	return pd.Cmd()
}

func reflectPt(pt, rp math32.Vector2) math32.Vector2 {
	return pt.MulScalar(2).Sub(rp)
}

// PathDataRender traverses the path data and renders it using paint.
// We assume all the data has been validated and that n's are sufficient, etc
func PathDataRender(data []PathData, pc *paint.Context) {
	sz := len(data)
	if sz == 0 {
		return
	}
	lastCmd := PcErr
	var st, cp, xp, ctrl math32.Vector2
	for i := 0; i < sz; {
		cmd, n := PathDataNextCmd(data, &i)
		rel := false
		switch cmd {
		case PcM:
			cp = PathDataNextVector(data, &i)
			pc.MoveTo(cp.X, cp.Y)
			st = cp
			for np := 1; np < n/2; np++ {
				cp = PathDataNextVector(data, &i)
				pc.LineTo(cp.X, cp.Y)
			}
		case Pcm:
			cp = PathDataNextRel(data, &i, cp)
			pc.MoveTo(cp.X, cp.Y)
			st = cp
			for np := 1; np < n/2; np++ {
				cp = PathDataNextRel(data, &i, cp)
				pc.LineTo(cp.X, cp.Y)
			}
		case PcL:
			for np := 0; np < n/2; np++ {
				cp = PathDataNextVector(data, &i)
				pc.LineTo(cp.X, cp.Y)
			}
		case Pcl:
			for np := 0; np < n/2; np++ {
				cp = PathDataNextRel(data, &i, cp)
				pc.LineTo(cp.X, cp.Y)
			}
		case PcH:
			for np := 0; np < n; np++ {
				cp.X = PathDataNext(data, &i)
				pc.LineTo(cp.X, cp.Y)
			}
		case Pch:
			for np := 0; np < n; np++ {
				cp.X += PathDataNext(data, &i)
				pc.LineTo(cp.X, cp.Y)
			}
		case PcV:
			for np := 0; np < n; np++ {
				cp.Y = PathDataNext(data, &i)
				pc.LineTo(cp.X, cp.Y)
			}
		case Pcv:
			for np := 0; np < n; np++ {
				cp.Y += PathDataNext(data, &i)
				pc.LineTo(cp.X, cp.Y)
			}
		case PcC:
			for np := 0; np < n/6; np++ {
				xp = PathDataNextVector(data, &i)
				ctrl = PathDataNextVector(data, &i)
				cp = PathDataNextVector(data, &i)
				pc.CubicTo(xp.X, xp.Y, ctrl.X, ctrl.Y, cp.X, cp.Y)
			}
		case Pcc:
			for np := 0; np < n/6; np++ {
				xp = PathDataNextRel(data, &i, cp)
				ctrl = PathDataNextRel(data, &i, cp)
				cp = PathDataNextRel(data, &i, cp)
				pc.CubicTo(xp.X, xp.Y, ctrl.X, ctrl.Y, cp.X, cp.Y)
			}
		case Pcs:
			rel = true
			fallthrough
		case PcS:
			for np := 0; np < n/4; np++ {
				switch lastCmd {
				case Pcc, PcC, Pcs, PcS:
					ctrl = reflectPt(cp, ctrl)
				default:
					ctrl = cp
				}
				if rel {
					xp = PathDataNextRel(data, &i, cp)
					cp = PathDataNextRel(data, &i, cp)
				} else {
					xp = PathDataNextVector(data, &i)
					cp = PathDataNextVector(data, &i)
				}
				pc.CubicTo(ctrl.X, ctrl.Y, xp.X, xp.Y, cp.X, cp.Y)
				lastCmd = cmd
				ctrl = xp
			}
		case PcQ:
			for np := 0; np < n/4; np++ {
				ctrl = PathDataNextVector(data, &i)
				cp = PathDataNextVector(data, &i)
				pc.QuadraticTo(ctrl.X, ctrl.Y, cp.X, cp.Y)
			}
		case Pcq:
			for np := 0; np < n/4; np++ {
				ctrl = PathDataNextRel(data, &i, cp)
				cp = PathDataNextRel(data, &i, cp)
				pc.QuadraticTo(ctrl.X, ctrl.Y, cp.X, cp.Y)
			}
		case Pct:
			rel = true
			fallthrough
		case PcT:
			for np := 0; np < n/2; np++ {
				switch lastCmd {
				case Pcq, PcQ, PcT, Pct:
					ctrl = reflectPt(cp, ctrl)
				default:
					ctrl = cp
				}
				if rel {
					cp = PathDataNextRel(data, &i, cp)
				} else {
					cp = PathDataNextVector(data, &i)
				}
				pc.QuadraticTo(ctrl.X, ctrl.Y, cp.X, cp.Y)
				lastCmd = cmd
			}
		case Pca:
			rel = true
			fallthrough
		case PcA:
			for np := 0; np < n/7; np++ {
				rad := PathDataNextVector(data, &i)
				ang := PathDataNext(data, &i)
				largeArc := (PathDataNext(data, &i) != 0)
				sweep := (PathDataNext(data, &i) != 0)
				prv := cp
				if rel {
					cp = PathDataNextRel(data, &i, cp)
				} else {
					cp = PathDataNextVector(data, &i)
				}
				ncx, ncy := paint.FindEllipseCenter(&rad.X, &rad.Y, ang*math.Pi/180, prv.X, prv.Y, cp.X, cp.Y, sweep, largeArc)
				cp.X, cp.Y = pc.DrawEllipticalArcPath(ncx, ncy, cp.X, cp.Y, prv.X, prv.Y, rad.X, rad.Y, ang, largeArc, sweep)
			}
		case PcZ:
			fallthrough
		case Pcz:
			pc.ClosePath()
			cp = st
		}
		lastCmd = cmd
	}
}

// PathDataIterFunc traverses the path data and calls given function on each
// coordinate point, passing overall starting index of coords in data stream,
// command, index of the points within that command, and coord values
// (absolute, not relative, regardless of the command type), including
// special control points for path commands that have them (else nil).
// If function returns false (use [tree.Break] vs. [tree.Continue]) then
// traversal is aborted.
// For Control points, order is in same order as in standard path stream
// when multiple, e.g., C,S.
// For A: order is: nc, prv, rad, math32.Vector2{X: ang}, math32.Vec2(laf, sf)}
func PathDataIterFunc(data []PathData, fun func(idx int, cmd PathCmds, ptIndex int, cp math32.Vector2, ctrls []math32.Vector2) bool) {
	sz := len(data)
	if sz == 0 {
		return
	}
	lastCmd := PcErr
	var st, cp, xp, ctrl, nc math32.Vector2
	for i := 0; i < sz; {
		cmd, n := PathDataNextCmd(data, &i)
		rel := false
		switch cmd {
		case PcM:
			cp = PathDataNextVector(data, &i)
			if !fun(i-2, cmd, 0, cp, nil) {
				return
			}
			st = cp
			for np := 1; np < n/2; np++ {
				cp = PathDataNextVector(data, &i)
				if !fun(i-2, cmd, np, cp, nil) {
					return
				}
			}
		case Pcm:
			cp = PathDataNextRel(data, &i, cp)
			if !fun(i-2, cmd, 0, cp, nil) {
				return
			}
			st = cp
			for np := 1; np < n/2; np++ {
				cp = PathDataNextRel(data, &i, cp)
				if !fun(i-2, cmd, np, cp, nil) {
					return
				}
			}
		case PcL:
			for np := 0; np < n/2; np++ {
				cp = PathDataNextVector(data, &i)
				if !fun(i-2, cmd, np, cp, nil) {
					return
				}
			}
		case Pcl:
			for np := 0; np < n/2; np++ {
				cp = PathDataNextRel(data, &i, cp)
				if !fun(i-2, cmd, np, cp, nil) {
					return
				}
			}
		case PcH:
			for np := 0; np < n; np++ {
				cp.X = PathDataNext(data, &i)
				if !fun(i-1, cmd, np, cp, nil) {
					return
				}
			}
		case Pch:
			for np := 0; np < n; np++ {
				cp.X += PathDataNext(data, &i)
				if !fun(i-1, cmd, np, cp, nil) {
					return
				}
			}
		case PcV:
			for np := 0; np < n; np++ {
				cp.Y = PathDataNext(data, &i)
				if !fun(i-1, cmd, np, cp, nil) {
					return
				}
			}
		case Pcv:
			for np := 0; np < n; np++ {
				cp.Y += PathDataNext(data, &i)
				if !fun(i-1, cmd, np, cp, nil) {
					return
				}
			}
		case PcC:
			for np := 0; np < n/6; np++ {
				xp = PathDataNextVector(data, &i)
				ctrl = PathDataNextVector(data, &i)
				cp = PathDataNextVector(data, &i)
				if !fun(i-2, cmd, np, cp, []math32.Vector2{xp, ctrl}) {
					return
				}
			}
		case Pcc:
			for np := 0; np < n/6; np++ {
				xp = PathDataNextRel(data, &i, cp)
				ctrl = PathDataNextRel(data, &i, cp)
				cp = PathDataNextRel(data, &i, cp)
				if !fun(i-2, cmd, np, cp, []math32.Vector2{xp, ctrl}) {
					return
				}
			}
		case Pcs:
			rel = true
			fallthrough
		case PcS:
			for np := 0; np < n/4; np++ {
				switch lastCmd {
				case Pcc, PcC, Pcs, PcS:
					ctrl = reflectPt(cp, ctrl)
				default:
					ctrl = cp
				}
				if rel {
					xp = PathDataNextRel(data, &i, cp)
					cp = PathDataNextRel(data, &i, cp)
				} else {
					xp = PathDataNextVector(data, &i)
					cp = PathDataNextVector(data, &i)
				}
				if !fun(i-2, cmd, np, cp, []math32.Vector2{xp, ctrl}) {
					return
				}
				lastCmd = cmd
				ctrl = xp
			}
		case PcQ:
			for np := 0; np < n/4; np++ {
				ctrl = PathDataNextVector(data, &i)
				cp = PathDataNextVector(data, &i)
				if !fun(i-2, cmd, np, cp, []math32.Vector2{ctrl}) {
					return
				}
			}
		case Pcq:
			for np := 0; np < n/4; np++ {
				ctrl = PathDataNextRel(data, &i, cp)
				cp = PathDataNextRel(data, &i, cp)
				if !fun(i-2, cmd, np, cp, []math32.Vector2{ctrl}) {
					return
				}
			}
		case Pct:
			rel = true
			fallthrough
		case PcT:
			for np := 0; np < n/2; np++ {
				switch lastCmd {
				case Pcq, PcQ, PcT, Pct:
					ctrl = reflectPt(cp, ctrl)
				default:
					ctrl = cp
				}
				if rel {
					cp = PathDataNextRel(data, &i, cp)
				} else {
					cp = PathDataNextVector(data, &i)
				}
				if !fun(i-2, cmd, np, cp, []math32.Vector2{ctrl}) {
					return
				}
				lastCmd = cmd
			}
		case Pca:
			rel = true
			fallthrough
		case PcA:
			for np := 0; np < n/7; np++ {
				rad := PathDataNextVector(data, &i)
				ang := PathDataNext(data, &i)
				laf := PathDataNext(data, &i)
				largeArc := (laf != 0)
				sf := PathDataNext(data, &i)
				sweep := (sf != 0)

				prv := cp
				if rel {
					cp = PathDataNextRel(data, &i, cp)
				} else {
					cp = PathDataNextVector(data, &i)
				}
				nc.X, nc.Y = paint.FindEllipseCenter(&rad.X, &rad.Y, ang*math.Pi/180, prv.X, prv.Y, cp.X, cp.Y, sweep, largeArc)
				if !fun(i-2, cmd, np, cp, []math32.Vector2{nc, prv, rad, {X: ang}, {laf, sf}}) {
					return
				}
			}
		case PcZ:
			fallthrough
		case Pcz:
			cp = st
		}
		lastCmd = cmd
	}
}

// PathDataBBox traverses the path data and extracts the local bounding box
func PathDataBBox(data []PathData) math32.Box2 {
	bb := math32.B2Empty()
	PathDataIterFunc(data, func(idx int, cmd PathCmds, ptIndex int, cp math32.Vector2, ctrls []math32.Vector2) bool {
		bb.ExpandByPoint(cp)
		return tree.Continue
	})
	return bb
}

// PathDataStart gets the starting coords and angle from the path
func PathDataStart(data []PathData) (vec math32.Vector2, ang float32) {
	gotSt := false
	PathDataIterFunc(data, func(idx int, cmd PathCmds, ptIndex int, cp math32.Vector2, ctrls []math32.Vector2) bool {
		if gotSt {
			ang = math32.Atan2(cp.Y-vec.Y, cp.X-vec.X)
			return tree.Break
		}
		vec = cp
		gotSt = true
		return tree.Continue
	})
	return
}

// PathDataEnd gets the ending coords and angle from the path
func PathDataEnd(data []PathData) (vec math32.Vector2, ang float32) {
	gotSome := false
	PathDataIterFunc(data, func(idx int, cmd PathCmds, ptIndex int, cp math32.Vector2, ctrls []math32.Vector2) bool {
		if gotSome {
			ang = math32.Atan2(cp.Y-vec.Y, cp.X-vec.X)
		}
		vec = cp
		gotSome = true
		return tree.Continue
	})
	return
}

// PathCmdNMap gives the number of points per each command
var PathCmdNMap = map[PathCmds]int{
	PcM: 2,
	Pcm: 2,
	PcL: 2,
	Pcl: 2,
	PcH: 1,
	Pch: 1,
	PcV: 1,
	Pcv: 1,
	PcC: 6,
	Pcc: 6,
	PcS: 4,
	Pcs: 4,
	PcQ: 4,
	Pcq: 4,
	PcT: 2,
	Pct: 2,
	PcA: 7,
	Pca: 7,
	PcZ: 0,
	Pcz: 0,
}

// PathCmdIsRel returns true if the path command is relative, false for absolute
func PathCmdIsRel(pc PathCmds) bool {
	return pc%2 == 1 // odd ones are relative
}

// PathDataValidate validates the path data and emits error messages on log
func PathDataValidate(data *[]PathData, errstr string) error {
	sz := len(*data)
	if sz == 0 {
		return nil
	}

	di := 0
	fcmd, _ := PathDataNextCmd(*data, &di)
	if !(fcmd == Pcm || fcmd == PcM) {
		log.Printf("core.PathDataValidate on %v: doesn't start with M or m -- adding\n", errstr)
		ns := make([]PathData, 3, sz+3)
		ns[0] = PcM.EncCmd(2)
		ns[1], ns[2] = (*data)[1], (*data)[2]
		*data = append(ns, *data...)
	}
	sz = len(*data)

	for i := 0; i < sz; {
		cmd, n := PathDataNextCmd(*data, &i)
		trgn, ok := PathCmdNMap[cmd]
		if !ok {
			err := fmt.Errorf("core.PathDataValidate on %v: Path Command not valid: %v", errstr, cmd)
			log.Println(err)
			return err
		}
		if (trgn == 0 && n > 0) || (trgn > 0 && n%trgn != 0) {
			err := fmt.Errorf("core.PathDataValidate on %v: Path Command %v has invalid n: %v -- should be: %v", errstr, cmd, n, trgn)
			log.Println(err)
			return err
		}
		for np := 0; np < n; np++ {
			PathDataNext(*data, &i)
		}
	}
	return nil
}

// PathRuneToCmd maps rune to path command
var PathRuneToCmd = map[rune]PathCmds{
	'M': PcM,
	'm': Pcm,
	'L': PcL,
	'l': Pcl,
	'H': PcH,
	'h': Pch,
	'V': PcV,
	'v': Pcv,
	'C': PcC,
	'c': Pcc,
	'S': PcS,
	's': Pcs,
	'Q': PcQ,
	'q': Pcq,
	'T': PcT,
	't': Pct,
	'A': PcA,
	'a': Pca,
	'Z': PcZ,
	'z': Pcz,
}

// PathCmdToRune maps command to rune
var PathCmdToRune = map[PathCmds]rune{}

func init() {
	for k, v := range PathRuneToCmd {
		PathCmdToRune[v] = k
	}
}

// PathDecodeCmd decodes rune into corresponding command
func PathDecodeCmd(r rune) PathCmds {
	cmd, ok := PathRuneToCmd[r]
	if ok {
		return cmd
	} else {
		// log.Printf("core.PathDecodeCmd unrecognized path command: %v %v\n", string(r), r)
		return PcErr
	}
}

// PathDataParse parses a string representation of the path data into compiled path data
func PathDataParse(d string) ([]PathData, error) {
	var pd []PathData
	endi := len(d) - 1
	numSt := -1
	numGotDec := false // did last number already get a decimal point -- if so, then an additional decimal point now acts as a delimiter -- some crazy paths actually leverage that!
	lr := ' '
	lstCmd := -1
	// first pass: just do the raw parse into commands and numbers
	for i, r := range d {
		num := unicode.IsNumber(r) || (r == '.' && !numGotDec) || (r == '-' && lr == 'e') || r == 'e'
		notn := !num
		if i == endi || notn {
			if numSt != -1 || (i == endi && !notn) {
				if numSt == -1 {
					numSt = i
				}
				nstr := d[numSt:i]
				if i == endi && !notn {
					nstr = d[numSt : i+1]
				}
				p, err := strconv.ParseFloat(nstr, 32)
				if err != nil {
					log.Printf("core.PathDataParse could not parse string: %v into float\n", nstr)
					return nil, err
				}
				pd = append(pd, PathData(p))
			}
			if r == '-' || r == '.' {
				numSt = i
				if r == '.' {
					numGotDec = true
				} else {
					numGotDec = false
				}
			} else {
				numSt = -1
				numGotDec = false
				if lstCmd != -1 { // update number of args for previous command
					lcm, _ := pd[lstCmd].Cmd()
					n := (len(pd) - lstCmd) - 1
					pd[lstCmd] = lcm.EncCmd(n)
				}
				if !unicode.IsSpace(r) && r != ',' {
					cmd := PathDecodeCmd(r)
					if cmd == PcErr {
						if i != endi {
							err := fmt.Errorf("core.PathDataParse invalid command rune: %v", r)
							log.Println(err)
							return nil, err
						}
					} else {
						pc := cmd.EncCmd(0) // encode with 0 length to start
						lstCmd = len(pd)
						pd = append(pd, pc) // push on
					}
				}
			}
		} else if numSt == -1 { // got start of a number
			numSt = i
			if r == '.' {
				numGotDec = true
			} else {
				numGotDec = false
			}
		} else { // inside a number
			if r == '.' {
				numGotDec = true
			}
		}
		lr = r
	}
	return pd, nil
	// todo: add some error checking..
}

// PathDataString returns the string representation of the path data
func PathDataString(data []PathData) string {
	sz := len(data)
	if sz == 0 {
		return ""
	}
	var sb strings.Builder
	var rp, cp, xp, ctrl math32.Vector2
	for i := 0; i < sz; {
		cmd, n := PathDataNextCmd(data, &i)
		sb.WriteString(fmt.Sprintf("%c ", PathCmdToRune[cmd]))
		switch cmd {
		case PcM, Pcm:
			cp = PathDataNextVector(data, &i)
			sb.WriteString(fmt.Sprintf("%g,%g ", cp.X, cp.Y))
			for np := 1; np < n/2; np++ {
				cp = PathDataNextVector(data, &i)
				sb.WriteString(fmt.Sprintf("%g,%g ", cp.X, cp.Y))
			}
		case PcL, Pcl:
			for np := 0; np < n/2; np++ {
				rp = PathDataNextVector(data, &i)
				sb.WriteString(fmt.Sprintf("%g,%g ", rp.X, rp.Y))
			}
		case PcH, Pch, PcV, Pcv:
			for np := 0; np < n; np++ {
				cp.Y = PathDataNext(data, &i)
				sb.WriteString(fmt.Sprintf("%g ", cp.Y))
			}
		case PcC, Pcc:
			for np := 0; np < n/6; np++ {
				xp = PathDataNextVector(data, &i)
				sb.WriteString(fmt.Sprintf("%g,%g ", xp.X, xp.Y))
				ctrl = PathDataNextVector(data, &i)
				sb.WriteString(fmt.Sprintf("%g,%g ", ctrl.X, ctrl.Y))
				cp = PathDataNextVector(data, &i)
				sb.WriteString(fmt.Sprintf("%g,%g ", cp.X, cp.Y))
			}
		case Pcs, PcS:
			for np := 0; np < n/4; np++ {
				xp = PathDataNextVector(data, &i)
				sb.WriteString(fmt.Sprintf("%g,%g ", xp.X, xp.Y))
				cp = PathDataNextVector(data, &i)
				sb.WriteString(fmt.Sprintf("%g,%g ", cp.X, cp.Y))
			}
		case PcQ, Pcq:
			for np := 0; np < n/4; np++ {
				ctrl = PathDataNextVector(data, &i)
				sb.WriteString(fmt.Sprintf("%g,%g ", ctrl.X, ctrl.Y))
				cp = PathDataNextVector(data, &i)
				sb.WriteString(fmt.Sprintf("%g,%g ", cp.X, cp.Y))
			}
		case PcT, Pct:
			for np := 0; np < n/2; np++ {
				cp = PathDataNextVector(data, &i)
				sb.WriteString(fmt.Sprintf("%g,%g ", cp.X, cp.Y))
			}
		case PcA, Pca:
			for np := 0; np < n/7; np++ {
				rad := PathDataNextVector(data, &i)
				sb.WriteString(fmt.Sprintf("%g,%g ", rad.X, rad.Y))
				ang := PathDataNext(data, &i)
				largeArc := PathDataNext(data, &i)
				sweep := PathDataNext(data, &i)
				sb.WriteString(fmt.Sprintf("%g %g %g ", ang, largeArc, sweep))
				cp = PathDataNextVector(data, &i)
				sb.WriteString(fmt.Sprintf("%g,%g ", cp.X, cp.Y))
			}
		case PcZ, Pcz:
		}
	}
	return sb.String()
}

//////////////////////////////////////////////////////////////////////////////////
//  Transforms

// ApplyTransform applies the given 2D transform to the geometry of this node
// each node must define this for itself
func (g *Path) ApplyTransform(sv *SVG, xf math32.Matrix2) {
	// path may have horiz, vert elements -- only gen soln is to transform
	g.Paint.Transform.SetMul(xf)
	g.SetProperty("transform", g.Paint.Transform.String())
}

// PathDataTransformAbs does the transform of next two data points as absolute coords
func PathDataTransformAbs(data []PathData, i *int, xf math32.Matrix2, lpt math32.Vector2) math32.Vector2 {
	cp := PathDataNextVector(data, i)
	tc := xf.MulVector2AsPointCenter(cp, lpt)
	data[*i-2] = PathData(tc.X)
	data[*i-1] = PathData(tc.Y)
	return tc
}

// PathDataTransformRel does the transform of next two data points as relative coords
// compared to given cp coordinate.  returns new *absolute* coordinate
func PathDataTransformRel(data []PathData, i *int, xf math32.Matrix2, cp math32.Vector2) math32.Vector2 {
	rp := PathDataNextVector(data, i)
	tc := xf.MulVector2AsVector(rp)
	data[*i-2] = PathData(tc.X)
	data[*i-1] = PathData(tc.Y)
	return cp.Add(tc) // new abs
}

// ApplyDeltaTransform applies the given 2D delta transforms to the geometry of this node
// relative to given point.  Trans translation and point are in top-level coordinates,
// so must be transformed into local coords first.
// Point is upper left corner of selection box that anchors the translation and scaling,
// and for rotation it is the center point around which to rotate
func (g *Path) ApplyDeltaTransform(sv *SVG, trans math32.Vector2, scale math32.Vector2, rot float32, pt math32.Vector2) {
	crot := g.Paint.Transform.ExtractRot()
	if rot != 0 || crot != 0 {
		xf, lpt := g.DeltaTransform(trans, scale, rot, pt, false) // exclude self
		g.Paint.Transform.SetMulCenter(xf, lpt)
		g.SetProperty("transform", g.Paint.Transform.String())
	} else {
		xf, lpt := g.DeltaTransform(trans, scale, rot, pt, true) // include self
		g.ApplyTransformImpl(xf, lpt)
		g.GradientApplyTransformPt(sv, xf, lpt)
	}
}

// ApplyTransformImpl does the implementation of applying a transform to all points
func (g *Path) ApplyTransformImpl(xf math32.Matrix2, lpt math32.Vector2) {
	sz := len(g.Data)
	data := g.Data
	lastCmd := PcErr
	var cp, st math32.Vector2
	var xp, ctrl, rp math32.Vector2
	for i := 0; i < sz; {
		cmd, n := PathDataNextCmd(data, &i)
		rel := false
		switch cmd {
		case PcM:
			cp = PathDataTransformAbs(data, &i, xf, lpt)
			st = cp
			for np := 1; np < n/2; np++ {
				cp = PathDataTransformAbs(data, &i, xf, lpt)
			}
		case Pcm:
			if i == 1 { // starting
				cp = PathDataTransformAbs(data, &i, xf, lpt)
			} else {
				cp = PathDataTransformRel(data, &i, xf, cp)
			}
			st = cp
			for np := 1; np < n/2; np++ {
				cp = PathDataTransformRel(data, &i, xf, cp)
			}
		case PcL:
			for np := 0; np < n/2; np++ {
				cp = PathDataTransformAbs(data, &i, xf, lpt)
			}
		case Pcl:
			for np := 0; np < n/2; np++ {
				cp = PathDataTransformRel(data, &i, xf, cp)
			}
		case PcH:
			for np := 0; np < n; np++ {
				cp.X = PathDataNext(data, &i)
				tc := xf.MulVector2AsPointCenter(cp, lpt)
				data[i-1] = PathData(tc.X)
			}
		case Pch:
			for np := 0; np < n; np++ {
				rp.X = PathDataNext(data, &i)
				rp.Y = 0
				rp = xf.MulVector2AsVector(rp)
				data[i-1] = PathData(rp.X)
				cp.SetAdd(rp) // new abs
			}
		case PcV:
			for np := 0; np < n; np++ {
				cp.Y = PathDataNext(data, &i)
				tc := xf.MulVector2AsPointCenter(cp, lpt)
				data[i-1] = PathData(tc.Y)
			}
		case Pcv:
			for np := 0; np < n; np++ {
				rp.Y = PathDataNext(data, &i)
				rp.X = 0
				rp = xf.MulVector2AsVector(rp)
				data[i-1] = PathData(rp.Y)
				cp.SetAdd(rp) // new abs
			}
		case PcC:
			for np := 0; np < n/6; np++ {
				xp = PathDataTransformAbs(data, &i, xf, lpt)
				ctrl = PathDataTransformAbs(data, &i, xf, lpt)
				cp = PathDataTransformAbs(data, &i, xf, lpt)
			}
		case Pcc:
			for np := 0; np < n/6; np++ {
				xp = PathDataTransformRel(data, &i, xf, cp)
				ctrl = PathDataTransformRel(data, &i, xf, cp)
				cp = PathDataTransformRel(data, &i, xf, cp)
			}
		case Pcs:
			rel = true
			fallthrough
		case PcS:
			for np := 0; np < n/4; np++ {
				switch lastCmd {
				case Pcc, PcC, Pcs, PcS:
					ctrl = reflectPt(cp, ctrl)
				default:
					ctrl = cp
				}
				if rel {
					xp = PathDataTransformRel(data, &i, xf, cp)
					cp = PathDataTransformRel(data, &i, xf, cp)
				} else {
					xp = PathDataTransformAbs(data, &i, xf, lpt)
					cp = PathDataTransformAbs(data, &i, xf, lpt)
				}
				lastCmd = cmd
				ctrl = xp
			}
		case PcQ:
			for np := 0; np < n/4; np++ {
				ctrl = PathDataTransformAbs(data, &i, xf, lpt)
				cp = PathDataTransformAbs(data, &i, xf, lpt)
			}
		case Pcq:
			for np := 0; np < n/4; np++ {
				ctrl = PathDataTransformRel(data, &i, xf, cp)
				cp = PathDataTransformRel(data, &i, xf, cp)
			}
		case Pct:
			rel = true
			fallthrough
		case PcT:
			for np := 0; np < n/2; np++ {
				switch lastCmd {
				case Pcq, PcQ, PcT, Pct:
					ctrl = reflectPt(cp, ctrl)
				default:
					ctrl = cp
				}
				if rel {
					cp = PathDataTransformRel(data, &i, xf, cp)
				} else {
					cp = PathDataTransformAbs(data, &i, xf, lpt)
				}
				lastCmd = cmd
			}
		case Pca:
			rel = true
			fallthrough
		case PcA:
			for np := 0; np < n/7; np++ {
				rad := PathDataTransformRel(data, &i, xf, math32.Vector2{})
				ang := PathDataNext(data, &i)
				largeArc := (PathDataNext(data, &i) != 0)
				sweep := (PathDataNext(data, &i) != 0)
				pc := cp
				if rel {
					cp = PathDataTransformRel(data, &i, xf, cp)
				} else {
					cp = PathDataTransformAbs(data, &i, xf, lpt)
				}
				ncx, ncy := paint.FindEllipseCenter(&rad.X, &rad.Y, ang*math.Pi/180, pc.X, pc.Y, cp.X, cp.Y, sweep, largeArc)
				_ = ncx
				_ = ncy
			}
		case PcZ:
			fallthrough
		case Pcz:
			cp = st
		}
		lastCmd = cmd
	}

}

// WriteGeom writes the geometry of the node to a slice of floating point numbers
// the length and ordering of which is specific to each node type.
// Slice must be passed and will be resized if not the correct length.
func (g *Path) WriteGeom(sv *SVG, dat *[]float32) {
	sz := len(g.Data)
	*dat = slicesx.SetLength(*dat, sz+6)
	for i := range g.Data {
		(*dat)[i] = float32(g.Data[i])
	}
	g.WriteTransform(*dat, sz)
	g.GradientWritePts(sv, dat)
}

// ReadGeom reads the geometry of the node from a slice of floating point numbers
// the length and ordering of which is specific to each node type.
func (g *Path) ReadGeom(sv *SVG, dat []float32) {
	sz := len(g.Data)
	for i := range g.Data {
		g.Data[i] = PathData(dat[i])
	}
	g.ReadTransform(dat, sz)
	g.GradientReadPts(sv, dat)
}
