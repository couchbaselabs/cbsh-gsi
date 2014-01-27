package api

import (
	"fmt"
	"strings"
)

const (
	Reset      = "\x1b[0m"
	Bright     = "\x1b[1m"
	Dim        = "\x1b[2m"
	Underscore = "\x1b[4m"
	Blink      = "\x1b[5m"
	Reverse    = "\x1b[7m"
	Hidden     = "\x1b[8m"

	FgBlack   = "\x1b[30m"
	FgRed     = "\x1b[31m"
	FgGreen   = "\x1b[32m"
	FgYellow  = "\x1b[33m"
	FgBlue    = "\x1b[34m"
	FgMagenta = "\x1b[35m"
	FgCyan    = "\x1b[36m"
	FgWhite   = "\x1b[37m"

	BgBlack   = "\x1b[40m"
	BgRed     = "\x1b[41m"
	BgGreen   = "\x1b[42m"
	BgYellow  = "\x1b[43m"
	BgBlue    = "\x1b[44m"
	BgMagenta = "\x1b[45m"
	BgCyan    = "\x1b[46m"
	BgWhite   = "\x1b[47m"
)

// color the string s with color 'color' unless s is already colored
func Colorize(s string, color string) string {
	if len(s) > 2 && s[:2] == "\x1b[" {
		return s
	} else {
		return color + s + Reset
	}
}

func ColorizeAll(color string, args ...interface{}) string {
	var parts []string
	for _, arg := range args {
		parts = append(parts, Colorize(fmt.Sprintf("%v", arg), color))
	}
	return strings.Join(parts, "")
}

func Black(args ...interface{}) string {
	return ColorizeAll(FgBlack, args...)
}

func Red(args ...interface{}) string {
	return ColorizeAll(FgRed, args...)
}

func Green(args ...interface{}) string {
	return ColorizeAll(FgGreen, args...)
}

func Yellow(args ...interface{}) string {
	return ColorizeAll(FgYellow, args...)
}

func Blue(args ...interface{}) string {
	return ColorizeAll(FgBlue, args...)
}

func Magenta(args ...interface{}) string {
	return ColorizeAll(FgMagenta, args...)
}

func Cyan(args ...interface{}) string {
	return ColorizeAll(FgCyan, args...)
}

func White(args ...interface{}) string {
	return ColorizeAll(FgWhite, args...)
}
