package common

import (
	"fmt"
	"strconv"
)

var MAJOR_VERSION string
var MINOR_VERSION string
var PATCH_VERSION string

func GetMajorVersion() int {
	out, _ := strconv.Atoi(MAJOR_VERSION)
	return out
}

func GetMinorVersion() int {
	out, _ := strconv.Atoi(MINOR_VERSION)
	return out
}

func GetPatchVersion() int {
	out, _ := strconv.Atoi(PATCH_VERSION)
	return out
}

func GetVersionString() string {
	return fmt.Sprintf("0.%s.%s-%s", MAJOR_VERSION, MINOR_VERSION, PATCH_VERSION)
}
