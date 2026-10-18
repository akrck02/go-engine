package editor

import ui "github.com/akrck02/go-ui"

// A file is chosen through go-ui's file picker, shown over the whole editor while a workspace
// asks for one: the path of a tool, a model to import.

// filePicking is a file asked for: how the picker is shown, and what to do with the path chosen.
type filePicking struct {
	options ui.FilePickerOptions
	chosen  func(path string)
}

// pickFile asks the person for a file or a folder; chosen gets its path when they choose one.
func (view *View) pickFile(options ui.FilePickerOptions, chosen func(path string)) {
	view.picking = &filePicking{options: options, chosen: chosen}
}

// filePicker shows the file picker while a file is asked for.
func (view *View) filePicker(context *ui.Context) {
	if view.picking == nil {
		return
	}
	path, closed := context.FilePicker(view.picking.options)
	if !closed {
		return
	}
	asked := view.picking
	view.picking = nil
	if path != "" {
		asked.chosen(path)
	}
}
