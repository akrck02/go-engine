package editor

import "github.com/akrck02/go-engine/project"

// assetIcon is the icon of a kind of asset.
func assetIcon(kind project.AssetKind) string {
	switch kind {
	case project.ModelAsset:
		return iconModel
	case project.ImageAsset:
		return iconPicture
	case project.SoundAsset:
		return iconSound
	}
	return iconFile
}
