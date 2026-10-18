package install

import "github.com/akrck02/go-engine/failure"

// Error codes of the install package.
var (
	DownloadFailed     = failure.Define("install.download-failed", "the file could not be downloaded: check the network, and try again")
	ChecksumWrong      = failure.Define("install.checksum-wrong", "the file downloaded is not the one published (its checksum differs): it was not used")
	UnpackFailed       = failure.Define("install.unpack-failed", "the archive downloaded could not be unpacked")
	PathOutside        = failure.Define("install.path-outside", "the archive names a file outside the folder it unpacks into: it was not unpacked")
	NoBuildForSystem   = failure.Define("install.no-build-for-system", "there is no build of this tool for this system to download")
	DataFolderNotFound = failure.Define("install.data-folder-not-found", "the folder for the user's data could not be found")
)
