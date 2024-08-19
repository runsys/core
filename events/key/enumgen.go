// Code generated by "core generate"; DO NOT EDIT.

package key

import (
	"cogentcore.org/core/enums"
)

var _CodesValues = []Codes{0, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 117, 127, 128, 129, 224, 225, 226, 227, 228, 229, 230, 231, 65536}

// CodesN is the highest valid value for type Codes, plus one.
const CodesN Codes = 65537

var _CodesValueMap = map[string]Codes{`Unknown`: 0, `A`: 4, `B`: 5, `C`: 6, `D`: 7, `E`: 8, `F`: 9, `G`: 10, `H`: 11, `I`: 12, `J`: 13, `K`: 14, `L`: 15, `M`: 16, `N`: 17, `O`: 18, `P`: 19, `Q`: 20, `R`: 21, `S`: 22, `T`: 23, `U`: 24, `V`: 25, `W`: 26, `X`: 27, `Y`: 28, `Z`: 29, `1`: 30, `2`: 31, `3`: 32, `4`: 33, `5`: 34, `6`: 35, `7`: 36, `8`: 37, `9`: 38, `0`: 39, `ReturnEnter`: 40, `Escape`: 41, `Backspace`: 42, `Tab`: 43, `Spacebar`: 44, `HyphenMinus`: 45, `EqualSign`: 46, `LeftSquareBracket`: 47, `RightSquareBracket`: 48, `Backslash`: 49, `Semicolon`: 51, `Apostrophe`: 52, `GraveAccent`: 53, `Comma`: 54, `FullStop`: 55, `Slash`: 56, `CapsLock`: 57, `F1`: 58, `F2`: 59, `F3`: 60, `F4`: 61, `F5`: 62, `F6`: 63, `F7`: 64, `F8`: 65, `F9`: 66, `F10`: 67, `F11`: 68, `F12`: 69, `Pause`: 72, `Insert`: 73, `Home`: 74, `PageUp`: 75, `Delete`: 76, `End`: 77, `PageDown`: 78, `RightArrow`: 79, `LeftArrow`: 80, `DownArrow`: 81, `UpArrow`: 82, `KeypadNumLock`: 83, `KeypadSlash`: 84, `KeypadAsterisk`: 85, `KeypadHyphenMinus`: 86, `KeypadPlusSign`: 87, `KeypadEnter`: 88, `Keypad1`: 89, `Keypad2`: 90, `Keypad3`: 91, `Keypad4`: 92, `Keypad5`: 93, `Keypad6`: 94, `Keypad7`: 95, `Keypad8`: 96, `Keypad9`: 97, `Keypad0`: 98, `KeypadFullStop`: 99, `KeypadEqualSign`: 103, `F13`: 104, `F14`: 105, `F15`: 106, `F16`: 107, `F17`: 108, `F18`: 109, `F19`: 110, `F20`: 111, `F21`: 112, `F22`: 113, `F23`: 114, `F24`: 115, `Help`: 117, `Mute`: 127, `VolumeUp`: 128, `VolumeDown`: 129, `LeftControl`: 224, `LeftShift`: 225, `LeftAlt`: 226, `LeftMeta`: 227, `RightControl`: 228, `RightShift`: 229, `RightAlt`: 230, `RightMeta`: 231, `Compose`: 65536}

