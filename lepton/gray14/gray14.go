// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package gray14 implements functions specifically to manipulate image.Gray16
// containing a 14 bits intensity image.
package gray14

import (
	"fmt"
	"image"
	"image/color"
)

// Min returns the lowest intensity pixel of the image.
func Min(i *image.Gray16) uint16 {
	out := uint16(0xffff)
	b := i.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if j := i.Gray16At(x, y).Y; j < out {
				out = j
			}
		}
	}
	return out
}

// Max returns the highest intensity pixel of the image.
func Max(i *image.Gray16) uint16 {
	out := uint16(0)
	b := i.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if j := i.Gray16At(x, y).Y; j > out {
				out = j
			}
		}
	}
	return out
}

// Diff encodes the difference in the image as a 8 bit image centered at 128.
func Diff(a, b *image.Gray16) *image.Gray {
	bounds := a.Bounds()
	if bounds != b.Bounds() {
		return nil
	}
	dst := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		base := y * bounds.Dx()
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			i := int(a.Gray16At(x, y).Y) - int(b.Gray16At(x, y).Y)
			if i > 127 {
				i = 127
			} else if i < -128 {
				i = -128
			}
			dst.Pix[base+x] = uint8(i + 128)
		}
	}
	return dst
}

// AGCLinear reduces the dynamic range of a 14 bits down to 8 bits very naively
// without gamma.
func AGCLinear(i *image.Gray16) *image.Gray {
	b := i.Bounds()
	dst := image.NewGray(b)
	floor := Min(i)
	delta := int(Max(i) - floor)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		base := y * b.Dx()
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Pix[base+x] = uint8(int(i.Gray16At(x, y).Y-floor) * 255 / delta)
		}
	}
	return dst
}

// ToRGB converts the image into a RGB with pseudo-colors.
//
// Uses 9bits long palette (512) centered around 8192 for a total range of 512.
// TODO(maruel): Confirm it's real.
// With room temperature of 20C° and precision per unit of 0.025°K, range is
// 512*0.025 = 12.8. (?)
func ToRGB(intensity uint16) color.NRGBA {
	// Range is [-255, 255].
	i := (int(intensity) - 8192)
	if i < 0 {
		// Use gray scale, further cut the precision by 33% to scale [0, 171].
		if i <= -256 {
			i = -255
		}
		y := uint8((255 - i + 2) * 2 / 3)
		cb := uint8(0)
		cr := uint8(0)
		r, g, b := color.YCbCrToRGB(y, cb, cr)
		return color.NRGBA{r, g, b, 255}
	}
	// Use color. The palette slowly saturates then circle on the hue then
	// increases brightness.
	if i > 256 {
		i = 255
	}
	const base = 255 - (255+2)*2/3
	// Slowly increase brightness.
	y := uint8((i+2)/3 + base)
	cb := uint8(i - 255)
	cr := uint8(255 - i)
	r, g, b := color.YCbCrToRGB(y, cb, cr)
	return color.NRGBA{r, g, b, 255}
}

// PseudoColor reduces the dynamic range of a 14 bits down to RGB. It doesn't
// apply AGC.
func PseudoColor(i *image.Gray16) *image.NRGBA {
	b := i.Bounds()
	dst := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.SetNRGBA(x, y, ToRGB(i.Gray16At(x, y).Y))
		}
	}
	return dst
}

// Equal returns true if the two frames are equal.
func Equal(a, b *image.Gray16) bool {
	bounds := a.Bounds()
	if bounds != b.Bounds() {
		return false
	}
	for i := 0; i < len(a.Pix); i++ {
		if a.Pix[i] != b.Pix[i] {
			return false
		}
	}
	return true
}

// PaletteGray returns a gray palette.
func PaletteGray(vertical bool) *image.Gray {
	x, y := 256, 1
	if vertical {
		x, y = y, x
	}
	dst := image.NewGray(image.Rect(0, 0, x, y))
	for x := 0; x < 256; x++ {
		dst.Pix[x] = uint8(x)
	}
	return dst
}

// PaletteRGB returns the default colorful palette.
func PaletteRGB(vertical bool) *image.NRGBA {
	x, y := 256, 1
	if vertical {
		x, y = y, x
	}
	dst := image.NewNRGBA(image.Rect(0, 0, x, y))
	for x := 0; x < 256; x++ {
		dstBase := 4 * x
		palBase := 3 * x
		dst.Pix[dstBase] = palette[palBase]
		dst.Pix[dstBase+1] = palette[palBase+1]
		dst.Pix[dstBase+2] = palette[palBase+2]
		dst.Pix[dstBase+3] = 255
	}
	return dst
}

// Private details.

