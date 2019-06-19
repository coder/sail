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

	"go.coder.com/cli"
	"go.coder.com/flog"
	"golang.org/x/xerrors"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
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

type chromeExtInstall struct{}

func (c *chromeExtInstall) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name: "install-for-chrome-ext",
		Desc: `Installs the chrome native message host manifest.
This allows the sail chrome extension to manage sail.`,
	}
}

func (c *chromeExtInstall) Run(fl *flag.FlagSet) {
	nativeHostDirs, err := nativeMessageHostManifestDirectories()
	if err != nil {
		flog.Fatal("failed to get native message host manifest directory: %v", err)
	}

	for _, dir := range nativeHostDirs {
		if dir == "" {
			continue
		}

		err = os.MkdirAll(dir, 0755)
		if err != nil {
			flog.Fatal("failed to ensure manifest directory exists: %v", err)
		}
		err = writeNativeHostManifest(dir)
		if err != nil {
			flog.Fatal("failed to write native messaging host manifest: %v", err)
		}
	}
}

func writeNativeHostManifest(dir string) error {
	binPath, err := os.Executable()
	if err != nil {
		return err
	}

	manifest := fmt.Sprintf(`{
		"name": "com.coder.sail",
		"description": "sail message host",
		"path": "%v",
		"type": "stdio",
		"allowed_origins": [
			"chrome-extension://deeepphleikpinikcbjplcgojfhkcmna/"
		]
	}`, binPath)

	dst := path.Join(dir, "com.coder.sail.json")
	return ioutil.WriteFile(dst, []byte(manifest), 0644)
}

func nativeMessageHostManifestDirectories() ([]string, error) {
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
