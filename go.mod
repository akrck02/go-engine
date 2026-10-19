module github.com/akrck02/go-engine

go 1.25.6

require (
	github.com/akrck02/go-renderer v0.0.0
	github.com/akrck02/go-scene-builder v0.0.0-20261017184317-26708cf29060
	github.com/akrck02/go-ui v0.0.0-20261019173702-0c53a25e3cf4
)

require (
	github.com/go-gl/gl v0.0.0-20260331235117-4566fea9a276 // indirect
	github.com/go-gl/glfw/v3.4/glfw v0.1.0-pre.1 // indirect
	golang.org/x/image v0.45.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
)

// go-ui asks for go-renderer v0.0.0 (it builds against a local copy); this is go-renderer's
// dev branch, which draws on the desktop, Android and the Nintendo 3DS alike, as go-ui's dev.
replace github.com/akrck02/go-renderer => github.com/akrck02/go-renderer v0.0.0-20261019173110-5a5d4410828b
