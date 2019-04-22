// Package browserapp provides compatibility layer for opening URLs in an
// electron-style browser wrapper.
package browserapp

import (
	"os"
	"os/exec"

	"github.com/pkg/browser"
	"go.coder.com/sail/internal/nohup"
)

// Open opens a URL via the local preferred browser.
//
// TODO: move this into a location where sshcode and sail can use this.
func Open(url string) error {
	switch {
	case commandExists("google-chrome"):
		return nohup.Start("google-chrome", chromeOptions(url)...)

	case commandExists("google-chrome-stable"):
		return nohup.Start("google-chrome-stable", chromeOptions(url)...)

	case commandExists("chromium"):
		return nohup.Start("chromium", chromeOptions(url)...)

	case commandExists("chromium-browser"):
		return nohup.Start("chromium-browser", chromeOptions(url)...)

	case pathExists("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"):
		return nohup.Start("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", chromeOptions(url)...)

	default:
		return browser.OpenURL(url)
	}
}

func chromeOptions(url string) []string {
	return []string{"--app=" + url, "--disable-extensions", "--disable-plugins", "--incognito"}
}

// Checks if a command exists locally.
func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func pathExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
