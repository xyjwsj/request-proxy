package util

import (
	"runtime"
)

type SystemType string

const (
	Macos   SystemType = "macos"
	Windows            = "windows"
	Linux              = "linux"
)

func System() SystemType {
	goos := runtime.GOOS
	if goos == string(Linux) {
		return Linux
	}
	if goos == string(Windows) {
		return Windows
	}
	return Macos
}

func CreatePlatformPath(path ...string) string {
	split := "/"
	if System() == Windows {
		split = "\\"
	}
	strBuf := ""
	for _, str := range path {
		strBuf += str + split
	}
	return strBuf[:len(strBuf)-1]
}
