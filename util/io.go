package util

import (
	"io"
	"fmt"
)


// This function is a well-behaved substitution of fmt.Fscanln.
// The reader is recommended to be a buffered reader because
// this function only reads one character at a time.
// If the given reader is not buffered, this function could perform
// one system call per one character.
// Note that ...[\n][EOF] and ...[EOF] are not distinguished.
func WellBehavedScanLine(f io.Reader) ([]rune, error) {
	var buf = make([] rune, 0)
	for {
		var char rune
		var _, err = fmt.Fscanf(f, "%c", &char)
		if err != nil {
			if err == io.EOF && len(buf) > 0 {
				return buf, nil
			} else {
				return nil, err
			}
		}
		if char != '\n' {
			buf = append(buf, char)
		} else {
			return buf, nil
		}
	}
}

