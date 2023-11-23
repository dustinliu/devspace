package core

import (
	"fmt"
	"os"
)

var DebugMode bool

func Fatal(v ...any) {
	fmt.Fprint(os.Stderr, "[FATAL] ")
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}

func Error(v ...any) {
	fmt.Fprint(os.Stderr, "[ERROR] ")
	fmt.Fprintln(os.Stderr, v...)
}

func Print(v ...any) {
	fmt.Fprint(os.Stdout, v...)
}

func Println(v ...any) {
	fmt.Fprintln(os.Stdout, v...)
}

func Debug(v ...any) {
	if DebugMode {
		fmt.Fprint(os.Stdout, "[DEBUG] ")
		fmt.Fprintln(os.Stdout, v...)
	}
}
