package editor

import (
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/akrck02/go-engine/failure"
	"github.com/akrck02/go-engine/scenefile"
)

// The Art workspace's work, without its interface: a picture becomes a sprite (cut into frames,
// with animations) or a tileset (cut into tiles, some solid), each a file of its own beside the
// picture that every scene shares. One of them is open at a time; its changes are saved to its
// file, apart from the scene's.

// artFile is the sprite or tileset open in Art.
type artFile struct {
	path    string
	sprite  *scenefile.Sprite
	tileset *scenefile.Tileset
	saved   bool
}

// ArtOpen is the file open in Art, and whether it is a sprite or a tileset.
func (editor *Editor) ArtOpen() (file string, isSprite, isTileset bool) {
	return editor.art.path, editor.art.sprite != nil, editor.art.tileset != nil
}

// ArtSaved tells whether the file open in Art has no changes left to save.
func (editor *Editor) ArtSaved() bool { return editor.art.path == "" || editor.art.saved }

// OpenArt opens a sprite's or a tileset's file of the project in Art; the file open before must
// have no changes left to save.
func (editor *Editor) OpenArt(file string) error {
	if !editor.ArtSaved() {
		return failure.Wrap(ArtUnsaved, editor.art.path, nil)
	}
	data, err := fs.ReadFile(editor.Project.Files(), file)
	if err != nil {
		return failure.Wrap(ArtNotFound, file, err)
	}
	opened := artFile{path: file, saved: true}
	switch {
	case scenefile.IsSpriteFile(file):
		opened.sprite, err = scenefile.ReadSprite(data)
	case scenefile.IsTilesetFile(file):
		opened.tileset, err = scenefile.ReadTileset(data)
	default:
		return failure.Wrap(ArtWrongKind, file, nil)
	}
	if err != nil {
		return err
	}
	editor.art = opened
	return nil
}

// CloseArt closes the file open in Art, which must have no changes left to save.
func (editor *Editor) CloseArt() error {
	if !editor.ArtSaved() {
		return failure.Wrap(ArtUnsaved, editor.art.path, nil)
	}
	editor.art = artFile{}
	return nil
}

// MakeSprite makes a picture of the project a sprite of columns and rows of frames, in a file
// beside it ("lamp.png" makes "lamp.sprite.json"), and opens it; a sprite already made is opened.
func (editor *Editor) MakeSprite(picture string, columns, rows int) (string, error) {
	file := beside(picture, scenefile.SpriteSuffix)
	sprite := &scenefile.Sprite{Picture: picture, SpriteSheet: scenefile.SpriteSheet{Columns: max(columns, 1), Rows: max(rows, 1)}}
	return file, editor.makeArt(file, artFile{path: file, sprite: sprite})
}

// MakeTileset makes a picture of the project a tileset of square tiles of a size, in pixels, in a
// file beside it, and opens it; a tileset already made is opened.
func (editor *Editor) MakeTileset(picture string, tileSize int) (string, error) {
	file := beside(picture, scenefile.TilesetSuffix)
	tileset := &scenefile.Tileset{Picture: picture, TileSize: max(tileSize, 1)}
	return file, editor.makeArt(file, artFile{path: file, tileset: tileset})
}

// beside is a file named as a picture, with another suffix: "a/lamp.png" and ".sprite.json" make
// "a/lamp.sprite.json".
func beside(picture, suffix string) string {
	return strings.TrimSuffix(picture, path.Ext(picture)) + suffix
}

// makeArt writes a new sprite or tileset and opens it, or opens the one already there.
func (editor *Editor) makeArt(file string, made artFile) error {
	if !editor.ArtSaved() {
		return failure.Wrap(ArtUnsaved, editor.art.path, nil)
	}
	if _, err := os.Stat(editor.Project.Path(file)); err == nil {
		return editor.OpenArt(file)
	}
	editor.art = made
	return editor.SaveArt()
}

// SaveArt writes the file open in Art; the scene shows it changed at once.
func (editor *Editor) SaveArt() error {
	var data []byte
	var err error
	switch {
	case editor.art.sprite != nil:
		data, err = editor.art.sprite.Marshal()
	case editor.art.tileset != nil:
		data, err = editor.art.tileset.Marshal()
	default:
		return nil
	}
	if err == nil {
		err = os.WriteFile(editor.Project.Path(editor.art.path), append(data, '\n'), 0o644)
	}
	if err != nil {
		return failure.Wrap(ArtNotSaved, editor.art.path, err)
	}
	editor.art.saved = true
	editor.Scene.Forget(editor.art.path)
	editor.Scene.Sync()
	return nil
}

