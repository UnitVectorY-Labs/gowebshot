package capture

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/UnitVectorY-Labs/gowebshot/internal/config"
)

func TestCDPSenderJSON(t *testing.T) {
	tests := []struct {
		name, response, want string
		wantErr              bool
	}{
		{"result", `{"id":1,"result":{"data":"ok"}}`, `{"data":"ok"}`, false},
		{"error", `{"id":1,"error":{"message":"failed"}}`, "CDP error", true},
		{"duplicate name", `{"id":1,"id":1,"result":{}}`, "receiving response", true},
		{"duplicate nested name", `{"id":1,"result":{"data":"a","data":"b"}}`, "receiving response", true},
		{"case-sensitive ID", `{"ID":1,"result":{}}`, "receiving response", true},
		{"invalid UTF-8", "{\"id\":1,\"result\":{\"data\":\"\xff\"}}", "receiving response", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverErr := make(chan error, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
				if err != nil {
					serverErr <- err
					return
				}
				defer ws.Close()
				_, request, err := ws.ReadMessage()
				if err == nil {
					var got map[string]any
					err = json.Unmarshal(request, &got)
					_, hasParams := got["params"]
					if err == nil && (got["id"] != float64(1) || got["method"] != "Page.enable" || hasParams) {
						err = fmt.Errorf("unexpected request: %s", request)
					}
				}
				if err == nil {
					err = ws.WriteMessage(websocket.TextMessage, []byte(tt.response))
				}
				serverErr <- err
			}))
			defer server.Close()
			ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer ws.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			result, err := (&cdpSender{ws: ws}).call(ctx, "Page.enable", nil)
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.want) {
					t.Fatalf("expected error containing %q, got %v", tt.want, err)
				}
			} else if err != nil || string(result) != tt.want {
				t.Fatalf("got result %s, error %v; want %s", result, err, tt.want)
			}
			if err := <-serverErr; err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCropPNG(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 6, 4))
	for y := range 4 {
		for x := range 6 {
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
