package common

import (
	"strconv"
	"strings"
)

var VERSION string
var BUILD string

func GetVersionFloat() float64 {
	out, err := strconv.ParseFloat(strings.SplitN(VERSION, ".", 2)[1], 64)
	if err != nil {
		panic(err)
	}
	return out
}
