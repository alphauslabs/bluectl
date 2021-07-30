package logger

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

var (
	uf = func(s string) string {
		r, _ := strconv.ParseInt(strings.TrimPrefix(s, "\\U"), 16, 32)
		return string(r)
	}

	green = color.New(color.FgGreen).SprintFunc()
	red   = color.New(color.FgRed).SprintFunc()
	info  = log.New(os.Stdout, green(uf("\\U1F7E2")+" "), log.LstdFlags) // green circle
	fail  = log.New(os.Stdout, red(uf("\\U274C")+" "), log.LstdFlags)    // red x mark
)

func SetNoEmoji() {
	info.SetFlags(log.LstdFlags)
	info.SetPrefix(green("[info] "))
	fail.SetFlags(log.LstdFlags)
	fail.SetPrefix(red("[fail] "))
}

func SetCleanOutput() {
	info.SetFlags(0)
	info.SetPrefix("")
	fail.SetFlags(0)
	fail.SetPrefix("")
}

// Info prints `v` into standard output (via log) with a green prefix "info:".
func Info(v ...interface{}) {
	m := fmt.Sprintln(v...)
	info.Print(m)
}

// Infof is the formatted version of Info().
func Infof(format string, v ...interface{}) {
	m := fmt.Sprintf(format, v...)
	info.Print(m)
}

// Error prints `v` into standard output (via log) with a red prefix "error:".
func Error(v ...interface{}) {
	m := fmt.Sprintln(v...)
	fail.Print(m)
}

// Errorf is the formatted version of Error().
func Errorf(format string, v ...interface{}) {
	m := fmt.Sprintf(format, v...)
	fail.Print(m)
}