var palette = []uint8{
	255, 255, 255, 253, 253, 253, 251, 251, 251, 249, 249, 249, 247, 247, 247,
	245, 245, 245, 243, 243, 243, 241, 241, 241, 239, 239, 239, 237, 237, 237,
	235, 235, 235, 233, 233, 233, 231, 231, 231, 229, 229, 229, 227, 227, 227,
	225, 225, 225, 223, 223, 223, 221, 221, 221, 219, 219, 219, 217, 217, 217,
	215, 215, 215, 213, 213, 213, 211, 211, 211, 209, 209, 209, 207, 207, 207,
	205, 205, 205, 203, 203, 203, 201, 201, 201, 199, 199, 199, 197, 197, 197,
	195, 195, 195, 193, 193, 193, 191, 191, 191, 189, 189, 189, 187, 187, 187,
	185, 185, 185, 183, 183, 183, 181, 181, 181, 179, 179, 179, 177, 177, 177,
	175, 175, 175, 173, 173, 173, 171, 171, 171, 169, 169, 169, 167, 167, 167,
	165, 165, 165, 163, 163, 163, 161, 161, 161, 159, 159, 159, 157, 157, 157,
	155, 155, 155, 153, 153, 153, 151, 151, 151, 149, 149, 149, 147, 147, 147,
	145, 145, 145, 143, 143, 143, 141, 141, 141, 139, 139, 139, 137, 137, 137,
	135, 135, 135, 133, 133, 133, 131, 131, 131, 129, 129, 129, 126, 126, 126,
	124, 124, 124, 122, 122, 122, 120, 120, 120, 118, 118, 118, 116, 116, 116,
	114, 114, 114, 112, 112, 112, 110, 110, 110, 108, 108, 108, 106, 106, 106,
	104, 104, 104, 102, 102, 102, 100, 100, 100, 98, 98, 98, 96, 96, 96, 94, 94,
	94, 92, 92, 92, 90, 90, 90, 88, 88, 88, 86, 86, 86, 84, 84, 84, 82, 82, 82,
	80, 80, 80, 78, 78, 78, 76, 76, 76, 74, 74, 74, 72, 72, 72, 70, 70, 70, 68,
	68, 68, 66, 66, 66, 64, 64, 64, 62, 62, 62, 60, 60, 60, 58, 58, 58, 56, 56,
	56, 54, 54, 54, 52, 52, 52, 50, 50, 50, 48, 48, 48, 46, 46, 46, 44, 44, 44,
	42, 42, 42, 40, 40, 40, 38, 38, 38, 36, 36, 36, 34, 34, 34, 32, 32, 32, 30,
	30, 30, 28, 28, 28, 26, 26, 26, 24, 24, 24, 22, 22, 22, 20, 20, 20, 18, 18,
	18, 16, 16, 16, 14, 14, 14, 12, 12, 12, 10, 10, 10, 8, 8, 8, 6, 6, 6, 4, 4,
	4, 2, 2, 2, 0, 0, 0, 0, 0, 9, 2, 0, 16, 4, 0, 24, 6, 0, 31, 8, 0, 38, 10, 0,
	45, 12, 0, 53, 14, 0, 60, 17, 0, 67, 19, 0, 74, 21, 0, 82, 23, 0, 89, 25, 0,
	96, 27, 0, 103, 29, 0, 111, 31, 0, 118, 36, 0, 120, 41, 0, 121, 46, 0, 122,
	51, 0, 123, 56, 0, 124, 61, 0, 125, 66, 0, 126, 71, 0, 127, 76, 1, 128, 81,
	1, 129, 86, 1, 130, 91, 1, 131, 96, 1, 132, 101, 1, 133, 106, 1, 134, 111, 1,
	135, 116, 1, 136, 121, 1, 136, 125, 2, 137, 130, 2, 137, 135, 3, 137, 139, 3,
	138, 144, 3, 138, 149, 4, 138, 153, 4, 139, 158, 5, 139, 163, 5, 139, 167, 5,
	140, 172, 6, 140, 177, 6, 140, 181, 7, 141, 186, 7, 141, 189, 10, 137, 191,
	13, 132, 194, 16, 127, 196, 19, 121, 198, 22, 116, 200, 25, 111, 203, 28,
	106, 205, 31, 101, 207, 34, 95, 209, 37, 90, 212, 40, 85, 214, 43, 80, 216,
	46, 75, 218, 49, 69, 221, 52, 64, 223, 55, 59, 224, 57, 49, 225, 60, 47, 226,
	64, 44, 227, 67, 42, 228, 71, 39, 229, 74, 37, 230, 78, 34, 231, 81, 32, 231,
	85, 29, 232, 88, 27, 233, 92, 24, 234, 95, 22, 235, 99, 19, 236, 102, 17,
	237, 106, 14, 238, 109, 12, 239, 112, 12, 240, 116, 12, 240, 119, 12, 241,
	123, 12, 241, 127, 12, 242, 130, 12, 242, 134, 12, 243, 138, 12, 243, 141,
	13, 244, 145, 13, 244, 149, 13, 245, 152, 13, 245, 156, 13, 246, 160, 13,
	246, 163, 13, 247, 167, 13, 247, 171, 13, 248, 175, 14, 248, 178, 15, 249,
	182, 16, 249, 185, 18, 250, 189, 19, 250, 192, 20, 251, 196, 21, 251, 199,
	22, 252, 203, 23, 252, 206, 24, 253, 210, 25, 253, 213, 27, 254, 217, 28,
	254, 220, 29, 255, 224, 30, 255, 227, 39, 255, 229, 53, 255, 231, 67, 255,
	233, 81, 255, 234, 95, 255, 236, 109, 255, 238, 123, 255, 240, 137, 255, 242,
	151, 255, 244, 165, 255, 246, 179, 255, 248, 193, 255, 249, 207, 255, 251,
	221, 255, 253, 235, 255, 255, 24,
}

func init() {
	if len(palette) != 3*256 {
		panic(fmt.Sprintf("expected %d, got %d", 3*256, len(palette)))
	}
}