// saveArtIfChanged saves the file open in Art when it has changes.
func (editor *Editor) saveArtIfChanged() error {
	if editor.ArtSaved() {
		return nil
	}
	return editor.SaveArt()
}

// editedSheet is the sheet the sheet's tools change: the sprite open in Art, or else the chosen
// node's own.
func (editor *Editor) editedSheet() (*scenefile.SpriteSheet, bool) {
	if editor.art.sprite != nil {
		return &editor.art.sprite.SpriteSheet, true
	}
	if editor.Selected == nil {
		return nil, false
	}
	return editor.Selected.Sheet()
}

// sheetChanged marks what holds the edited sheet as changed: the file open in Art, or the scene.
func (editor *Editor) sheetChanged() {
	if editor.art.sprite != nil {
		editor.art.saved = false
		return
	}
	editor.Changed()
}

// AddFrame puts a frame of the sheet at the end of an animation: clicking frames on the picture
// makes an animation of them, in the order clicked. A frame the sheet does not have is left out.
func (editor *Editor) AddFrame(index, frame int) {
	animation, sheet, found := editor.selectedAnimation(index)
	if !found || frame < 0 || frame >= sheet.FrameCount() {
		return
	}
	animation.Frames = append(animation.Frames, frame)
	editor.sheetChanged()
}

// ClearFrames empties an animation, to click its frames again.
func (editor *Editor) ClearFrames(index int) {
	animation, _, found := editor.selectedAnimation(index)
	if !found {
		return
	}
	animation.Frames = nil
	editor.sheetChanged()
}

// RemoveFrameAt takes the frame at a place out of an animation.
func (editor *Editor) RemoveFrameAt(index, place int) {
	animation, _, found := editor.selectedAnimation(index)
	if !found || place < 0 || place >= len(animation.Frames) {
		return
	}
	animation.Frames = slices.Delete(animation.Frames, place, place+1)
	editor.sheetChanged()
}

// StartWith makes an animation the one that plays when the scene starts.
func (editor *Editor) StartWith(index int) {
	animation, sheet, found := editor.selectedAnimation(index)
	if !found {
		return
	}
	sheet.Start = animation.Name
	editor.sheetChanged()
}

// ToggleSolidTile makes a tile of the tileset open in Art solid, or not.
func (editor *Editor) ToggleSolidTile(tile int) {
	tileset := editor.art.tileset
	if tileset == nil || tile < 0 {
		return
	}
	if at := slices.Index(tileset.Solid, tile); at >= 0 {
		tileset.Solid = slices.Delete(tileset.Solid, at, at+1)
	} else {
		tileset.Solid = append(tileset.Solid, tile)
		slices.Sort(tileset.Solid)
	}
	editor.art.saved = false
}

// SetArtTileSize sets how many pixels a side the tiles of the tileset open in Art are.
func (editor *Editor) SetArtTileSize(pixels int) {
	if editor.art.tileset == nil {
		return
	}
	editor.art.tileset.TileSize = max(pixels, 1)
	editor.art.saved = false
}

// MakeSpriteOfNode moves the chosen node's own sheet into a sprite's file beside its picture, makes
// the node show that file, and opens it in Art: a scene written before sprites had files of their
// own is brought up to date one node at a time. When the file is there already (another node made
// it), the node shows it as it is: a shared sprite is never written over.
func (editor *Editor) MakeSpriteOfNode() (string, error) {
	node := editor.Selected
	if node == nil {
		return "", nil
	}
	sheet, own := node.Sheet()
	if !own || scenefile.IsSpriteFile(node.Sprite()) {
		return "", failure.Wrap(NoOwnSheet, node.Path(), nil)
	}
	if !editor.ArtSaved() {
		return "", failure.Wrap(ArtUnsaved, editor.art.path, nil)
	}
	file := beside(node.Sprite(), scenefile.SpriteSuffix)
	if _, err := os.Stat(editor.Project.Path(file)); err != nil {
		editor.art = artFile{path: file, sprite: &scenefile.Sprite{Picture: node.Sprite(), SpriteSheet: *sheet.Copy()}}
		if err := editor.SaveArt(); err != nil {
			return "", err
		}
	}
	node.SetSheet(nil)
	if err := editor.SetSprite(file); err != nil {
		return "", err
	}
	return file, editor.OpenArt(file)
}
