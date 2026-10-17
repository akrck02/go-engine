package scenefile

import "github.com/akrck02/go-engine/failure"

// Error codes of the scenefile package.
var (
	Unreadable      = failure.Define("scenefile.unreadable", "the file is not a scene's JSON")
	NodeWithoutName = failure.Define("scenefile.node-without-name", "a node has no name")
	SlashInName     = failure.Define("scenefile.slash-in-name", "a node's name has a slash, which separates the names of a path")
	SameName        = failure.Define("scenefile.same-name", "two nodes under the same parent have the same name")
	UnknownShape    = failure.Define("scenefile.unknown-shape", "a node is a shape that does not exist")
)
