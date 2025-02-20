package internal

import (
	"fmt"
	"io"
	"ztna-core/ztna/logtrace"
)

/*
Extends the standard FPrintF with overwriting the current line because it has the `\u001B[2K`
*/
func FPrintfReusingLine(writer io.Writer, format string, a ...any) (n int, err error) {
	logtrace.LogWithFunctionName()
	return fmt.Fprintf(writer, "\u001B[2K"+format+"\r", a...)
}
func FPrintflnReusingLine(writer io.Writer, format string, a ...any) (n int, err error) {
	logtrace.LogWithFunctionName()
	return FPrintfReusingLine(writer, format+"\n", a...)
}
