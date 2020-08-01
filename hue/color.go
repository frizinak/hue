package hue

import (
	"image/color"
	"math"
)

type HSV struct {
	Hue   uint16
	Sat   uint8
	Value uint8
}

type XY struct {
	X, Y  float32
	Value uint8
}

type Temp struct {
	T     uint16
	Value uint8
}

const n = 1<<16 - 1

func ColorTemp(c color.Color) Temp {
	_r, _, _b, _a := c.RGBA()
	r, b, a := float64(_r)/n, float64(_b)/n, float64(_a)/n
	temp := (r - b + 1) / 2
	return Temp{uint16(temp * (1<<9 - 1)), uint8(a * (1<<8 - 1))}
}

func ColorHSV(c color.Color) HSV {
	_r, _g, _b, _a := c.RGBA()
	r, g, b, a := float64(_r)/n, float64(_g)/n, float64(_b)/n, float64(_a)/n
	var h, s, v float64

	min := math.Min(math.Min(r, g), b)
	v = math.Max(math.Max(r, g), b)
	C := v - min

	s = 0.0
	if v != 0.0 {
		s = C / v
	}

	h = 0.0 // We use 0 instead of undefined as in wp.
	if min != v {
		if v == r {
			h = math.Mod((g-b)/C, 6.0)
		}
		if v == g {
			h = (b-r)/C + 2.0
		}
		if v == b {
			h = (r-g)/C + 4.0
		}
		h *= 60.0
		if h < 0.0 {
			h += 360.0
		}
	}

	ret := HSV{
		Hue:   uint16(h / 360 * (1<<16 - 1)),
		Sat:   uint8(s * (1<<8 - 1)),
		Value: uint8(v * a * (1<<8 - 1)),
	}

	return ret
}

func limit(n float64) float64 {
	if n > 1.0 {
		return 1.0
	}
	if n < 0.0 {
		return n
	}
	return n
}

type ColorProfile [9]float64

var (
	ProfileSRGB = ColorProfile{
		0.412453, 0.35758, 0.180423,
		0.212671, 0.71516, 0.072169,
		0.019334, 0.119193, 0.950227,
	}
	ProfileSRGBD50 = ColorProfile{
		0.4360747, 0.3850649, 0.1430804,
		0.2225045, 0.7168786, 0.0606169,
		0.0139322, 0.0971045, 0.7141733,
	}
	ProfileAdobeRGB = ColorProfile{
		0.5767309, 0.1855540, 0.1881852,
		0.2973769, 0.6273491, 0.0752741,
		0.019334, 0.119193, 0.950227,
	}
	ProfileWideGamut = ColorProfile{
		0.7161046, 0.1009296, 0.1471858,
		0.2581874, 0.7249378, 0.0168748,
		0.0000000, 0.0517813, 0.7734287,
	}
	ProfileWideGamutD50 = ColorProfile{
		1.4628067, -0.1840623, -0.2743606,
		-0.5217933, 1.4472381, 0.0677227,
		0.0349342, -0.0968930, 1.2884099,
	}
)

func ColorXY(c color.Color, profile ColorProfile) XY {
	_r, _g, _b, _a := c.RGBA()
	r, g, b, a := float64(_r)/n, float64(_g)/n, float64(_b)/n, float64(_a)/n
	r, g, b = lin(r), lin(g), lin(b)

	x := limit(profile[0+0]*r + profile[0+1]*g + profile[0+2]*b)
	y := limit(profile[3+0]*r + profile[3+1]*g + profile[3+2]*b)
	z := limit(profile[6+0]*r + profile[6+1]*g + profile[6+2]*b)
	rx := x / (x + y + z)
	ry := y / (x + y + z)
	return XY{float32(rx), float32(ry), uint8(z * a * (1<<8 - 1))}
}

func lin(comp float64) float64 {
	if comp < 0.04045 {
		return comp / 12.92
	}

	return math.Pow((comp+0.055)/1.055, 2.4)
}
