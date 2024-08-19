// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package styles

import (
	"log"
	"strconv"
	"strings"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/reflectx"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/colors/gradient"
	"cogentcore.org/core/enums"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles/units"
)

////////////////////////////////////////////////////////////////////////////
//   Styling functions for setting from properties
//     see style_properties.go for master version

// StyleFromProperties sets style field values based on map[string]any properties
func (pc *Paint) StyleFromProperties(parent *Paint, properties map[string]any, cc colors.Context) {
	for key, val := range properties {
		if len(key) == 0 {
			continue
		}
		if key[0] == '#' || key[0] == '.' || key[0] == ':' || key[0] == '_' {
			continue
		}
		if key == "display" {
			if inh, init := StyleInhInit(val, parent); inh || init {
				if inh {
					pc.Display = parent.Display
				} else if init {
					pc.Display = true
				}
				return
			}
			sval := reflectx.ToString(val)
			switch sval {
			case "none":
				pc.Display = false
			case "inline":
				pc.Display = true
			default:
				pc.Display = true
			}
			continue
		}
		if sfunc, ok := StyleStrokeFuncs[key]; ok {
			if parent != nil {
				sfunc(&pc.StrokeStyle, key, val, &parent.StrokeStyle, cc)
			} else {
				sfunc(&pc.StrokeStyle, key, val, nil, cc)
			}
			continue
		}
		if sfunc, ok := StyleFillFuncs[key]; ok {
			if parent != nil {
				sfunc(&pc.FillStyle, key, val, &parent.FillStyle, cc)
			} else {
				sfunc(&pc.FillStyle, key, val, nil, cc)
			}
			continue
		}
		if sfunc, ok := StyleFontFuncs[key]; ok {
			if parent != nil {
				sfunc(&pc.FontStyle.Font, key, val, &parent.FontStyle.Font, cc)
			} else {
				sfunc(&pc.FontStyle.Font, key, val, nil, cc)
			}
			continue
		}
		if sfunc, ok := StyleFontRenderFuncs[key]; ok {
			if parent != nil {
				sfunc(&pc.FontStyle, key, val, &parent.FontStyle, cc)
			} else {
				sfunc(&pc.FontStyle, key, val, nil, cc)
			}
			continue
		}
		if sfunc, ok := StyleTextFuncs[key]; ok {
			if parent != nil {
				sfunc(&pc.TextStyle, key, val, &parent.TextStyle, cc)
			} else {
				sfunc(&pc.TextStyle, key, val, nil, cc)
			}
			continue
		}
		if sfunc, ok := StylePaintFuncs[key]; ok {
			sfunc(pc, key, val, parent, cc)
			continue
		}
	}
}

/////////////////////////////////////////////////////////////////////////////////
//  Stroke

// StyleStrokeFuncs are functions for styling the Stroke object
var StyleStrokeFuncs = map[string]StyleFunc{
	"stroke": func(obj any, key string, val any, parent any, cc colors.Context) {
		fs := obj.(*Stroke)
		if inh, init := StyleInhInit(val, parent); inh || init {
			if inh {
				fs.Color = parent.(*Stroke).Color
			} else if init {
				fs.Color = colors.Uniform(colors.Black)
			}
			return
		}
		fs.Color = errors.Log1(gradient.FromAny(val, cc))
	},
	"stroke-opacity": StyleFuncFloat(float32(1),
		func(obj *Stroke) *float32 { return &(obj.Opacity) }),
	"stroke-width": StyleFuncUnits(units.Dp(1),
		func(obj *Stroke) *units.Value { return &(obj.Width) }),
	"stroke-min-width": StyleFuncUnits(units.Dp(1),
		func(obj *Stroke) *units.Value { return &(obj.MinWidth) }),
	"stroke-dasharray": func(obj any, key string, val any, parent any, cc colors.Context) {
		fs := obj.(*Stroke)
		if inh, init := StyleInhInit(val, parent); inh || init {
			if inh {
				fs.Dashes = parent.(*Stroke).Dashes
			} else if init {
				fs.Dashes = nil
			}
			return
		}
		switch vt := val.(type) {
		case string:
			fs.Dashes = ParseDashesString(vt)
		case []float32:
			math32.CopyFloat32s(&fs.Dashes, vt)
		case *[]float32:
			math32.CopyFloat32s(&fs.Dashes, *vt)
		}
	},
	"stroke-linecap": StyleFuncEnum(LineCapButt,
		func(obj *Stroke) enums.EnumSetter { return &(obj.Cap) }),
	"stroke-linejoin": StyleFuncEnum(LineJoinMiter,
		func(obj *Stroke) enums.EnumSetter { return &(obj.Join) }),
	"stroke-miterlimit": StyleFuncFloat(float32(1),
		func(obj *Stroke) *float32 { return &(obj.MiterLimit) }),
}

/////////////////////////////////////////////////////////////////////////////////
//  Fill

// StyleFillFuncs are functions for styling the Fill object
var StyleFillFuncs = map[string]StyleFunc{
	"fill": func(obj any, key string, val any, parent any, cc colors.Context) {
		fs := obj.(*Fill)
		if inh, init := StyleInhInit(val, parent); inh || init {
			if inh {
				fs.Color = parent.(*Fill).Color
			} else if init {
				fs.Color = colors.Uniform(colors.Black)
			}
			return
		}
		fs.Color = errors.Log1(gradient.FromAny(val, cc))
	},
	"fill-opacity": StyleFuncFloat(float32(1),
		func(obj *Fill) *float32 { return &(obj.Opacity) }),
	"fill-rule": StyleFuncEnum(FillRuleNonZero,
		func(obj *Fill) enums.EnumSetter { return &(obj.Rule) }),
}

/////////////////////////////////////////////////////////////////////////////////
//  Paint

// StylePaintFuncs are functions for styling the Stroke object
var StylePaintFuncs = map[string]StyleFunc{
	"vector-effect": StyleFuncEnum(VectorEffectNone,
		func(obj *Paint) enums.EnumSetter { return &(obj.VectorEffect) }),
	"transform": func(obj any, key string, val any, parent any, cc colors.Context) {
		pc := obj.(*Paint)
		if inh, init := StyleInhInit(val, parent); inh || init {
			if inh {
				pc.Transform = parent.(*Paint).Transform
			} else if init {
				pc.Transform = math32.Identity2()
			}
			return
		}
		switch vt := val.(type) {
		case string:
			pc.Transform.SetString(vt)
		case *math32.Matrix2:
			pc.Transform = *vt
		case math32.Matrix2:
			pc.Transform = vt
		}
	},
}

// ParseDashesString gets a dash slice from given string
func ParseDashesString(str string) []float32 {
	if len(str) == 0 || str == "none" {
		return nil
	}
	ds := strings.Split(str, ",")
	dl := make([]float32, len(ds))
	for i, dstr := range ds {
		d, err := strconv.ParseFloat(strings.TrimSpace(dstr), 32)
		if err != nil {
			log.Printf("core.ParseDashesString parsing error: %v\n", err)
			return nil
		}
		dl[i] = float32(d)
	}
	return dl
}
