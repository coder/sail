package codeserver

import (
	"fmt"
)

// CodeServerVersion stores the version of code-server to use.
// TODO (Dean): move this to build steps
const CodeServerVersion = "2.1523-vsc1.38.1"

// DownloadURL gets a download URL for the specified version of code-server.
func DownloadURL(version string) string {
	return fmt.Sprintf("https://codesrv-ci.cdr.sh/releases/%v/linux-x86_64/code-server", version)
}
