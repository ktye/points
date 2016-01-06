// Program conv converts binary gps data between float32 and uint32 coordinates.
//
// usage: conv input.bin output.bin
package main

import (
	"encoding/binary"
	"log"
	"math"
	"os"
)

// FloatCoords stores GPS coordinates as float32.
// Latitude: [-90, 90]
// Longitude: [-180, 180]
type FloatCoords struct {
	Lat, Lon float32
}

// UintCoords stores coordinates in Uint32.
// The coordinates are transformed to a carthesian coordinate system, using
// the mercator projection, with the origin (X,Y) = (0,0) at (lat,lon) = (90,-180)
// and the max value (X,Y) = (UINT32MAX, UINT32MAX) at (lat, lon) = (-90,180).
// So the Y coordinates is pointing southwards.
type UintCoords struct {
	X, Y uint32
}

// Tansform does the mercator transformation.
func (f FloatCoords) Transform() (u UintCoords) {
	max := float64(1<<32 - 1)
	u.X = uint32((float64(f.Lon) + 180.0) * max / 360.0)
	u.Y = uint32((math.Pi - math.Log(math.Tan(math.Pi/4.0+float64(f.Lat)*math.Pi/360.0))) * max / 2.0 / math.Pi)
	return u
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("expecting 2 arguments")
	}

	var err error
	var in *os.File
	if in, err = os.Open(os.Args[1]); err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	var out *os.File
	if out, err = os.Create(os.Args[2]); err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	var f FloatCoords
	for {
		if err := binary.Read(in, binary.LittleEndian, &f); err != nil {
			break
		}
		u := f.Transform()
		// fmt.Printf("%v => %v\n", f, u)
		if err := binary.Write(out, binary.LittleEndian, u); err != nil {
			log.Fatal(err)
		}
	}
}
