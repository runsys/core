// Code generated by "core generate"; DO NOT EDIT.

package units

import (
	"cogentcore.org/core/enums"
)

var _UnitsValues = []Units{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

// UnitsN is the highest valid value for type Units, plus one.
const UnitsN Units = 21

var _UnitsValueMap = map[string]Units{`dp`: 0, `px`: 1, `ew`: 2, `eh`: 3, `pw`: 4, `ph`: 5, `rem`: 6, `em`: 7, `ex`: 8, `ch`: 9, `vw`: 10, `vh`: 11, `vmin`: 12, `vmax`: 13, `cm`: 14, `mm`: 15, `q`: 16, `in`: 17, `pc`: 18, `pt`: 19, `dot`: 20}

var _UnitsDescMap = map[Units]string{0: `UnitDp represents density-independent pixels. 1dp is 1/160 in.`, 1: `UnitPx represents logical pixels. 1px is 1/96 in. These are not raw display pixels, for which you should use dots.`, 2: `UnitEw represents percentage of element width, which is equivalent to CSS % in some contexts.`, 3: `UnitEh represents percentage of element height, which is equivalent to CSS % in some contexts.`, 4: `UnitPw represents percentage of parent width, which is equivalent to CSS % in some contexts.`, 5: `UnitPh represents percentage of parent height, which is equivalent to CSS % in some contexts.`, 6: `UnitRem represents the font size of the root element, which is always 16dp.`, 7: `UnitEm represents the font size of the element.`, 8: `UnitEx represents x-height of the element&#39;s font (size of &#39;x&#39; glyph). It falls back to a default of 0.5em.`, 9: `UnitCh represents width of the &#39;0&#39; glyph in the element&#39;s font. It falls back to a default of 0.5em.`, 10: `UnitVw represents percentage of viewport (Scene) width.`, 11: `UnitVh represents percentage of viewport (Scene) height.`, 12: `UnitVmin represents percentage of the smaller dimension of the viewport (Scene).`, 13: `UnitVmax represents percentage of the larger dimension of the viewport (Scene).`, 14: `UnitCm represents centimeters. 1cm is 1/2.54 in.`, 15: `UnitMm represents millimeters. 1mm is 1/10 cm.`, 16: `UnitQ represents quarter-millimeters. 1q is 1/40 cm.`, 17: `UnitIn represents inches. 1in is 2.54cm or 96px.`, 18: `UnitPc represents picas. 1pc is 1/6 in.`, 19: `UnitPt represents points. 1pt is 1/72 in.`, 20: `UnitDot represents real display pixels. They are generally only used internally.`}

var _UnitsMap = map[Units]string{0: `dp`, 1: `px`, 2: `ew`, 3: `eh`, 4: `pw`, 5: `ph`, 6: `rem`, 7: `em`, 8: `ex`, 9: `ch`, 10: `vw`, 11: `vh`, 12: `vmin`, 13: `vmax`, 14: `cm`, 15: `mm`, 16: `q`, 17: `in`, 18: `pc`, 19: `pt`, 20: `dot`}

// String returns the string representation of this Units value.
func (i Units) String() string { return enums.String(i, _UnitsMap) }

// SetString sets the Units value from its string representation,
// and returns an error if the string is invalid.
func (i *Units) SetString(s string) error { return enums.SetString(i, s, _UnitsValueMap, "Units") }

// Int64 returns the Units value as an int64.
func (i Units) Int64() int64 { return int64(i) }

// SetInt64 sets the Units value from an int64.
func (i *Units) SetInt64(in int64) { *i = Units(in) }

// Desc returns the description of the Units value.
func (i Units) Desc() string { return enums.Desc(i, _UnitsDescMap) }

// UnitsValues returns all possible values for the type Units.
func UnitsValues() []Units { return _UnitsValues }

// Values returns all possible values for the type Units.
func (i Units) Values() []enums.Enum { return enums.Values(_UnitsValues) }

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i Units) MarshalText() ([]byte, error) { return []byte(i.String()), nil }

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *Units) UnmarshalText(text []byte) error { return enums.UnmarshalText(i, text, "Units") }
