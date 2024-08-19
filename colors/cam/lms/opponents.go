// Copyright (c) 2021, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lms

//go:generate core generate

// OpVals holds color opponent values based on cone-like L,M,S inputs
// These values are useful for generating inputs to vision models that
// simulate color opponency representations in the brain.
type OpVals struct {

	// red vs. green (long vs. medium)
	RedGreen float32

	// blue vs. yellow (short vs. avg(long, medium))
	BlueYellow float32

	// greyscale luminance channel -- typically use L* from LAB as best
	Grey float32
}

// NewOpVals returns a new opponent color values from values representing
// the LMS long, medium, short cone responses, and an overall grey value
func NewOpVals(l, m, s, lm, grey float32) OpVals {
	return OpVals{RedGreen: l - m, BlueYellow: s - lm, Grey: grey}
}

// Opponents enumerates the three primary opponency channels:
// WhiteBlack, RedGreen, BlueYellow
// using colloquial "everyday" terms.
type Opponents int32 //enums:enum

const (
	// White vs. Black greyscale
	WhiteBlack Opponents = iota

	// Red vs. Green
	RedGreen

	// Blue vs. Yellow
	BlueYellow
)
