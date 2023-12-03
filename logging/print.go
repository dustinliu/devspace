package logging

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// TODO: use logging framwork for better traceability
var DebugMode bool

func Fatal(v ...any) {
	color.New(color.FgCyan).Fprint(os.Stderr, "[FATAL] ")
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}

func Error(v ...any) {
	color.New(color.FgCyan).Fprint(os.Stderr, "[ERROR] ")
	fmt.Fprintln(os.Stderr, v...)
}

func Debug(v ...any) {
	if DebugMode {
		color.New(color.FgCyan).Fprint(os.Stdout, "[DEBUG] ")
		fmt.Fprintln(os.Stdout, v...)
	}
}
