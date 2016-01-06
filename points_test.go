package points

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ktye/points/palette"
)

func TestPoints(t *testing.T) {
	img := Image{
		Zoom:     11,
		Center:   Coord{2250678794, 1459106335},
		Blur:     0,
		Contrast: 8000, // 1<<16 - 1,
		Palette:  palette.Palette("viridis"),
		Img:      image.NewRGBA(image.Rect(0, 0, 512, 512)),
	}
	if c, err := LoadBinary32("data/points.xy"); err != nil {
		t.Fatal(err)
	} else {
		img.Coords = c
	}
	img.Draw()

	var buf bytes.Buffer
	if err := png.Encode(&buf, img.Img); err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile("out.png", buf.Bytes(), 0666); err != nil {
		t.Fatal(err)
	}
}

// LoadBinary32 reads coordinates stored in binary form from a file.
func LoadBinary32(filename string) ([]Coord, error) {
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
	coords := make([]Coord, size/8)
	if err := binary.Read(f, binary.LittleEndian, &coords); err != nil {
		return nil, err
	}
	return coords, nil
}
