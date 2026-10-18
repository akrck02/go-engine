package engine

import "github.com/akrck02/go-engine/failure"

// Error codes of the engine package.
var (
	SettingsNotFound      = failure.Define("engine.settings-not-found", "the project has no engine.json")
	SettingsUnreadable    = failure.Define("engine.settings-unreadable", "engine.json is not the project's settings")
	NoStartScene          = failure.Define("engine.no-start-scene", "engine.json names no start scene")
	SceneNotFound         = failure.Define("engine.scene-not-found", "the scene file does not exist")
	SceneUnreadable       = failure.Define("engine.scene-unreadable", "the scene file is not a scene")
	BehaviourUnregistered = failure.Define("engine.behaviour-unregistered", "no behaviour is registered by that name (engine.Register in the game's main)")
	BehaviourFields       = failure.Define("engine.behaviour-fields", "the scene's fields do not fit the behaviour's")
	NoFiles               = failure.Define("engine.no-files", "the scene was made without the project's files")
	ModelFormat           = failure.Define("engine.model-format", "only .glb models are read for now")
	ModelNotFound         = failure.Define("engine.model-not-found", "the model's file does not exist")
	ModelUnreadable       = failure.Define("engine.model-unreadable", "the model's file is not a .glb go-renderer reads")
	PictureNotFound       = failure.Define("engine.picture-not-found", "the picture's file does not exist")
	PictureUnreadable     = failure.Define("engine.picture-unreadable", "the picture is not a PNG or a JPEG")
	SpriteNotFound        = failure.Define("engine.sprite-not-found", "the sprite's file does not exist")
	TilesetNotFound       = failure.Define("engine.tileset-not-found", "the tileset's file does not exist")
)
