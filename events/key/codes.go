// Copyright (c) 2018, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// based on golang.org/x/mobile/event:
//
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package key defines an event for physical keyboard keys, for the Cogent Core GUI
// system.
//
// On-screen software keyboards do not send key events.
package key

// Codes is the identity of a key relative to a notional "standard" keyboard.
type Codes uint32 //enums:enum -trim-prefix Code

// Physical key codes.
//
// For standard key codes, its value matches USB HID key codes.
// TODO: add missing codes.
const (
	CodeUnknown Codes = 0

	CodeA Codes = 4
	CodeB Codes = 5
	CodeC Codes = 6
	CodeD Codes = 7
	CodeE Codes = 8
	CodeF Codes = 9
	CodeG Codes = 10
	CodeH Codes = 11
	CodeI Codes = 12
	CodeJ Codes = 13
	CodeK Codes = 14
	CodeL Codes = 15
	CodeM Codes = 16
	CodeN Codes = 17
	CodeO Codes = 18
	CodeP Codes = 19
	CodeQ Codes = 20
	CodeR Codes = 21
	CodeS Codes = 22
	CodeT Codes = 23
	CodeU Codes = 24
	CodeV Codes = 25
	CodeW Codes = 26
	CodeX Codes = 27
	CodeY Codes = 28
	CodeZ Codes = 29

	Code1 Codes = 30
	Code2 Codes = 31
	Code3 Codes = 32
	Code4 Codes = 33
	Code5 Codes = 34
	Code6 Codes = 35
	Code7 Codes = 36
	Code8 Codes = 37
	Code9 Codes = 38
	Code0 Codes = 39

	CodeReturnEnter        Codes = 40
	CodeEscape             Codes = 41
	CodeBackspace          Codes = 42
	CodeTab                Codes = 43
	CodeSpacebar           Codes = 44
	CodeHyphenMinus        Codes = 45 // -
	CodeEqualSign          Codes = 46 // =
	CodeLeftSquareBracket  Codes = 47 // [
	CodeRightSquareBracket Codes = 48 // ]
	CodeBackslash          Codes = 49 // \
	CodeSemicolon          Codes = 51 // ;
	CodeApostrophe         Codes = 52 // '
	CodeGraveAccent        Codes = 53 // `
	CodeComma              Codes = 54 // ,
	CodeFullStop           Codes = 55 // .
	CodeSlash              Codes = 56 // /
	CodeCapsLock           Codes = 57

	CodeF1  Codes = 58
	CodeF2  Codes = 59
	CodeF3  Codes = 60
	CodeF4  Codes = 61
	CodeF5  Codes = 62
	CodeF6  Codes = 63
	CodeF7  Codes = 64
	CodeF8  Codes = 65
	CodeF9  Codes = 66
	CodeF10 Codes = 67
	CodeF11 Codes = 68
	CodeF12 Codes = 69

	CodePause    Codes = 72
	CodeInsert   Codes = 73
	CodeHome     Codes = 74
	CodePageUp   Codes = 75
	CodeDelete   Codes = 76
	CodeEnd      Codes = 77
	CodePageDown Codes = 78

	CodeRightArrow Codes = 79
	CodeLeftArrow  Codes = 80
	CodeDownArrow  Codes = 81
	CodeUpArrow    Codes = 82

	CodeKeypadNumLock     Codes = 83
	CodeKeypadSlash       Codes = 84
	CodeKeypadAsterisk    Codes = 85 // *
	CodeKeypadHyphenMinus Codes = 86
	CodeKeypadPlusSign    Codes = 87 // +
	CodeKeypadEnter       Codes = 88
	CodeKeypad1           Codes = 89
	CodeKeypad2           Codes = 90
	CodeKeypad3           Codes = 91
	CodeKeypad4           Codes = 92
	CodeKeypad5           Codes = 93
	CodeKeypad6           Codes = 94
	CodeKeypad7           Codes = 95
	CodeKeypad8           Codes = 96
	CodeKeypad9           Codes = 97
	CodeKeypad0           Codes = 98
	CodeKeypadFullStop    Codes = 99
	CodeKeypadEqualSign   Codes = 103

	CodeF13 Codes = 104
	CodeF14 Codes = 105
	CodeF15 Codes = 106
	CodeF16 Codes = 107
	CodeF17 Codes = 108
	CodeF18 Codes = 109
	CodeF19 Codes = 110
	CodeF20 Codes = 111
	CodeF21 Codes = 112
	CodeF22 Codes = 113
	CodeF23 Codes = 114
	CodeF24 Codes = 115

	CodeHelp Codes = 117

	CodeMute       Codes = 127
	CodeVolumeUp   Codes = 128
	CodeVolumeDown Codes = 129

	CodeLeftControl  Codes = 224
	CodeLeftShift    Codes = 225
	CodeLeftAlt      Codes = 226
	CodeLeftMeta     Codes = 227 // Command on mac, win key on windows, ? on linux
	CodeRightControl Codes = 228
	CodeRightShift   Codes = 229
	CodeRightAlt     Codes = 230
	CodeRightMeta    Codes = 231

	// The following codes are not part of the standard USB HID Usage IDs for
	// keyboards. See http://www.usb.org/developers/hidpage/Hut1_12v2.pdf
	//
	// Usage IDs are uint16s, so these non-standard values start at 0x10000.

	// CodeCompose is the Code for a compose key, sometimes called a multi key,
	// used to input non-ASCII characters such as Ã± being composed of n and ~.
	//
	// See https://en.wikipedia.org/wiki/Compose_key
	CodeCompose Codes = 0x10000
)

