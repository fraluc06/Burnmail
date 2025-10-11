package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetDownloadsDir(t *testing.T) {
	var expectedPath string

	switch runtime.GOOS {
	case "darwin":
		home := os.Getenv("HOME")
		expectedPath = filepath.Join(home, "Downloads")
	case "windows":
		userProfile := os.Getenv("USERPROFILE")
		if userProfile == "" {
			userProfile = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		}
		expectedPath = filepath.Join(userProfile, "Downloads")
	case "linux":
		xdgDownload := os.Getenv("XDG_DOWNLOAD_DIR")
		if xdgDownload != "" {
			expectedPath = xdgDownload
		} else {
			home := os.Getenv("HOME")
			expectedPath = filepath.Join(home, "Downloads")
		}
	}

	if expectedPath == "" {
		t.Skip("Unable to determine expected downloads directory for this platform")
	}

	info, err := os.Stat(expectedPath)
	if err == nil && info.IsDir() {
		t.Logf("Downloads directory exists at: %s", expectedPath)
	} else {
		t.Logf("Downloads directory does not exist, would fallback to current directory")
	}
}
