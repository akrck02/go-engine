package scenefile

import "github.com/akrck02/go-engine/failure"

// Error codes of the scenefile package.
var (
	Unreadable      = failure.Define("scenefile.unreadable", "the file is not a scene's JSON")
	NodeWithoutName = failure.Define("scenefile.node-without-name", "a node has no name")
	SlashInName     = failure.Define("scenefile.slash-in-name", "a node's name has a slash, which separates the names of a path")
	SameName        = failure.Define("scenefile.same-name", "two nodes under the same parent have the same name")
	UnknownShape    = failure.Define("scenefile.unknown-shape", "a node is a shape that does not exist")

	UnknownBodyKind   = failure.Define("scenefile.unknown-body-kind", "a body's kind is not one of BodyKinds")
	TileSizeWrong     = failure.Define("scenefile.tile-size-wrong", "a tile map's tiles have no size")
	TileCountWrong    = failure.Define("scenefile.tile-count-wrong", "a tile map's tiles don't fill its columns and rows")
	SheetSizeWrong    = failure.Define("scenefile.sheet-size-wrong", "a sprite sheet has no columns or no rows")
	FrameOutsideSheet = failure.Define("scenefile.frame-outside-sheet", "an animation shows a frame the sheet does not have")
	SameAnimationName = failure.Define("scenefile.same-animation-name", "two animations of a sheet have the same name")
	UnknownAnimation  = failure.Define("scenefile.unknown-animation", "a sheet starts with an animation it does not have")
	NoPicture         = failure.Define("scenefile.no-picture", "a sprite or a tileset names no picture")
	UnknownLightKind  = failure.Define("scenefile.unknown-light-kind", "a light's kind is not point or sprite")
	LightRadiiWrong   = failure.Define("scenefile.light-radii-wrong", "a light's radii are below zero, or the inner one is beyond the outer")
)
