// Gui is the front-end program for points.
//
// For the time being, it's windows-only and is built upon lxn/walk.
// It assumes the data points in ../data/points.xy.
// See ../conv/conv.go how to create them.
//
// Mouse navigation:
// A single left click centers the point and zooms in.
// A single right click centers and zooms out.
// Click-move-release drags the image.
package main

import (
	"encoding/binary"
	"image"
	"log"
	"os"

	"github.com/ktye/points"
	"github.com/ktye/points/palette"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	var err error
	var mw *walk.MainWindow
	var iv *walk.ImageView
	var cb *walk.ComboBox

	d := points.Image{
		Zoom:     23,
		Center:   points.Coord{2250678794, 1459106335},
		Blur:     0,
		Contrast: 1<<16 - 1,
		Img:      image.NewRGBA(image.Rect(0, 0, 256, 256)),
	}
	if d.Coords, err = LoadBinary32("../data/points.xy"); err != nil {
		log.Fatal(err)
	}
	d.Draw()
	bitmap, err := walk.NewBitmapFromImage(d.Img)
	if err != nil {
		log.Fatal(err)
	}

	update := func() {
		bounds := iv.ClientBounds()
		r := d.Img.Rect
		r.Max.X = r.Min.X + bounds.Width
		r.Max.Y = r.Min.Y + bounds.Height
		d.Img = image.NewRGBA(r)
		d.Draw()
		bitmap, err := walk.NewBitmapFromImage(d.Img)
		if err != nil {
			log.Fatal(err)
		}
		iv.SetImage(bitmap)
	}

	var downState struct {
		x, y   int
		button walk.MouseButton
	}

	mouseDown := func(x, y int, button walk.MouseButton) {
		downState.x = x
		downState.y = y
		downState.button = button
	}

	mouseUp := func(x, y int, button walk.MouseButton) {
		if downState.x == x && downState.y == y {
			d.SetCenter(x, y)
			if downState.button == walk.LeftButton {
				d.Scale(true)
			} else {
				d.Scale(false)
			}
			update()
		} else {
			dx := downState.x - x
			dy := downState.y - y
			d.Move(dx, dy)
			update()
		}
	}

	MainWindow{
		AssignTo: &mw,
		Title:    "Points",
		Layout:   VBox{},
		OnSizeChanged: func() {
			if mw != nil {
				update()
			}
		},
		DataBinder: DataBinder{
			DataSource: &d,
			AutoSubmit: true,
		},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					ImageView{
						AssignTo:    &iv,
						MinSize:     Size{256, 256},
						Image:       bitmap,
						OnMouseDown: mouseDown,
						OnMouseUp:   mouseUp,
					},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							PushButton{Text: "Zoom In", OnClicked: func() { d.Scale(true); update() }},
							PushButton{Text: "Zoom out", OnClicked: func() { d.Scale(false); update() }},
							PushButton{Text: "Reset", OnClicked: func() { d.Reset(); update() }},
							Label{Text: "Blur"},
							NumberEdit{
								Value:          Bind("Blur", Range{0, 100}),
								OnValueChanged: update,
							},
							Label{Text: "Contrast"},
							NumberEdit{ // I want a slider here instead!
								Value:          Bind("Contrast", Range{0, 1<<16 - 1}),
								OnValueChanged: update,
							},
							Label{Text: "Colormap"},
							ComboBox{
								AssignTo: &cb,
								Value:    "grey",
								Model:    palette.GetAllPalettes(),
								OnCurrentIndexChanged: func() {
									if cb != nil {
										d.Palette = palette.Palette(cb.Text())
										// update() // crashes on startup...
									}
								},
							},
							VSpacer{},
							PushButton{Text: "Up", OnClicked: func() { d.Pan(0, -1); update() }},
							PushButton{Text: "Down", OnClicked: func() { d.Pan(0, 1); update() }},
						},
					},
				},
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{Text: "Left", OnClicked: func() { d.Pan(-1, 0); update() }},
					PushButton{Text: "Right", OnClicked: func() { d.Pan(1, 0); update() }},
				},
			},
		},
	}.Run()
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
