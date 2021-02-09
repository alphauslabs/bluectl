package logger

import (
	"fmt"
	"log"

	"github.com/fatih/color"
)

var (
	green = color.New(color.FgGreen).SprintFunc()
	red   = color.New(color.FgRed).SprintFunc()
)

// Info prints `v` into standard output (via log) with a green prefix "info:".
func Info(v ...interface{}) {
	m := fmt.Sprintln(v...)
	log.Printf("%s %s", green("[info]"), m)
}

// Infof is the formatted version of Info().
func Infof(format string, v ...interface{}) {
	m := fmt.Sprintf(format, v...)
	log.Printf("%s %s", green("[info]"), m)
}

// Error prints `v` into standard output (via log) with a red prefix "error:".
func Error(v ...interface{}) {
	m := fmt.Sprintln(v...)
	log.Printf("%s %s", red("[error]"), m)
}

// Errorf is the formatted version of Error().
func Errorf(format string, v ...interface{}) {
	m := fmt.Sprintf(format, v...)
	log.Printf("%s %s", red("[error]"), m)
}
