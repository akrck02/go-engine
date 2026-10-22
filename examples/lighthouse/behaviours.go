package main

import (
	"math"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/space"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/scene"
)

// Beam is the lamp's light: a cone built in code whose light adds to what it crosses (additive),
// turning every frame, so it is lit on the GPU (moving).
type Beam struct {
	DegreesPerSecond float64 `json:"degreesPerSecond"`
	Length           float64 `json:"length"`
}

// beamRadiusShare is how wide the beam's open end is, for its length.
const beamRadiusShare = 0.18

func (beam *Beam) Start(context *engine.Context) {
	material := &scene.Material{Name: "beam", Kind: scene.UnlitKind, BaseColor: graphics.Vector{1, 0.92, 0.6, 0.55},
		Additive: true, DoubleSided: true}
	context.Node.SetMesh(beamCone(float32(beam.Length), float32(beam.Length*beamRadiusShare), material))
	context.Node.SetMoving(true)
}

func (beam *Beam) Update(context *engine.Context) {
	context.Node.Transform.Turn(space.Up, beam.DegreesPerSecond*context.Delta)
}

// Fog is a translucent disc over the sea whose opacity breathes, changed every frame.
type Fog struct {
	Radius   float64 `json:"radius"`
	material *scene.Material
}

// The fog's opacity: its middle, how far it swings, and how many seconds a breath takes.
const (
	fogOpacity       = 0.25
	fogOpacitySwing  = 0.15
	fogBreathSeconds = 6
)

func (fog *Fog) Start(context *engine.Context) {
	fog.material = &scene.Material{Name: "fog", Kind: scene.UnlitKind, BaseColor: graphics.Vector{0.75, 0.8, 0.85, fogOpacity},
		Transparent: true, DoubleSided: true}
	context.Node.SetMesh(flatDisc(float32(fog.Radius), fog.material))
}

func (fog *Fog) Update(context *engine.Context) {
	phase := 2 * math.Pi * context.Time / fogBreathSeconds
	fog.material.BaseColor[3] = float32(fogOpacity + fogOpacitySwing*math.Sin(phase))
}

// Marker is a buoy's diamond in a colour of its own, unlit so it reads in the dark.
type Marker struct {
	Colour string `json:"colour"`
}

func (marker *Marker) Start(context *engine.Context) {
	colour := draw2d.MustHex(marker.Colour)
	material := &scene.Material{Name: "marker", Kind: scene.UnlitKind, BaseColor: graphics.Vector(colour)}
	context.Node.SetMesh(diamond(material))
}
