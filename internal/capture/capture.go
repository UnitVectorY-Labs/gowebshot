package capture

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"

	"github.com/UnitVectorY-Labs/gowebshot/internal/config"
)

// cdpRequest represents a Chrome DevTools Protocol command.
type cdpRequest struct {
	ID     int            `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

// cdpResponse represents a Chrome DevTools Protocol response or event.
type cdpResponse struct {
	ID     int             `json:"id,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  json.RawMessage `json:"error,omitempty"`
}

// cdpSender manages sending CDP commands and reading responses over a WebSocket.
type cdpSender struct {
	ws     *websocket.Conn
	nextID int
}

// call sends a CDP command and blocks until the matching response arrives,
// discarding any interleaved events.
func (s *cdpSender) call(ctx context.Context, method string, params map[string]any) (json.RawMessage, error) {
	s.nextID++
	id := s.nextID

	msg := cdpRequest{
		ID:     id,
		Method: method,
		Params: params,
	}
	if err := s.ws.WriteJSON(msg); err != nil {
		return nil, fmt.Errorf("sending %s: %w", method, err)
	}

	for {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context cancelled waiting for %s: %w", method, ctx.Err())
		}

		if deadline, ok := ctx.Deadline(); ok {
			if err := s.ws.SetReadDeadline(deadline); err != nil {
				return nil, fmt.Errorf("setting read deadline: %w", err)
			}
		}

		var resp cdpResponse
		if err := s.ws.ReadJSON(&resp); err != nil {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("context cancelled waiting for %s: %w", method, ctx.Err())
			}
			return nil, fmt.Errorf("receiving response for %s: %w", method, err)
		}

		if resp.ID != id {
			continue // discard events and responses for other IDs
		}
		if resp.Error != nil {
			return nil, fmt.Errorf("CDP error for %s: %s", method, string(resp.Error))
		}
		return resp.Result, nil
	}
}

// Capture takes a screenshot of the URL described by cfg and returns PNG bytes.
func Capture(cfg config.Config) ([]byte, error) {
	chromePath, err := FindChrome(cfg.ChromePath)
	if err != nil {
		return nil, err
	}

	tmpDir, err := os.MkdirTemp("", "gowebshot-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	port, err := freePort()
	if err != nil {
		return nil, fmt.Errorf("finding free port: %w", err)
	}

	captureWidth := cfg.CaptureWidth()
	captureHeight := cfg.CaptureHeight()

	cmd := exec.Command(chromePath,
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-software-rasterizer",
		"--disable-dev-shm-usage",
		fmt.Sprintf("--remote-debugging-port=%d", port),
		fmt.Sprintf("--user-data-dir=%s", tmpDir),
		fmt.Sprintf("--window-size=%d,%d", captureWidth, captureHeight),
	)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting Chrome: %w", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	debugURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	if err := waitForCDP(ctx, debugURL); err != nil {
		return nil, err
	}

	wsURL, err := getPageWSURL(ctx, debugURL)
	if err != nil {
		return nil, err
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	ws, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("connecting to CDP WebSocket: %w", err)
	}
	defer ws.Close()

	send := &cdpSender{ws: ws}

	if _, err := send.call(ctx, "Page.enable", nil); err != nil {
		return nil, fmt.Errorf("Page.enable: %w", err)
	}
	if _, err := send.call(ctx, "Runtime.enable", nil); err != nil {
		return nil, fmt.Errorf("Runtime.enable: %w", err)
	}

	if _, err := send.call(ctx, "Emulation.setDeviceMetricsOverride", map[string]any{
		"width":             captureWidth,
		"height":            captureHeight,
		"deviceScaleFactor": 1,
		"mobile":            false,
	}); err != nil {
		return nil, fmt.Errorf("Emulation.setDeviceMetricsOverride: %w", err)
	}

	if _, err := send.call(ctx, "Page.navigate", map[string]any{
		"url": cfg.URL,
	}); err != nil {
		return nil, fmt.Errorf("Page.navigate: %w", err)
	}

	if err := waitForPageLoad(ctx, send); err != nil {
		return nil, err
	}

	if cfg.Zoom != 1.0 {
		expr := fmt.Sprintf("document.body.style.zoom = '%g'", cfg.Zoom)
		if _, err := send.call(ctx, "Runtime.evaluate", map[string]any{
			"expression": expr,
		}); err != nil {
			return nil, fmt.Errorf("applying zoom: %w", err)
		}
	}

	if cfg.Scroll > 0 {
		expr := fmt.Sprintf("window.scrollBy(0, %d)", cfg.Scroll)
		if _, err := send.call(ctx, "Runtime.evaluate", map[string]any{
			"expression": expr,
		}); err != nil {
			return nil, fmt.Errorf("applying scroll: %w", err)
		}
	}

	// Allow the page to settle after navigation, zoom, and scroll adjustments.
	if cfg.Delay > 0 {
		time.Sleep(cfg.Delay)
	}

	result, err := send.call(ctx, "Page.captureScreenshot", map[string]any{
		"format": "png",
	})
	if err != nil {
		return nil, fmt.Errorf("capturing screenshot: %w", err)
	}

	var screenshot struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(result, &screenshot); err != nil {
		return nil, fmt.Errorf("parsing screenshot result: %w", err)
	}

	pngBytes, err := base64.StdEncoding.DecodeString(screenshot.Data)
	if err != nil {
		return nil, fmt.Errorf("decoding screenshot data: %w", err)
	}

	pngBytes, err = cropPNG(pngBytes, cfg)
	if err != nil {
		return nil, err
	}

	return pngBytes, nil
}

