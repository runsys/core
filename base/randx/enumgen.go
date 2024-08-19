// Code generated by "core generate -add-types"; DO NOT EDIT.

package randx

import (
	"cogentcore.org/core/enums"
)

var _RandDistsValues = []RandDists{0, 1, 2, 3, 4, 5, 6}

// RandDistsN is the highest valid value for type RandDists, plus one.
const RandDistsN RandDists = 7

var _RandDistsValueMap = map[string]RandDists{`Uniform`: 0, `Binomial`: 1, `Poisson`: 2, `Gamma`: 3, `Gaussian`: 4, `Beta`: 5, `Mean`: 6}

var _RandDistsDescMap = map[RandDists]string{0: `Uniform has a uniform probability distribution over Var = range on either side of the Mean`, 1: `Binomial represents number of 1&#39;s in n (Par) random (Bernouli) trials of probability p (Var)`, 2: `Poisson represents number of events in interval, with event rate (lambda = Var) plus Mean`, 3: `Gamma represents maximum entropy distribution with two parameters: scaling parameter (Var) and shape parameter k (Par) plus Mean`, 4: `Gaussian normal with Var = stddev plus Mean`, 5: `Beta with Var = alpha and Par = beta shape parameters`, 6: `Mean is just the constant Mean, no randomness`}

var _RandDistsMap = map[RandDists]string{0: `Uniform`, 1: `Binomial`, 2: `Poisson`, 3: `Gamma`, 4: `Gaussian`, 5: `Beta`, 6: `Mean`}

// String returns the string representation of this RandDists value.
func (i RandDists) String() string { return enums.String(i, _RandDistsMap) }

// SetString sets the RandDists value from its string representation,
// and returns an error if the string is invalid.
func (i *RandDists) SetString(s string) error {
	return enums.SetString(i, s, _RandDistsValueMap, "RandDists")
}

// Int64 returns the RandDists value as an int64.
func (i RandDists) Int64() int64 { return int64(i) }

// SetInt64 sets the RandDists value from an int64.
func (i *RandDists) SetInt64(in int64) { *i = RandDists(in) }

// Desc returns the description of the RandDists value.
func (i RandDists) Desc() string { return enums.Desc(i, _RandDistsDescMap) }

// RandDistsValues returns all possible values for the type RandDists.
func RandDistsValues() []RandDists { return _RandDistsValues }

// Values returns all possible values for the type RandDists.
func (i RandDists) Values() []enums.Enum { return enums.Values(_RandDistsValues) }

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i RandDists) MarshalText() ([]byte, error) { return []byte(i.String()), nil }

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *RandDists) UnmarshalText(text []byte) error {
	return enums.UnmarshalText(i, text, "RandDists")
}
