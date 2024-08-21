// Code generated by "core generate"; DO NOT EDIT.

package texteditor

import (
	"cogentcore.org/core/enums"
)

var _bufferSignalsValues = []bufferSignals{0, 1, 2, 3, 4, 5, 6}

// bufferSignalsN is the highest valid value for type bufferSignals, plus one.
const bufferSignalsN bufferSignals = 7

var _bufferSignalsValueMap = map[string]bufferSignals{`Done`: 0, `New`: 1, `Mods`: 2, `Insert`: 3, `Delete`: 4, `MarkupUpdated`: 5, `Closed`: 6}

var _bufferSignalsDescMap = map[bufferSignals]string{0: `bufferDone means that editing was completed and applied to Txt field -- data is Txt bytes`, 1: `bufferNew signals that entirely new text is present. All views should do full layout update.`, 2: `bufferMods signals that potentially diffuse modifications have been made. Views should do a Layout and Render.`, 3: `bufferInsert signals that some text was inserted. data is text.Edit describing change. The Buf always reflects the current state *after* the edit.`, 4: `bufferDelete signals that some text was deleted. data is text.Edit describing change. The Buf always reflects the current state *after* the edit.`, 5: `bufferMarkupUpdated signals that the Markup text has been updated This signal is typically sent from a separate goroutine, so should be used with a mutex`, 6: `bufferClosed signals that the text was closed.`}

var _bufferSignalsMap = map[bufferSignals]string{0: `Done`, 1: `New`, 2: `Mods`, 3: `Insert`, 4: `Delete`, 5: `MarkupUpdated`, 6: `Closed`}

// String returns the string representation of this bufferSignals value.
func (i bufferSignals) String() string { return enums.String(i, _bufferSignalsMap) }

// SetString sets the bufferSignals value from its string representation,
// and returns an error if the string is invalid.
func (i *bufferSignals) SetString(s string) error {
	return enums.SetString(i, s, _bufferSignalsValueMap, "bufferSignals")
}

// Int64 returns the bufferSignals value as an int64.
func (i bufferSignals) Int64() int64 { return int64(i) }

// SetInt64 sets the bufferSignals value from an int64.
func (i *bufferSignals) SetInt64(in int64) { *i = bufferSignals(in) }

// Desc returns the description of the bufferSignals value.
func (i bufferSignals) Desc() string { return enums.Desc(i, _bufferSignalsDescMap) }

// bufferSignalsValues returns all possible values for the type bufferSignals.
func bufferSignalsValues() []bufferSignals { return _bufferSignalsValues }

// Values returns all possible values for the type bufferSignals.
func (i bufferSignals) Values() []enums.Enum { return enums.Values(_bufferSignalsValues) }

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i bufferSignals) MarshalText() ([]byte, error) { return []byte(i.String()), nil }

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *bufferSignals) UnmarshalText(text []byte) error {
	return enums.UnmarshalText(i, text, "bufferSignals")
}
