// Copyright (c) 2021, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cie

// SRGBLinToXYZ converts sRGB linear into XYZ CIE standard color space
func SRGBLinToXYZ(rl, gl, bl float32) (x, y, z float32) {
	x = 0.41233895*rl + 0.35762064*gl + 0.18051042*bl
	y = 0.2126*rl + 0.7152*gl + 0.0722*bl
	z = 0.01932141*rl + 0.11916382*gl + 0.95034478*bl
	return
}

// XYZToSRGBLin converts XYZ CIE standard color space to sRGB linear
func XYZToSRGBLin(x, y, z float32) (rl, gl, bl float32) {
	rl = 3.2406*x + -1.5372*y + -0.4986*z
	gl = -0.9689*x + 1.8758*y + 0.0415*z
	bl = 0.0557*x + -0.2040*y + 1.0570*z
	return
}

// SRGBToXYZ converts sRGB into XYZ CIE standard color space
func SRGBToXYZ(r, g, b float32) (x, y, z float32) {
	rl, gl, bl := SRGBToLinear(r, g, b)
	x, y, z = SRGBLinToXYZ(rl, gl, bl)
	return
}

// SRGBToXYZ100 converts sRGB into XYZ CIE standard color space
// with 100-base sRGB values -- used for CAM16 but not CAM02
func SRGBToXYZ100(r, g, b float32) (x, y, z float32) {
	rl, gl, bl := SRGB100ToLinear(r, g, b)
	x, y, z = SRGBLinToXYZ(rl, gl, bl)
	return
}

// XYZToSRGB converts XYZ CIE standard color space into sRGB
func XYZToSRGB(x, y, z float32) (r, g, b float32) {
	rl, bl, gl := XYZToSRGBLin(x, y, z)
	r, g, b = SRGBFromLinear(rl, bl, gl)
	return
}

// XYZ100ToSRGB converts XYZ CIE standard color space, 100 base units,
// into sRGB
func XYZ100ToSRGB(x, y, z float32) (r, g, b float32) {
	rl, bl, gl := XYZToSRGBLin(x/100, y/100, z/100)
	r, g, b = SRGBFromLinear(rl, bl, gl)
	return
}

// XYZNormD65 normalizes XZY values relative to the D65 outdoor white light values
func XYZNormD65(x, y, z float32) (xr, yr, zr float32) {
	xr = x / 0.95047
	zr = z / 1.08883
	yr = y
	return
}

// XYZDenormD65 de-normalizes XZY values relative to the D65 outdoor white light values
func XYZDenormD65(x, y, z float32) (xr, yr, zr float32) {
	xr = x * 0.95047
	zr = z * 1.08883
	yr = y
	return
}
