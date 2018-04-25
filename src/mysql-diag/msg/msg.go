package msg

import (
	"fmt"
	"github.com/fatih/color"
	"os"
)

var (
	Error = color.New(color.FgRed).SprintFunc()
	Alert = color.New(color.FgHiYellow).SprintFunc()
	Happy = color.New(color.FgGreen).SprintFunc()
)

func PrintfError(format string, a ...interface{}) {
	m := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stderr, Error(m))
}

func PrintfErrorIntro(intro string, format string, a ...interface{}) {
	m := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%s%s\n", intro, Error(m))
}
