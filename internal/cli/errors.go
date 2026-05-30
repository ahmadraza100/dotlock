package cli

import "errors"

// errSilent is returned by commands after ui.Error() has been called,
// signaling main not to print the error a second time.
var errSilent = errors.New("")
