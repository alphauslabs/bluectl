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

	soliddot    = "\\U25CF"
	greencircle = "\\U1F7E2"
	xmark       = "\\U274C"

	green = color.New(color.FgGreen).SprintFunc()
	red   = color.New(color.FgRed).SprintFunc()

	info = log.New(os.Stdout, green(uf(soliddot)+" "), log.LstdFlags)
	fail = log.New(os.Stdout, red(uf(soliddot)+" "), log.LstdFlags)
)

const (
	PrefixNone  = iota // empty prefix
	PrefixText         // info/fail text with timestamp
	PrefixEmoji        // use emoji prefix with timestamp
)

// SetPrefix sets the prefix style to p. Default is colored dots with timestamps.
func SetPrefix(p ...int) {
	info.SetFlags(log.LstdFlags)
	fail.SetFlags(log.LstdFlags)
	info.SetPrefix(green(uf(soliddot)) + " ")
	fail.SetPrefix(red(uf(soliddot)) + " ")
	if len(p) == 0 {
		return
	}

	switch p[0] {
	case PrefixNone:
		SetNoTimestamp()
		info.SetPrefix("")
		fail.SetPrefix("")
	case PrefixText:
		info.SetFlags(log.LstdFlags)
		info.SetPrefix(green("[info]") + " ")
		fail.SetFlags(log.LstdFlags)
		fail.SetPrefix(red("[fail]") + " ")
	case PrefixEmoji:
		info.SetFlags(log.LstdFlags)
		info.SetPrefix(green(uf(greencircle)) + " ")
		fail.SetFlags(log.LstdFlags)
		fail.SetPrefix(red(uf(xmark)) + " ")
	}
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
