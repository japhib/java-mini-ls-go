package util

import (
	"github.com/pkg/errors"
	"net/url"
	"runtime"
	"unicode"
)

const isWindows = runtime.GOOS == "windows"

func FileURIToPath(uri string) (string, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return "", errors.Wrap(err, "error parsing filepath URI")
	}

	path := u.Path

	if isWindows && isWindowsAbsPath(path) {
		// Strip leading "/" on windows absolute paths
		path = path[1:]
	}

	return path, nil
}

// returns whether the path starts with something like "/c:/" (leading slash needs to be trimmed)
func isWindowsAbsPath(path string) bool {
	return path[0] == '/' && unicode.IsLetter(rune(path[1])) && path[2] == ':' && path[3] == '/'
}
