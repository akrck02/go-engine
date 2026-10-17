// Package notices holds the licenses' notices of the software a game made with go-engine carries,
// to write next to it: the desktop's (Go, go-renderer, go-ui, GLFW…) and the Nintendo 3DS's
// (TinyGo, libctru, citro3d, newlib…).
package notices

import _ "embed"

//go:embed desktop.md
var Desktop []byte

//go:embed 3ds.md
var Console []byte