func cropPNG(pngBytes []byte, cfg config.Config) ([]byte, error) {
	if cfg.Crop.IsZero() {
		return pngBytes, nil
	}

	src, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, fmt.Errorf("decoding PNG for crop: %w", err)
	}

	bounds := src.Bounds()
	cropRect := image.Rect(
		bounds.Min.X+cfg.Crop.Left,
		bounds.Min.Y+cfg.Crop.Top,
		bounds.Max.X-cfg.Crop.Right,
		bounds.Max.Y-cfg.Crop.Bottom,
	)

	if cropRect.Empty() {
		return nil, fmt.Errorf("crop removed the entire image")
	}
	if !cropRect.In(bounds) {
		return nil, fmt.Errorf("crop exceeded screenshot bounds")
	}

	dst := image.NewRGBA(image.Rect(0, 0, cropRect.Dx(), cropRect.Dy()))
	draw.Draw(dst, dst.Bounds(), src, cropRect.Min, draw.Src)

	var out bytes.Buffer
	if err := png.Encode(&out, dst); err != nil {
		return nil, fmt.Errorf("encoding cropped PNG: %w", err)
	}

	return out.Bytes(), nil
}

// freePort asks the OS for an available TCP port on the loopback interface.
func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port, nil
}

// waitForCDP polls the Chrome /json/version endpoint until it responds or
// 15 seconds elapse.
func waitForCDP(ctx context.Context, debugURL string) error {
	versionURL := debugURL + "/json/version"
	deadline := time.After(15 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled waiting for CDP: %w", ctx.Err())
		case <-deadline:
			return fmt.Errorf("timeout waiting for Chrome CDP endpoint")
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, versionURL, nil)
			if err != nil {
				continue
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				continue
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
}

// getPageWSURL fetches /json from the Chrome debug endpoint and returns the
// WebSocket debugger URL of the first page target.
func getPageWSURL(ctx context.Context, debugURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, debugURL+"/json", nil)
	if err != nil {
		return "", fmt.Errorf("creating targets request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching targets: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading targets response: %w", err)
	}

	var targets []struct {
		Type                 string `json:"type"`
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.Unmarshal(body, &targets); err != nil {
		return "", fmt.Errorf("parsing targets: %w", err)
	}

	for _, t := range targets {
		if t.Type == "page" && t.WebSocketDebuggerURL != "" {
			return t.WebSocketDebuggerURL, nil
		}
	}

	return "", fmt.Errorf("no page target found in Chrome debugging targets")
}

// waitForPageLoad polls document.readyState via Runtime.evaluate until it
// equals "complete", or 30 seconds elapse.
func waitForPageLoad(ctx context.Context, send *cdpSender) error {
	deadline := time.After(30 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled waiting for page load: %w", ctx.Err())
		case <-deadline:
			return fmt.Errorf("timeout waiting for page to load")
		case <-ticker.C:
			result, err := send.call(ctx, "Runtime.evaluate", map[string]any{
				"expression": "document.readyState",
			})
			if err != nil {
				continue
			}

			var evalResult struct {
				Result struct {
					Value string `json:"value"`
				} `json:"result"`
			}
			if err := json.Unmarshal(result, &evalResult); err != nil {
				continue
			}

			if evalResult.Result.Value == "complete" {
				return nil
			}
		}
	}
}
