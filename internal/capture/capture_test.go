package capture

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/UnitVectorY-Labs/gowebshot/internal/config"
)

func TestCropPNG(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 6, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 6; x++ {
			src.Set(x, y, color.RGBA{R: uint8(x * 10), G: uint8(y * 20), B: 200, A: 255})
		}
	}

	var input bytes.Buffer
	if err := png.Encode(&input, src); err != nil {
		t.Fatalf("encoding input PNG: %v", err)
	}

	out, err := cropPNG(input.Bytes(), config.Config{
		Crop: config.Crop{Top: 1, Bottom: 1, Left: 2, Right: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cropped, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("decoding cropped PNG: %v", err)
	}

	if got := cropped.Bounds().Dx(); got != 3 {
		t.Fatalf("expected cropped width 3, got %d", got)
	}
	if got := cropped.Bounds().Dy(); got != 2 {
		t.Fatalf("expected cropped height 2, got %d", got)
	}

	if got := color.RGBAModel.Convert(cropped.At(0, 0)); got != (color.RGBA{R: 20, G: 20, B: 200, A: 255}) {
		t.Fatalf("unexpected top-left pixel after crop: %#v", got)
	}
}

func TestCropPNGWithoutCropReturnsOriginal(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))

	var input bytes.Buffer
	if err := png.Encode(&input, src); err != nil {
		t.Fatalf("encoding input PNG: %v", err)
	}

	out, err := cropPNG(input.Bytes(), config.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(out, input.Bytes()) {
		t.Fatal("expected PNG bytes to be returned unchanged")
	}
}

func TestCropPNGRejectsOutOfBoundsCrop(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))

	var input bytes.Buffer
	if err := png.Encode(&input, src); err != nil {
		t.Fatalf("encoding input PNG: %v", err)
	}

	_, err := cropPNG(input.Bytes(), config.Config{
		Crop: config.Crop{Left: 3},
	})
	if err == nil || err.Error() != "crop exceeded screenshot bounds" {
		t.Fatalf("expected bounds error, got %v", err)
	}
}
