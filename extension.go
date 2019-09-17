package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"
	"unsafe"

	"golang.org/x/xerrors"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"go.coder.com/cli"
	"go.coder.com/flog"
)

func runNativeMsgHost() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		flog.Fatal("failed to listen: %v", err)
	}
	defer l.Close()

	url := "http://" + l.Addr().String()

	err = writeNativeHostMessage(struct {
		URL string `json:"url"`
	}{url})
	if err != nil {
		flog.Fatal("%v", err)
	}

	m := http.NewServeMux()
	m.HandleFunc("/api/v1/run", handleRun)

	err = http.Serve(l, m)
	flog.Fatal("failed to serve: %v", err)
}

func writeNativeHostMessage(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return xerrors.Errorf("failed to marshal url: %w", err)
	}

	// Converts the length of URL into native byte order.
	msgLen := uint32(len(b))
	msgLenHostByteOrder := *(*[4]byte)(unsafe.Pointer(&msgLen))

	os.Stdout.Write(msgLenHostByteOrder[:])
	os.Stdout.Write(b)

	return nil
}

type runRequest struct {
	Project string `json:"project"`
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "something failed")

	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*5)
	defer cancel()

	var req runRequest
	err = wsjson.Read(ctx, c, &req)
	if err != nil {
		log.Printf("failed to read request: %v\n", err)
		c.Close(websocket.StatusInvalidFramePayloadData, "failed to read")
		return
	}

	if streamRun(ctx, c, "run", req.Project) {
		c.Close(websocket.StatusNormalClosure, "")
	}
}

type installExtHostCmd struct{}

func (c *installExtHostCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name: "install-ext-host",
		Desc: `Installs the native message host manifest into Chrome and Firefox.
This allows the sail extension to manage sail.`,
	}
}

func (c *installExtHostCmd) Run(fl *flag.FlagSet) {
	binPath, err := os.Executable()
	if err != nil {
		flog.Fatal("failed to get sail binary location")
	}

	nativeHostDirsChrome, err := nativeMessageHostManifestDirectoriesChrome()
	if err != nil {
		flog.Fatal("failed to get chrome native message host manifest directory: %v", err)
	}
	err = installManifests(nativeHostDirsChrome, "com.coder.sail.json", chromeManifest(binPath))
	if err != nil {
		flog.Fatal("failed to write chrome manifest files: %v", err)
	}

	nativeHostDirsFirefox, err := nativeMessageHostManifestDirectoriesFirefox()
	if err != nil {
		flog.Fatal("failed to get firefox native message host manifest directory: %v", err)
	}
	err = installManifests(nativeHostDirsFirefox, "com.coder.sail.json", firefoxManifest(binPath))
	if err != nil {
		flog.Fatal("failed to write firefox manifest files: %v", err)
	}

	flog.Info("Successfully installed manifests.")
}

func nativeMessageHostManifestDirectoriesChrome() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, xerrors.Errorf("failed to get user home dir: %w", err)
	}

	var chromeDir string
	var chromeBetaDir string
	var chromeDevDir string
	var chromeCanaryDir string
	var chromiumDir string

	switch runtime.GOOS {
	case "linux":
		chromeDir = path.Join(homeDir, ".config", "google-chrome", "NativeMessagingHosts")
		chromeBetaDir = path.Join(homeDir, ".config", "google-chrome-beta", "NativeMessagingHosts")
		chromeDevDir = path.Join(homeDir, ".config", "google-chrome-unstable", "NativeMessagingHosts")
		chromiumDir = path.Join(homeDir, ".config", "chromium", "NativeMessagingHosts")
	case "darwin":
		chromeDir = path.Join(homeDir, "Library", "Application Support", "Google", "Chrome", "NativeMessagingHosts")
		chromeCanaryDir = path.Join(homeDir, "Library", "Application Support", "Google", "Chrome Canary", "NativeMessagingHosts")
		chromiumDir = path.Join(homeDir, "Library", "Application Support", "Chromium", "NativeMessagingHosts")
	default:
		return nil, xerrors.Errorf("unsupported os %q", runtime.GOOS)
	}

	return []string{
		chromeDir,
		chromiumDir,
		chromeBetaDir,
		chromeDevDir,
		chromeCanaryDir,
	}, nil
}

func chromeManifest(binPath string) string {
	return fmt.Sprintf(`{
		"name": "com.coder.sail",
		"description": "sail message host",
		"path": "%v",
		"type": "stdio",
		"allowed_origins": [
			"chrome-extension://deeepphleikpinikcbjplcgojfhkcmna/"
		]
	}`, binPath)
}

func nativeMessageHostManifestDirectoriesFirefox() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, xerrors.Errorf("failed to get user home dir: %w", err)
	}

	var firefoxDir string

	switch runtime.GOOS {
	case "linux":
		firefoxDir = path.Join(homeDir, ".mozilla", "native-messaging-hosts")
	case "darwin":
		firefoxDir = path.Join(homeDir, "Library", "Application Support", "Mozilla", "NativeMessagingHosts")
	default:
		return nil, xerrors.Errorf("unsupported os %q", runtime.GOOS)
	}

	return []string{
		firefoxDir,
	}, nil
}

func firefoxManifest(binPath string) string {
	return fmt.Sprintf(`{
		"name": "com.coder.sail",
		"description": "sail message host",
		"path": "%v",
		"type": "stdio",
		"allowed_extensions": [
			"sail@coder.com"
		]
	}`, binPath)
}

func installManifests(nativeHostDirs []string, file string, content string) error {
	data := []byte(content)

	for _, dir := range nativeHostDirs {
		if dir == "" {
			continue
		}

		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return xerrors.Errorf("failed to ensure manifest directory exists: %w", err)
		}

		dst := path.Join(dir, file)
		err = ioutil.WriteFile(dst, data, 0644)
		if err != nil {
			return xerrors.Errorf("failed to write native messaging host manifest: %w", err)
		}
	}

	return nil
}

type chromeExtInstallCmd struct{
	cmd *installExtHostCmd
}

func (c *chromeExtInstallCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name: "install-for-chrome-ext",
		Desc: "DEPRECATED: alias of install-ext-host.",
	}
}

func (c *chromeExtInstallCmd) Run(fl *flag.FlagSet) {
	c.cmd.Run(fl)
}
