// +build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

func main() {
	sailJS, err := ioutil.ReadFile("sail.js")
	if err != nil {
		log.Fatalf("failed to read sail.js: %v", err)
	}

	err = ioutil.WriteFile("sail.js.go", genFile(sailJS), 0644)
	if err != nil {
		log.Fatalf("failed to write sail.js: %v", err)
	}
}

func genFile(sailJS []byte) []byte {
	tmpl := `package main

//go:generate go run sail.js_gen.go
const sailJS = %q
`

	return []byte(fmt.Sprintf(tmpl, sailJS))
}
