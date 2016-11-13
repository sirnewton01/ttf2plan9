package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

var (
	dpi      = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile = flag.String("fontfile", "anonymous-pro.ttf", "filename of the ttf font")
	size     = flag.Float64("size", 14, "font size in points")
)

func main() {
	flag.Parse()

	// Read the font data.
	fontBytes, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		log.Println(err)
		return
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}
	ff := truetype.NewFace(f, &truetype.Options{Size: *size, DPI: *dpi, Hinting: font.HintingFull})

	// Scan all of the glyphs with the current scale in order
	//  to figure out widths
	sizepx := ff.Metrics().Ascent.Ceil() + ff.Metrics().Descent.Ceil()
	ascentpx := ff.Metrics().Ascent.Ceil()
	widthpx := 0

	type Fontchar struct {
		x      int16
		top    uint8
		bottom uint8
		left   int8
		width  uint8
	}

	fontchars := []Fontchar{}

	x := int16(0)
	for i := 0; i < 127; i++ {
		// TODO handle missing glyphs
		_, advance, _ := ff.GlyphBounds(rune(i))
		wide := advance.Ceil()
		widthpx += wide

		fc := Fontchar{}
		fc.x = x
		fc.top = uint8(0)
		fc.bottom = uint8(sizepx)
		fc.left = int8(0) // Is this really correct, even for fixed width fonts?
		fc.width = uint8(wide)

		x += int16(wide)

		fontchars = append(fontchars, fc)
	}
	fontchars = append(fontchars, Fontchar{x: x})

	// Initialize the context.
	fg, bg := image.White, image.Black

	// TODO take the size from the size chosen by the user
	rgba := image.NewGray(image.Rect(0, 0, widthpx, sizepx))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	c := freetype.NewContext()
	c.SetDPI(*dpi)
	c.SetFont(f)
	c.SetFontSize(*size)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	c.SetHinting(font.HintingFull)

	// Draw the text.
	pt := freetype.Pt(0, ascentpx)
	for i := 0; i < 127; i++ {
		pt, err = c.DrawString(string(i), pt)
		if err != nil {
			log.Println(err)
			return
		}
	}

	// Save that RGBA image to disk.
	outFile, err := os.Create("out.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer outFile.Close()
	b := bufio.NewWriter(outFile)
	err = png.Encode(b, rgba)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = b.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("Wrote out.png OK.")

	// Save the image in plan9 image format.
	outFile, err = os.Create("R." + strconv.Itoa(int(*size)) + ".1")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer outFile.Close()
	buf := bufio.NewWriter(outFile)
	r := rgba.Bounds()
	cm := color.GrayModel
	fmt.Fprintf(buf, "%11s %11d %11d %11d %11d ", "k8", 0, 0, r.Max.X, r.Max.Y)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			colr := cm.Convert(rgba.At(x, y))
			gcolr := colr.(color.Gray)
			buf.Write([]byte{gcolr.Y})
		}
	}
	fmt.Fprintf(buf, "%11d %11d %11d ", 127, sizepx, ascentpx)
	for _, fc := range fontchars {
		fmt.Printf("%+v\n", fc)
		buf.Write([]byte{byte(fc.x), byte(fc.x >> 8), byte(fc.top), byte(fc.bottom), byte(fc.left), byte(fc.width)})
	}
	err = buf.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	outFile, err = os.Create("R." + strconv.Itoa(int(*size)) + ".font")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer outFile.Close()
	buf = bufio.NewWriter(outFile)
	fmt.Fprintf(buf, "%d %d\n", sizepx, ascentpx)
	fmt.Fprintf(buf, "0x0000 0x007F R.%v.1\n", strconv.Itoa(int(*size)))
	buf.Flush()
}
