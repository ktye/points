// Palette offers color palettes for color plots.
//
// A Palette is just an array of colors.
// The interface of this package allows these palettes to be
// built by a simple array, or a function that generates it.
//
// To add a new palette, only one file must be added.
// See inferno.go for an example.
//
// The palettes are taken from matplotlib.
// See also for great video on the subject: www.youtube.com/watch?v=xAoljeRJ3lU
package palette

import (
	"fmt"
	"image/color"
	"sort"
)

// Palette is a string which identifies the palette by name.
// A minus sign before the name inverts the palette.
type Palette string

// generator is a function which can produce a palette.
// number of colors.
type generator func() [256]color.Color

// All known palette generators are stored in this map.
var palettes = map[Palette]generator{
	"":     grey, // default Palette if unset.
	"grey": grey,
}

// grey make a grey-scale palette.
func grey() (g [256]color.Color) {
	for i := 0; i < 256; i++ {
		g[i] = color.Gray{Y: uint8(i)}
	}
	return g
}

func (p Palette) Verify() error {
	if len(p) > 0 && p[0] == '-' {
		p = Palette(string(p)[1:])
	}
	if _, ok := palettes[p]; !ok {
		return fmt.Errorf("unknown palette: '%s'", p)
	}
	return nil
}

// Palette returns a [256]color.Color with any number of colors.
// If the palette name starts with "-" it is inverted.
func (p Palette) Palette() [256]color.Color {
	inv := false
	if len(p) > 0 && p[0] == '-' {
		inv = true
		p = p[1:]
	}
	if f, ok := palettes[p]; !ok {
		return grey()
	} else {
		pal := f()
		if inv {
			var invpal [256]color.Color
			for i := 0; i < 256; i++ {
				invpal[i] = pal[255-i]
			}
			return invpal
		} else {
			return pal
		}
	}
}

// GetAllPalettes returns the names of all palettes that are compiled
// into the package as well as their inverts.
func GetAllPalettes() []string {
	var all []string
	for key := range palettes {
		all = append(all, string(key), "-"+string(key))
	}
	sort.Strings(all)
	return all
}
