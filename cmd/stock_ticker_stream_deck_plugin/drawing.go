package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"strings"
	"sync"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var (
	white  = &color.RGBA{255, 255, 255, 255}
	orange = &color.RGBA{248, 136, 28, 255}
	green  = &color.RGBA{62, 158, 62, 255}
	red    = &color.RGBA{181, 26, 40, 255}
	blue   = &color.RGBA{61, 117, 164, 255}
	grey   = &color.RGBA{255, 255, 255, 255}
)

// DrawTile renders the tile given context and stock data
func DrawTile(title string, price, change, changePercent float32, status string, statusColor *color.RGBA, arrow string, arrowColor *color.RGBA) *[]byte {
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(width)))
	drawLabel(&Label{
		text:     title,
		fontName: "Muli-ExtraBold.ttf",
		fontSize: 17,
		x:        4,
		y:        25,
		clr:      white,
	}, img)
	drawLabel(&Label{
		text:     status,
		fontName: "icons.ttf",
		fontSize: 15,
		x:        57,
		y:        25,
		clr:      statusColor,
	}, img)
	drawLine(5, 30, 11, 2, &color.RGBA{102, 102, 102, 255}, img)
	if arrow != "" {
		drawLabel(&Label{
			text:     arrow,
			fontName: "icons.ttf",
			fontSize: 15,
			x:        57,
			y:        52,
			clr:      arrowColor,
		}, img)
	}
	drawLabel(&Label{
		text:     fmt.Sprintf("%.2f", price),
		fontName: "Lato-Bold.ttf",
		fontSize: 14,
		x:        4,
		y:        50,
		clr:      white,
	}, img)
	drawLabel(&Label{
		text:     fmt.Sprintf("%.2f %.2f%%", change, changePercent),
		fontName: "Lato-Regular.ttf",
		fontSize: 11,
		x:        4,
		y:        65,
		clr:      white,
	}, img)
	b, err := EncodePNG(img)
	if err != nil {
		log.Fatalf("EncodePNG: %v\n", err)
	}
	return &b
}

const width = 72

type fpair struct {
	fontName string
	fontSize float64
}

// Label struct contains text, position and color information
type Label struct {
	text     string
	y        uint
	x        uint
	fontName string
	fontSize float64
	center   bool
	clr      *color.RGBA
}
type singleshared struct {
	fonts  map[string]*truetype.Font
	faces  map[fpair]font.Face
	pngEnc *png.Encoder
	pngBuf *bytes.Buffer
}

var sharedinstance *singleshared
var once sync.Once

func shared() *singleshared {
	once.Do(func() {
		sharedinstance = &singleshared{
			pngEnc: &png.Encoder{
				CompressionLevel: png.NoCompression,
			},
			pngBuf: bytes.NewBuffer(make([]byte, 0, 15697)),
		}
		sharedinstance.fonts = make(map[string]*truetype.Font)
		sharedinstance.faces = make(map[fpair]font.Face)
	})
	return sharedinstance
}

func (ss *singleshared) face(fontName string, fontSize float64) font.Face {
	if face, ok := ss.faces[fpair{fontName, fontSize}]; ok {
		return face
	}

	font := ss.fonts[fontName]
	if font == nil {
		b, err := ioutil.ReadFile(fontName)
		if err != nil {
			log.Fatal(err)
		}
		font, err = truetype.Parse(b)
		if err != nil {
			log.Fatal("failed to parse font")
		}
		ss.fonts[fontName] = font
	}

	face := truetype.NewFace(font, &truetype.Options{Size: fontSize, DPI: 72})
	ss.faces[fpair{fontName, fontSize}] = face

	return face
}

func unfix(x fixed.Int26_6) float64 {
	const shift, mask = 6, 1<<6 - 1
	if x >= 0 {
		return float64(x>>shift) + float64(x&mask)/64
	}
	x = -x
	if x >= 0 {
		return -(float64(x>>shift) + float64(x&mask)/64)
	}
	return 0
}

func drawLine(x, y, width, height int, c *color.RGBA, img *image.RGBA) {
	maxX := x + width - 1
	maxY := y + height - 1
	for x := x; x < maxX; x++ {
		for y := y; y < maxY; y++ {
			img.Set(x, y, c)
		}
	}
}

func drawLabel(l *Label, img *image.RGBA) {
	shared := shared()
	lines := strings.Split(l.text, "\n")
	curY := l.y
	face := shared.face(l.fontName, l.fontSize)

	for _, line := range lines {
		var lwidth float64
		for _, x := range line {
			awidth, ok := face.GlyphAdvance(rune(x))
			if ok != true {
				log.Println("drawLabel: Failed to GlyphAdvance")
				return
			}
			lwidth += unfix(awidth)
		}

		lx := float64(l.x)
		if l.center {
			lx = (float64(width) / 2.) - (lwidth / 2.)
		}
		point := fixed.Point26_6{X: fixed.Int26_6(lx * 64), Y: fixed.Int26_6(curY * 64)}

		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(l.clr),
			Face: face,
			Dot:  point,
		}
		d.DrawString(line)
		curY += 12
	}
}

// EncodePNG renders the current state of the graph
func EncodePNG(img *image.RGBA) ([]byte, error) {
	bak := append(img.Pix[:0:0], img.Pix...)
	shared := shared()
	err := shared.pngEnc.Encode(shared.pngBuf, img)
	if err != nil {
		return nil, err
	}
	img.Pix = bak
	bts := shared.pngBuf.Bytes()
	shared.pngBuf.Reset()
	return bts, nil
}
