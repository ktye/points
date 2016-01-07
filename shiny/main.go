// Frontend application for points based on shiny
//
// This is an alternative to the lxn/walk based ../gui
//
// The program assumes the file ../data/points.xy to be present.
// See ../conv/conv.go on the data definition.
//
// Usage: go run main.go
package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"log"
	"os"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"

	"github.com/ktye/points"
	"github.com/ktye/points/palette"
)

func usage() {
	fmt.Println(`Usage:
Left Mouse Click: Center at point and zoom in
Right Mouse Click: Center at point and zoom out
Mouse Click - Move - Release: Pan
Mouse Wheel Up: Increase Contrast
Mouse Wheel Down: Decrease Contrast
Mouse Wheel Up/Down + Shift: Switch color palette
Esc: Exit
`)
}

func Draw(b screen.Buffer, w screen.Window) {
	w.Upload(image.Point{}, b, b.Bounds())
}

func main() {
	var err error
	d := points.Image{
		Zoom:     19,
		Center:   points.Coord{2250678794, 1459106335},
		Blur:     0,
		Contrast: 1<<16 - 1,
	}
	if d.Coords, err = LoadBinary32("../data/points.xy"); err != nil {
		log.Fatal(err)
	}
	usage()

	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer w.Release()

		winSize := image.Point{512, 512}
		b, err := s.NewBuffer(winSize)
		if err != nil {
			log.Fatal(err)
		}
		defer b.Release()

		var downState struct {
			x, y int
			b    mouse.Button
		}
		palettes := palette.GetAllPalettes()
		paletteNumber := 0

		var sz size.Event
		var me mouse.Event
		for e := range w.Events() {
			switch e := e.(type) {
			case key.Event:
				if e.Code == key.CodeEscape {
					return
				}

			case mouse.Event:
				me = e
				x, y := int(me.X), int(me.Y)
				if me.Button == mouse.ButtonWheelUp {
					if me.Modifiers == key.ModShift {
						if paletteNumber++; paletteNumber == len(palettes) {
							paletteNumber = 0
						}
					} else {
						if d.Contrast <<= 1; d.Contrast > 1<<16-1 {
							d.Contrast = 1<<16 - 1
						}
					}
				} else if me.Button == mouse.ButtonWheelDown {
					if me.Modifiers == key.ModShift {
						if paletteNumber--; paletteNumber == -1 {
							paletteNumber = len(palettes) - 1
						}
					} else {
						d.Contrast >>= 1
					}
				} else if me.Direction == mouse.DirNone {
					break
				} else if me.Direction == mouse.DirPress {
					downState.x = x
					downState.y = y
					downState.b = me.Button
					break
				} else if me.Direction == mouse.DirRelease {
					if downState.x == x && downState.y == y {
						d.SetCenter(x, y)
						if downState.b == mouse.ButtonLeft {
							d.Scale(true)
						} else if downState.b == mouse.ButtonRight {
							d.Scale(false)
						}
					} else {
						dx := downState.x - x
						dy := downState.y - y
						d.Move(dx, dy)
					}
				}
				fmt.Println("Center", d.Center, "Zoom", d.Zoom, "Contrast", d.Contrast, "Palette", d.Palette)
				d.Palette = palette.Palette(palettes[paletteNumber])
				d.Draw()
				Draw(b, w)
				w.Publish()

			case paint.Event:
				Draw(b, w)
				w.Publish()

			case size.Event:
				sz = e
				b, err = s.NewBuffer(image.Point{sz.WidthPx, sz.HeightPx})
				if err != nil {
					log.Fatal(err)
				}
				d.Img = b.RGBA()
				d.Draw()

			case error:
				log.Print(e)
			}
		}
	})
}

// LoadBinary32 reads coordinates stored in binary form from a file.
// See conv/conv.go for the data definition.
func LoadBinary32(filename string) ([]points.Coord, error) {
	var f *os.File
	var err error
	if f, err = os.Open(filename); err != nil {
		return nil, err
	}
	defer f.Close()
	var size int64
	if s, err := f.Stat(); err != nil {
		return nil, err
	} else {
		size = s.Size()
	}
	coords := make([]points.Coord, size/8)
	if err := binary.Read(f, binary.LittleEndian, &coords); err != nil {
		return nil, err
	}
	return coords, nil
}
