package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

// NativeMessage represents a message sent over stdio
type NativeMessage struct {
	Type string `json:"type,omitempty"`

	// Client events
	RunEvent *runEvent `json:"run_event,omitempty"`

	// Server events
	ListEvent  *listEvent  `json:"list_event,omitempty"`
	ErrorEvent *errorEvent `json:"error_event,omitempty"`
}

type runEvent struct {
	Repo string `json:"repo,omitempty"`
}

type listEvent struct {
	Projects []projectInfo `json:"projects,omitempty"`
}

type errorEvent struct {
	Error string `json:"error,omitempty"`
}

func handleNativeMessaging() {
	writeNativeMessage(os.Stdout, &NativeMessage{
		Type: "active",
	})

	reader := bufio.NewReader(os.Stdin)
	buf := make([]byte, 512)
	for {
		_, err := reader.Read(buf)
		if err == io.EOF {
			return
		}
		nm, err := readNativeMessage(bytes.NewReader(buf))
		if err != nil {
			continue
		}

		setError := func(message string) {
			nm = &NativeMessage{
				Type: "error",
				ErrorEvent: &errorEvent{
					Error: message,
				},
			}
		}

		switch nm.Type {
		case "list":
			projects, err := listProjects()
			if err != nil {
				setError(err.Error())
				break
			}

			nm = &NativeMessage{
				Type: "list",
				ListEvent: &listEvent{
					Projects: projects,
				},
			}
			break
		case "run":
			cmd := exec.Command(os.Args[0], "run", nm.RunEvent.Repo)
			err = cmd.Run()
			if err != nil {
				setError(err.Error())
				break
			}

			nm = &NativeMessage{
				Type: "success",
			}
			break
		default:
			setError("unkown event type: " + nm.Type)
			break
		}

		err = writeNativeMessage(os.Stdout, nm)
		if err != nil {
			continue
		}
	}
}

// WriteNativeChromeManifest will write the native manifest
func WriteNativeChromeManifest() error {
	nativeHostDir, err := NativeMessagingDirectory()
	if err != nil {
		return err
	}
	err = os.MkdirAll(nativeHostDir, os.ModePerm)
	if err != nil {
		return err
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path.Join(nativeHostDir, "com.coder.sail.json"), []byte(fmt.Sprintf(`{
	"name": "com.coder.sail",
	"description": "Example app",
	"path": "%s",
	"type": "stdio",
	"allowed_origins": [
		"chrome-extension://emhbehmeaolbmnbafkmdmiedpfhaabjg/"
	]
}`, path.Join(homeDir, "go", "bin", "sail"))), 0644)
}

// NativeMessagingDirectory returns the directory to place the native messaging host manifest within
func NativeMessagingDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(homeDir, ".config", "google-chrome", "NativeMessagingHosts"), nil
}

func readNativeMessage(reader io.Reader) (*NativeMessage, error) {
	var header struct {
		Length uint32
	}
	err := binary.Read(reader, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}
	rawJSON := make([]byte, header.Length)
	_, err = reader.Read(rawJSON)
	if err != nil {
		return nil, err
	}
	m := &NativeMessage{}
	return m, json.Unmarshal(rawJSON, m)
}

func writeNativeMessage(writer io.Writer, nm *NativeMessage) error {
	content, err := json.Marshal(nm)
	if err != nil {
		return err
	}

	f := bufio.NewWriter(os.Stdout)
	err = binary.Write(f, binary.LittleEndian, uint32(len(content)))
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	if err != nil {
		return err
	}
	return f.Flush()
}
