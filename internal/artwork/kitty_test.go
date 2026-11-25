package artwork

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestEncodePNG(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1600, 1200))
	for i := 0; i < len(src.Pix); i += 4 {
		copy(src.Pix[i:], []byte{200, 100, 50, 255})
	}
	path := filepath.Join(t.TempDir(), "art.jpg")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(f, src, nil); err != nil {
		t.Fatal(err)
	}
	f.Close()

	data, err := encodePNG(path)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 800 || b.Dy() != 600 {
		t.Fatalf("size = %v, want 800x600", b)
	}
	got := color.RGBAModel.Convert(img.At(400, 300)).(color.RGBA)
	for _, d := range []int{int(got.R) - 200, int(got.G) - 100, int(got.B) - 50} {
		if d < -4 || d > 4 {
			t.Fatalf("color = %v, want about {200 100 50}", got)
		}
	}

	if _, err := encodePNG(filepath.Join(t.TempDir(), "missing.jpg")); err == nil {
		t.Fatal("no error for a missing file")
	}
}

func TestTransmitSequence(t *testing.T) {
	for _, size := range []int{0, 1, 3072, 3073, 10000} {
		data := bytes.Repeat([]byte{7}, size)
		parts := regexp.MustCompile(`\x1b_G([^;]*);([^\x1b]*)\x1b\\`).FindAllStringSubmatch(transmitSequence(5, data), -1)

		var payload strings.Builder
		for i, p := range parts {
			if strings.Contains(p[1], "i=5") != (i == 0) {
				t.Errorf("size %d, chunk %d: control %q", size, i, p[1])
			}
			if strings.HasSuffix(p[1], "m=0") != (i == len(parts)-1) {
				t.Errorf("size %d, chunk %d: control %q", size, i, p[1])
			}
			if len(p[2]) > chunkSize {
				t.Errorf("size %d, chunk %d: %d bytes", size, i, len(p[2]))
			}
			payload.WriteString(p[2])
		}
		got, err := base64.StdEncoding.DecodeString(payload.String())
		if err != nil || !bytes.Equal(got, data) {
			t.Errorf("size %d: payload does not round-trip (%v)", size, err)
		}
	}
}

func TestCells(t *testing.T) {
	out := Cells(0x010203, 38, 19)
	if w, h := lipgloss.Width(out), lipgloss.Height(out); w != 38 || h != 19 {
		t.Fatalf("size = %dx%d, want 38x19", w, h)
	}
	if n := strings.Count(out, "\x1b[38;2;1;2;3m"); n != 19 {
		t.Fatalf("color set %d times, want 19", n)
	}
}

func TestHasPlaceholders(t *testing.T) {
	for reply, want := range map[string]bool{
		"\x1bP>|ghostty 1.4.0\x1b\\\x1b[?62;22c": true,
		"\x1bP>|kitty(0.40.0)\x1b\\\x1b[?62;c":   true,
		"\x1bP>|WezTerm 2024\x1b\\\x1b[?65;c":    false,
		"\x1bP>|tmux 3.5a\x1b\\\x1b[?1;2c":       false,
		"\x1b[?1;2c":                             false,
		"":                                       false,
	} {
		if got := hasPlaceholders([]byte(reply)); got != want {
			t.Errorf("%q = %v, want %v", reply, got, want)
		}
	}
}

func TestAnswered(t *testing.T) {
	for reply, want := range map[string]bool{
		"\x1b[?62;22c":                        true,
		"\x1b_Gi=31;OK\x1b\\\x1b[?62c":        true,
		"\x1bP>|tmux 3.5\x1b\\":               false,
		"\x1b[?62;22":                         false,
		"\x1b_Gi=31;EINVAL:bad command\x1b\\": false,
	} {
		if got := answered([]byte(reply)); got != want {
			t.Errorf("%q = %v, want %v", reply, got, want)
		}
	}
}