var _CodesDescMap = map[Codes]string{0: ``, 4: ``, 5: ``, 6: ``, 7: ``, 8: ``, 9: ``, 10: ``, 11: ``, 12: ``, 13: ``, 14: ``, 15: ``, 16: ``, 17: ``, 18: ``, 19: ``, 20: ``, 21: ``, 22: ``, 23: ``, 24: ``, 25: ``, 26: ``, 27: ``, 28: ``, 29: ``, 30: ``, 31: ``, 32: ``, 33: ``, 34: ``, 35: ``, 36: ``, 37: ``, 38: ``, 39: ``, 40: ``, 41: ``, 42: ``, 43: ``, 44: ``, 45: ``, 46: ``, 47: ``, 48: ``, 49: ``, 51: ``, 52: ``, 53: ``, 54: ``, 55: ``, 56: ``, 57: ``, 58: ``, 59: ``, 60: ``, 61: ``, 62: ``, 63: ``, 64: ``, 65: ``, 66: ``, 67: ``, 68: ``, 69: ``, 72: ``, 73: ``, 74: ``, 75: ``, 76: ``, 77: ``, 78: ``, 79: ``, 80: ``, 81: ``, 82: ``, 83: ``, 84: ``, 85: ``, 86: ``, 87: ``, 88: ``, 89: ``, 90: ``, 91: ``, 92: ``, 93: ``, 94: ``, 95: ``, 96: ``, 97: ``, 98: ``, 99: ``, 103: ``, 104: ``, 105: ``, 106: ``, 107: ``, 108: ``, 109: ``, 110: ``, 111: ``, 112: ``, 113: ``, 114: ``, 115: ``, 117: ``, 127: ``, 128: ``, 129: ``, 224: ``, 225: ``, 226: ``, 227: ``, 228: ``, 229: ``, 230: ``, 231: ``, 65536: `CodeCompose is the Code for a compose key, sometimes called a multi key, used to input non-ASCII characters such as Ã± being composed of n and ~. See https://en.wikipedia.org/wiki/Compose_key`}

var _CodesMap = map[Codes]string{0: `Unknown`, 4: `A`, 5: `B`, 6: `C`, 7: `D`, 8: `E`, 9: `F`, 10: `G`, 11: `H`, 12: `I`, 13: `J`, 14: `K`, 15: `L`, 16: `M`, 17: `N`, 18: `O`, 19: `P`, 20: `Q`, 21: `R`, 22: `S`, 23: `T`, 24: `U`, 25: `V`, 26: `W`, 27: `X`, 28: `Y`, 29: `Z`, 30: `1`, 31: `2`, 32: `3`, 33: `4`, 34: `5`, 35: `6`, 36: `7`, 37: `8`, 38: `9`, 39: `0`, 40: `ReturnEnter`, 41: `Escape`, 42: `Backspace`, 43: `Tab`, 44: `Spacebar`, 45: `HyphenMinus`, 46: `EqualSign`, 47: `LeftSquareBracket`, 48: `RightSquareBracket`, 49: `Backslash`, 51: `Semicolon`, 52: `Apostrophe`, 53: `GraveAccent`, 54: `Comma`, 55: `FullStop`, 56: `Slash`, 57: `CapsLock`, 58: `F1`, 59: `F2`, 60: `F3`, 61: `F4`, 62: `F5`, 63: `F6`, 64: `F7`, 65: `F8`, 66: `F9`, 67: `F10`, 68: `F11`, 69: `F12`, 72: `Pause`, 73: `Insert`, 74: `Home`, 75: `PageUp`, 76: `Delete`, 77: `End`, 78: `PageDown`, 79: `RightArrow`, 80: `LeftArrow`, 81: `DownArrow`, 82: `UpArrow`, 83: `KeypadNumLock`, 84: `KeypadSlash`, 85: `KeypadAsterisk`, 86: `KeypadHyphenMinus`, 87: `KeypadPlusSign`, 88: `KeypadEnter`, 89: `Keypad1`, 90: `Keypad2`, 91: `Keypad3`, 92: `Keypad4`, 93: `Keypad5`, 94: `Keypad6`, 95: `Keypad7`, 96: `Keypad8`, 97: `Keypad9`, 98: `Keypad0`, 99: `KeypadFullStop`, 103: `KeypadEqualSign`, 104: `F13`, 105: `F14`, 106: `F15`, 107: `F16`, 108: `F17`, 109: `F18`, 110: `F19`, 111: `F20`, 112: `F21`, 113: `F22`, 114: `F23`, 115: `F24`, 117: `Help`, 127: `Mute`, 128: `VolumeUp`, 129: `VolumeDown`, 224: `LeftControl`, 225: `LeftShift`, 226: `LeftAlt`, 227: `LeftMeta`, 228: `RightControl`, 229: `RightShift`, 230: `RightAlt`, 231: `RightMeta`, 65536: `Compose`}

