package capture

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// FindChrome locates a Chrome or Chromium binary. If explicit is non-empty it
// is validated directly; otherwise OS-specific candidate paths are probed.
func FindChrome(explicit string) (string, error) {
	if explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("specified Chrome path does not exist: %s", explicit)
		}
		return explicit, nil
	}

	candidates := chromeCandidates()
	for _, c := range candidates {
		if filepath.IsAbs(c) {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
		} else {
			if p, err := exec.LookPath(c); err == nil {
				return p, nil
			}
		}
	}

	return "", fmt.Errorf("could not find Chrome or Chromium; install Chrome or use --chrome to specify the path")
}

func chromeCandidates() []string {
	switch runtime.GOOS {
	case "linux":
		return []string{
			"google-chrome",
			"google-chrome-stable",
			"chromium-browser",
			"chromium",
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium-browser",
			"/usr/bin/chromium",
			"/snap/bin/chromium",
		}
	case "darwin":
		return []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"google-chrome",
			"chromium",
		}
	case "windows":
		candidates := []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		}
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			candidates = append(candidates, filepath.Join(localAppData, `Google\Chrome\Application\chrome.exe`))
		}
		candidates = append(candidates, "chrome")
		return candidates
	default:
		return nil
	}
}
