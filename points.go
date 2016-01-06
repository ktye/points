// Package points draws images of GPS data points.
//
// The package visualizes gps data points as an image.
// On full scale, each data point is drawn as a pixel.
// Depending on how many points lie on a pixel, it may
// be drawn in a brighter color.
// This is controlled by Contrast and the color Palette.
// A blur factor takes neighbour points into account
// resulting in a smearing effect. On the downside
// it slows down the compuation.
//
// This is only the package. An interactive program
// can be found in gui/gui.go
//
// The package is a fairly simple implementation which
// is not optimized for speed.
// It is currently good enough for a some million points,
// with blur == 0.
package points

import (
	"image"
	"image/color"

	"github.com/ktye/points/palette"
)

// Coord stores a pair of GPS coordinates.
// The mercator transformation has already been applied.
// the range for X and Y is 0..1<<32-1, Y points downwards (to the south).
// See conv/conv.go for an example on the transform from
// commonly used latitude/longitude pairs.
//
// There are 32 zoom levels. Zoom level 0 is the full resolution.
// zooming out (applying a zoom level > 0) is done by simple bit-shifting
// of the coordinates: x, y := X >> zoom, Y >> zoom.
// The image limits are for both width and height: (1<<32-1) >> zoom.
type Coord struct {
	X, Y uint32
}

// Image is the main type of the package.
type Image struct {
	Coords   []Coord         // Data points
	Zoom     uint            // Zoom level
	Center   Coord           // Center coordinate
	Blur     int             // Blur factor (number of neighbour points to consider).
	Contrast int             // Contrast, 0..1<<32-1
	Palette  palette.Palette // Color palette
	Img      *image.RGBA     // The actual color image
	gray     *image.Gray16   // Grayscale image
}

// scale is the factor between Gray16 max value
// and the max value of the color palettes.
const scale = 1 << 16 / 256

// Draw fills the image's Img, which should be preallocated.
// The data drawn depends on the image's rectangle.
func (im Image) Draw() {
	// Draw all data points to a gray scale image.
	setPoint := func(x, y int, weight int) {
		if v := im.gray.At(x, y).(color.Gray16).Y; v <= uint16(1<<16-1)-uint16(weight) {
			im.gray.Set(x, y, color.Gray16{v + uint16(weight)})
		} else {
			im.gray.Set(x, y, color.White)
		}
	}
	im.gray = image.NewGray16(im.Img.Bounds())
	var cx, cy uint32
	cx, cy = im.Center.X>>im.Zoom-uint32(im.Img.Rect.Max.X/2), im.Center.Y>>im.Zoom-uint32(im.Img.Rect.Max.Y/2)
	for _, c := range im.Coords {

		x, y := c.X>>im.Zoom-cx, c.Y>>im.Zoom-cy
		r := im.Img.Rect
		if x > uint32(r.Max.X) || y > uint32(r.Max.Y) {
			continue
		}
		if im.Blur > 0 {
			for i := -im.Blur; i <= im.Blur; i++ {
				for k := -im.Blur; k <= im.Blur; k++ {
					setPoint(int(x)+i, int(y)+k, im.Contrast*(2*im.Blur*im.Blur-(i*i+k*k)))
				}
			}
		} else {
			setPoint(int(x), int(y), im.Contrast)
		}
	}

	// Create the color image from gray scale.
	palette := im.Palette.Palette()
	for x := im.Img.Rect.Min.X; x < im.Img.Rect.Max.X; x++ {
		for y := im.Img.Rect.Min.Y; y < im.Img.Rect.Max.Y; y++ {
			idx := int(uint16(im.gray.At(x, y).(color.Gray16).Y) / scale)
			c := palette[idx]
			im.Img.Set(x, y, c)
		}
	}
}

// The following methods can be used by a frontend as callbacks
// to interactive events.
// They only affect Zoom and Center. A redraw operation must
// be done explicitly.

// Reset zoom and translation.
func (im *Image) Reset() {
	im.Zoom = 0
	im.Center.X = (1<<32 - 1) / 2
	im.Center.Y = (1<<32 - 1) / 2
}

// Scale zooms in or out.
func (im *Image) Scale(zoomIn bool) {
	d := 1
	if zoomIn {
		d = -1
	}
	if z := int(im.Zoom) + d; z >= 0 || z < 32 {
		im.Zoom = uint(z)
	}
}

// Pan translates by 1/4th of the current display to left/right or up/down.
func (im *Image) Pan(right int, up int) {
	scale := uint32(1 << im.Zoom)
	if dx := scale * uint32(im.Img.Rect.Dx()/4); right > 0 {
		im.Center.X += dx
	} else if right < 0 {
		im.Center.X -= dx
	}
	if dy := scale * uint32(im.Img.Rect.Dy()/4); up > 0 {
		im.Center.Y += dy
	} else if up < 0 {
		im.Center.Y -= dy
	}
}

// Move translates the current image's center by discrete amounts.
func (im *Image) Move(dx, dy int) {
	scale := uint32(1 << im.Zoom)
	im.Center.X += scale * uint32(dx)
	im.Center.Y += scale * uint32(dy)
}

// SetCenter translates the image, such that x and y move to the center.
func (im *Image) SetCenter(x, y int) {
	scale := uint32(1 << im.Zoom)
	dx := x - im.Img.Rect.Dx()/2
	dy := y - im.Img.Rect.Dy()/2
	im.Center.X += scale * uint32(dx)
	im.Center.Y += scale * uint32(dy)
}