// String returns the string representation of this Codes value.
func (i Codes) String() string { return enums.String(i, _CodesMap) }

// SetString sets the Codes value from its string representation,
// and returns an error if the string is invalid.
func (i *Codes) SetString(s string) error { return enums.SetString(i, s, _CodesValueMap, "Codes") }

// Int64 returns the Codes value as an int64.
func (i Codes) Int64() int64 { return int64(i) }

// SetInt64 sets the Codes value from an int64.
func (i *Codes) SetInt64(in int64) { *i = Codes(in) }

// Desc returns the description of the Codes value.
func (i Codes) Desc() string { return enums.Desc(i, _CodesDescMap) }

// CodesValues returns all possible values for the type Codes.
func CodesValues() []Codes { return _CodesValues }

// Values returns all possible values for the type Codes.
func (i Codes) Values() []enums.Enum { return enums.Values(_CodesValues) }

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i Codes) MarshalText() ([]byte, error) { return []byte(i.String()), nil }

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *Codes) UnmarshalText(text []byte) error { return enums.UnmarshalText(i, text, "Codes") }

var _ModifiersValues = []Modifiers{0, 1, 2, 3}

// ModifiersN is the highest valid value for type Modifiers, plus one.
const ModifiersN Modifiers = 4

var _ModifiersValueMap = map[string]Modifiers{`Control`: 0, `Meta`: 1, `Alt`: 2, `Shift`: 3}

var _ModifiersDescMap = map[Modifiers]string{0: `Control is the &#34;Control&#34; (Ctrl) key.`, 1: `Meta is the system meta key (the &#34;Command&#34; key on macOS and the Windows key on Windows).`, 2: `Alt is the &#34;Alt&#34; (&#34;Option&#34; on macOS) key.`, 3: `Shift is the &#34;Shift&#34; key.`}

var _ModifiersMap = map[Modifiers]string{0: `Control`, 1: `Meta`, 2: `Alt`, 3: `Shift`}

// String returns the string representation of this Modifiers value.
func (i Modifiers) String() string { return enums.BitFlagString(i, _ModifiersValues) }

// BitIndexString returns the string representation of this Modifiers value
// if it is a bit index value (typically an enum constant), and
// not an actual bit flag value.
func (i Modifiers) BitIndexString() string { return enums.String(i, _ModifiersMap) }

// SetString sets the Modifiers value from its string representation,
// and returns an error if the string is invalid.
func (i *Modifiers) SetString(s string) error { *i = 0; return i.SetStringOr(s) }

// SetStringOr sets the Modifiers value from its string representation
// while preserving any bit flags already set, and returns an
// error if the string is invalid.
func (i *Modifiers) SetStringOr(s string) error {
	return enums.SetStringOr(i, s, _ModifiersValueMap, "Modifiers")
}

// Int64 returns the Modifiers value as an int64.
func (i Modifiers) Int64() int64 { return int64(i) }

// SetInt64 sets the Modifiers value from an int64.
func (i *Modifiers) SetInt64(in int64) { *i = Modifiers(in) }

// Desc returns the description of the Modifiers value.
func (i Modifiers) Desc() string { return enums.Desc(i, _ModifiersDescMap) }

// ModifiersValues returns all possible values for the type Modifiers.
func ModifiersValues() []Modifiers { return _ModifiersValues }

// Values returns all possible values for the type Modifiers.
func (i Modifiers) Values() []enums.Enum { return enums.Values(_ModifiersValues) }

// HasFlag returns whether these bit flags have the given bit flag set.
func (i Modifiers) HasFlag(f enums.BitFlag) bool { return enums.HasFlag((*int64)(&i), f) }

// SetFlag sets the value of the given flags in these flags to the given value.
func (i *Modifiers) SetFlag(on bool, f ...enums.BitFlag) { enums.SetFlag((*int64)(i), on, f...) }

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i Modifiers) MarshalText() ([]byte, error) { return []byte(i.String()), nil }

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *Modifiers) UnmarshalText(text []byte) error {
	return enums.UnmarshalText(i, text, "Modifiers")
}
