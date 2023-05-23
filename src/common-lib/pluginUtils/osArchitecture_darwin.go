package pluginUtils

// Is64BitOS verifies os architecture
func Is64BitOS() bool {
	// TODO: implement logic of this method when linux support will be added
	return true
}

// DisableRedirection disables file system redirection for the calling thread. File system redirection is enabled by default.
// OldValue is file system redirection value.
// The system uses this parameter to store information necessary to re-enable file system redirection.
func DisableRedirection(OldValue *uintptr) {
}

// RevertRedirection restores file system redirection for the calling thread.
// This function should not be called without a previous call to the DisableRedirection function.
func RevertRedirection(OldValue *uintptr) {
}
