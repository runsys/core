// Code generated by "core generate -add-types"; DO NOT EDIT.

package pi

import (
	"errors"
	"log"
	"strconv"

	"cogentcore.org/core/enums"
)

var _LangFlagsValues = []LangFlags{0, 1, 2, 3}

// LangFlagsN is the highest valid value
// for type LangFlags, plus one.
const LangFlagsN LangFlags = 4

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the enumgen command to generate them again.
func _LangFlagsNoOp() {
	var x [1]struct{}
	_ = x[NoFlags-(0)]
	_ = x[IndentSpace-(1)]
	_ = x[IndentTab-(2)]
	_ = x[ReAutoIndent-(3)]
}

var _LangFlagsNameToValueMap = map[string]LangFlags{
	`NoFlags`:      0,
	`IndentSpace`:  1,
	`IndentTab`:    2,
	`ReAutoIndent`: 3,
}

var _LangFlagsDescMap = map[LangFlags]string{
	0: `NoFlags = nothing special`,
	1: `IndentSpace means that spaces must be used for this language`,
	2: `IndentTab means that tabs must be used for this language`,
	3: `ReAutoIndent causes current line to be re-indented during AutoIndent for Enter (newline) -- this should only be set for strongly indented languages where the previous + current line can tell you exactly what indent the current line should be at.`,
}

var _LangFlagsMap = map[LangFlags]string{
	0: `NoFlags`,
	1: `IndentSpace`,
	2: `IndentTab`,
	3: `ReAutoIndent`,
}

// String returns the string representation
// of this LangFlags value.
func (f LangFlags) String() string {
	if str, ok := _LangFlagsMap[f]; ok {
		return str
	}
	return strconv.FormatInt(int64(f), 10)
}

// SetString sets the LangFlags value from its
// string representation, and returns an
// error if the string is invalid.
func (f *LangFlags) SetString(s string) error {
	if val, ok := _LangFlagsNameToValueMap[s]; ok {
		*f = val
		return nil
	}
	return errors.New(s + " is not a valid value for type LangFlags")
}

// Int64 returns the LangFlags value as an int64.
func (f LangFlags) Int64() int64 {
	return int64(f)
}

// SetInt64 sets the LangFlags value from an int64.
func (f *LangFlags) SetInt64(in int64) {
	*f = LangFlags(in)
}

// Desc returns the description of the LangFlags value.
func (f LangFlags) Desc() string {
	if str, ok := _LangFlagsDescMap[f]; ok {
		return str
	}
	return f.String()
}

// LangFlagsValues returns all possible values
// for the type LangFlags.
func LangFlagsValues() []LangFlags {
	return _LangFlagsValues
}

// Values returns all possible values
// for the type LangFlags.
func (f LangFlags) Values() []enums.Enum {
	res := make([]enums.Enum, len(_LangFlagsValues))
	for i, d := range _LangFlagsValues {
		res[i] = d
	}
	return res
}

// IsValid returns whether the value is a
// valid option for type LangFlags.
func (f LangFlags) IsValid() bool {
	_, ok := _LangFlagsMap[f]
	return ok
}

// MarshalText implements the [encoding.TextMarshaler] interface.
func (f LangFlags) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (f *LangFlags) UnmarshalText(text []byte) error {
	if err := f.SetString(string(text)); err != nil {
		log.Println("LangFlags.UnmarshalText:", err)
	}
	return nil
}
