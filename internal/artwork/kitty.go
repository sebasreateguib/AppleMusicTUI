// Package artwork shows real images in terminals that implement the Kitty
// graphics protocol with Unicode placeholders (Kitty and Ghostty).
//
// The image is sent to the terminal once, as PNG. The UI then prints ordinary
// placeholder characters and the terminal paints the image over them, so the
// artwork moves, gets covered and disappears like any other text.
package artwork

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"
	"strings"
	"sync/atomic"
)

const (
	// maxPixels bounds the longest side of the image sent to the terminal.
	maxPixels = 800
	// chunkSize is the largest payload the protocol allows per escape code.
	chunkSize = 4096

	placeholder = '\U0010EEEE'
)

// diacritics encode the row and column of a placeholder cell. They come, in
// order, from Kitty's rowcolumn-diacritics.txt.
var diacritics = []rune{
	0x0305, 0x030D, 0x030E, 0x0310, 0x0312, 0x033D, 0x033E, 0x033F,
	0x0346, 0x034A, 0x034B, 0x034C, 0x0350, 0x0351, 0x0352, 0x0357,
	0x035B, 0x0363, 0x0364, 0x0365, 0x0366, 0x0367, 0x0368, 0x0369,
	0x036A, 0x036B, 0x036C, 0x036D, 0x036E, 0x036F, 0x0483, 0x0484,
	0x0485, 0x0486, 0x0487, 0x0592, 0x0593, 0x0594, 0x0595, 0x0597,
}

var lastID atomic.Uint32

// Upload converts the image at path to PNG, sends it to the terminal and
// returns the ID it can be placed with. It is slow: call it off the UI loop.
func Upload(path string) (uint32, error) {
	data, err := encodePNG(path)
	if err != nil {
		return 0, err
	}
	// IDs travel as a 24-bit color; the PID keeps them apart from the ones
	// used by another instance or a previous run in the same terminal.
	id := uint32(os.Getpid()&0xFFF)<<12 | lastID.Add(1)&0xFFF
	return id, write(transmitSequence(id, data))
}

// Place makes the image fill cols x rows placeholder cells.
func Place(id uint32, cols, rows int) {
	_ = write(fmt.Sprintf("\x1b_Ga=p,U=1,i=%d,p=1,c=%d,r=%d,q=2\x1b\\", id, cols, rows))
}

// Delete frees the image in the terminal.
func Delete(id uint32) {
	if id != 0 {
		_ = write(fmt.Sprintf("\x1b_Ga=d,d=I,i=%d,q=2\x1b\\", id))
	}
}

// Cells returns the placeholder text for an image placed with the same size.
func Cells(id uint32, cols, rows int) string {
	cols = min(cols, len(diacritics))
	rows = min(rows, len(diacritics))

	var sb strings.Builder
	for y := 0; y < rows; y++ {
		if y > 0 {
			sb.WriteByte('\n')
		}
		fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm", id>>16&0xFF, id>>8&0xFF, id&0xFF)
		for x := 0; x < cols; x++ {
			sb.WriteRune(placeholder)
			sb.WriteRune(diacritics[y])
			sb.WriteRune(diacritics[x])
		}
		sb.WriteString("\x1b[39m")
	}
	return sb.String()
}

// write uses a single Write call: os.File serializes those, so a sequence never
// ends up in the middle of a frame drawn by Bubble Tea.
func write(seq string) error {
	_, err := os.Stdout.WriteString(seq)
	return err
}

func transmitSequence(id uint32, data []byte) string {
	payload := base64.StdEncoding.EncodeToString(data)

	var sb strings.Builder
	for first := true; first || payload != ""; first = false {
		n := min(chunkSize, len(payload))
		more := 0
		if n < len(payload) {
			more = 1
		}
		sb.WriteString("\x1b_G")
		if first {
			fmt.Fprintf(&sb, "a=t,f=100,t=d,i=%d,", id)
		}
		fmt.Fprintf(&sb, "q=2,m=%d;%s\x1b\\", more, payload[:n])
		payload = payload[n:]
	}
	return sb.String()
}

func encodePNG(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := enc.Encode(&buf, shrink(img, maxPixels)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// shrink averages pixels so that no side exceeds limit.
func shrink(src image.Image, limit int) image.Image {
	b := src.Bounds()
	longest := max(b.Dx(), b.Dy())
	if longest <= limit {
		return src
	}

	w := max(1, b.Dx()*limit/longest)
	h := max(1, b.Dy()*limit/longest)
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		y0, y1 := b.Min.Y+y*b.Dy()/h, b.Min.Y+(y+1)*b.Dy()/h
		for x := 0; x < w; x++ {
			x0, x1 := b.Min.X+x*b.Dx()/w, b.Min.X+(x+1)*b.Dx()/w

			var r, g, bl, a, n uint32
			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					pr, pg, pb, pa := src.At(sx, sy).RGBA()
					r, g, bl, a, n = r+pr>>8, g+pg>>8, bl+pb>>8, a+pa>>8, n+1
				}
			}
			if n > 0 {
				dst.SetRGBA(x, y, color.RGBA{uint8(r / n), uint8(g / n), uint8(bl / n), uint8(a / n)})
			}
		}
	}
	return dst
}
