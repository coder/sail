// Package browserapp provides compatibility layer for opening URLs in an
// electron-style browser wrapper.
package browserapp

import (
	"go.coder.com/sail/internal/nohup"
)

// Open opens a URL via the local preferred browser.
func Open(u string) error {
	// TODO: add
	return nohup.Start("google-chrome", "--disable-plugins", "--app="+u)
}
