package linux

import "path/filepath"

// HomeDir returns the home directory for a Linux user.
func HomeDir(username string) string {
	if username == "root"  {
		return "/root"
	}
	return filepath.Join("/home", username)
}