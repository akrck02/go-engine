package install

import "github.com/akrck02/go-engine/failure"

// Error codes of the install package.
var (
	DownloadFailed       = failure.Define("install.download-failed", "the file could not be downloaded: check the network, and try again")
	ChecksumWrong        = failure.Define("install.checksum-wrong", "the file downloaded is not the one published (its checksum differs): it was not used")
	UnpackFailed         = failure.Define("install.unpack-failed", "the archive downloaded could not be unpacked")
	PathOutside          = failure.Define("install.path-outside", "the archive names a file outside the folder it unpacks into: it was not unpacked")
	NoBuildForSystem     = failure.Define("install.no-build-for-system", "there is no build of this tool for this system to download")
	JavaNotFound         = failure.Define("install.java-not-found", "Android's tools need a JDK, and none was given")
	LicenceNotAccepted   = failure.Define("install.licence-not-accepted", "a licence the packages need was not the one read and accepted: nothing was accepted for it")
	AndroidManagerFailed = failure.Define("install.android-manager-failed", "Android's sdkmanager could not install the packages")
	DataFolderNotFound   = failure.Define("install.data-folder-not-found", "the folder for the user's data could not be found")
)
