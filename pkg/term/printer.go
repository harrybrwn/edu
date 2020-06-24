package term

import (
	"bytes"
	"fmt"
)

func Colorf(format string, a ...string) string {
	end := len(format)
	arg := 0
	buf := &bytes.Buffer{}

	for i := 0; i < end; i++ {
		prev := i
		for i < end && format[i] != '%' {
			i++
		}
		if i > prev {
			buf.WriteString(format[prev:i])
		}
		if i >= end {
			break
		}

		i++
		color := getForground(format[i])
		// if we don't know what it is
		// then put it back assuming it
		// is meant for the fmt package
		if color == 0 {
			buf.WriteByte('%')
			buf.WriteByte(format[i])
			continue
		}

		fmt.Fprintf(buf, "%s[%dm", escape, color)
		buf.WriteString(a[arg])
		fmt.Fprintf(buf, "%s[0m", escape)
		arg++
	}
	return buf.String()
}

func getForground(code byte) int {
	switch code {
	case 'r':
		return FgRed
	case 'g':
		return FgGreen
	case 'y':
		return FgYellow
	case 'b':
		return FgBlue
	case 'm':
		return FgMagenta
	case 'c':
		return FgCyan
	case 'w':
		return FgWhite
	}
	return 0
}
