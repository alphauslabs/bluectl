package logger

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/fatih/color"
)

var (
	uf = func(s string) string {
		r, _ := strconv.ParseInt(strings.TrimPrefix(s, "\\U"), 16, 32)
		return string(rune(r))
	}

	soliddot    = "\\U25CF"
	greencircle = "\\U1F7E2"
	xmark       = "\\U274C"

	green = color.New(color.FgGreen).SprintFunc()
	red   = color.New(color.FgRed).SprintFunc()
	info  = log.New(os.Stdout, "", log.LstdFlags)
	fail  = log.New(os.Stdout, "", log.LstdFlags)

	pfx int32
)

const (
	PrefixDefault int32 = iota // solid dot with timestamp
	PrefixNone                 // empty prefix
	PrefixText                 // info/fail text with timestamp
	PrefixEmoji                // use emoji prefix with timestamp
)

// SetPrefix sets the prefix style to p. Default is colored dots with timestamps.
func SetPrefix(p ...int32) {
	info.SetFlags(log.LstdFlags)
	fail.SetFlags(log.LstdFlags)
	if len(p) == 0 {
		return
	}

	switch p[0] {
	case PrefixNone:
		SetNoTimestamp()
	default:
		info.SetFlags(log.LstdFlags)
		fail.SetFlags(log.LstdFlags)
	}

	atomic.StoreInt32(&pfx, p[0])
}

func SetNoTimestamp() {
	info.SetFlags(0)
	fail.SetFlags(0)
}

func SendToStderr(all ...bool) {
	fail.SetOutput(os.Stderr)
	if len(all) > 0 {
		if all[0] {
			info.SetOutput(os.Stderr)
		}
	}
}

// Value of f doesn't matter, just its presence.
func getPfx(f ...bool) string {
	switch atomic.LoadInt32(&pfx) {
	case PrefixNone:
		return ""
	case PrefixText:
		if len(f) > 0 {
			return "[fail] "
		} else {
			return "[info] "
		}
	case PrefixEmoji:
		if len(f) > 0 {
			return uf(xmark) + " "
		} else {
			return uf(greencircle) + " "
		}
	case PrefixDefault:
		fallthrough
	default:
		return uf(soliddot) + " "
	}
}

// Info prints `v` into standard output (via log) with a green prefix "info:".
func Info(v ...interface{}) {
	m := fmt.Sprintln(v...)
	info.Printf("%v%s", green(getPfx()), m)
}

// Infof is the formatted version of Info().
func Infof(format string, v ...interface{}) {
	m := fmt.Sprintf(format, v...)
	info.Printf("%v%s", green(getPfx()), m)
}

// Error prints `v` into standard output (via log) with a red prefix "error:".
func Error(v ...interface{}) {
	m := fmt.Sprintln(v...)
	fail.Printf("%v%s", red(getPfx(true)), m)
}

// Errorf is the formatted version of Error().
func Errorf(format string, v ...interface{}) {
	m := fmt.Sprintf(format, v...)
	fail.Printf("%v%s", red(getPfx(true)), m)
}
