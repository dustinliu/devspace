package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/fatih/color"
)

var DebugMode bool

func Fatal(v ...any) {
	color.New(color.FgRed).Fprint(os.Stderr, "[FATAL] ")
	if DebugMode {
		color.New(color.FgHiCyan).Fprint(os.Stderr, "("+trace()+") ")
	}
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}

func Error(v ...any) {
	color.New(color.FgHiRed).Fprint(os.Stderr, "[ERROR] ")
	if DebugMode {
		color.New(color.FgCyan).Fprint(os.Stderr, "("+trace()+") ")
	}
	fmt.Fprintln(os.Stderr, v...)
}

func Debug(v ...any) {
	if DebugMode {
		color.New(color.FgYellow).Fprint(os.Stdout, "[DEBUG] ")
		color.New(color.FgCyan).Fprint(os.Stdout, "("+trace()+") ")
		fmt.Fprintln(os.Stdout, v...)
	}
}

func trace() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[1:n])
	frame, _ := frames.Next()
	return filepath.Base(frame.Function) + ":" + strconv.Itoa(frame.Line)
}