// TODO: currently have to use official go stringer for this,
// not custom one, which doesn't currently handle discontinuities
// // go: generate stringer -type=Codes

// TODO: Given we use runes outside the unicode space, should we provide a
// printing function? Related: it's a little unfortunate that printing a
// events.Key with %v gives not very readable output like:
//	{100 7 key.Modifiers() Press}

var CodeRuneMap = map[Codes]rune{
	CodeA: 'A',
	CodeB: 'B',
	CodeC: 'C',
	CodeD: 'D',
	CodeE: 'E',
	CodeF: 'F',
	CodeG: 'G',
	CodeH: 'H',
	CodeI: 'I',
	CodeJ: 'J',
	CodeK: 'K',
	CodeL: 'L',
	CodeM: 'M',
	CodeN: 'N',
	CodeO: 'O',
	CodeP: 'P',
	CodeQ: 'Q',
	CodeR: 'R',
	CodeS: 'S',
	CodeT: 'T',
	CodeU: 'U',
	CodeV: 'V',
	CodeW: 'W',
	CodeX: 'X',
	CodeY: 'Y',
	CodeZ: 'Z',

	Code1: '1',
	Code2: '2',
	Code3: '3',
	Code4: '4',
	Code5: '5',
	Code6: '6',
	Code7: '7',
	Code8: '8',
	Code9: '9',
	Code0: '0',

	CodeTab:                '\t',
	CodeSpacebar:           ' ',
	CodeHyphenMinus:        '-',
	CodeEqualSign:          '=',
	CodeLeftSquareBracket:  '[',
	CodeRightSquareBracket: ']',
	CodeBackslash:          '\\',
	CodeSemicolon:          ';',
	CodeApostrophe:         '\'',
	CodeGraveAccent:        '`',
	CodeComma:              ',',
	CodeFullStop:           '.',
	CodeSlash:              '/',

	CodeKeypadSlash:       '/',
	CodeKeypadAsterisk:    '*',
	CodeKeypadHyphenMinus: '-',
	CodeKeypadPlusSign:    '+',
	CodeKeypadFullStop:    '.',
	CodeKeypadEqualSign:   '=',
}
