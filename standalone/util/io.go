package util

import (
	"io"
	"fmt"
	"strings"
)


// This function is a well-behaved substitution of fmt.Fscanln.
// The reader is recommended to be a buffered reader because
// this function only reads one character at a time.
// If the given reader is not buffered, this function could perform
// one system call per one character.
// Note that ...[\n][EOF] and ...[EOF] are not distinguished.
func WellBehavedReadLine(f io.Reader) ([]rune, error) {
	var buf = make([] rune, 0)
	var char rune
	for {
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

// standard-library-like version of WellBehavedReadLine
func WellBehavedFscanln(f io.Reader, output *string) (int, error) {
	var buf strings.Builder
	var total = 0
	var one_byte = [] byte { 0 }
	for {
		var _, err = f.Read(one_byte)
		total += 1
		if err != nil {
			if err == io.EOF && buf.Len() > 0 {
				*output = buf.String()
				return total, nil
			} else {
				return total, err
			}
		}
		if rune(one_byte[0]) != '\n' {
			buf.WriteByte(one_byte[0])
		} else {
			*output = buf.String()
			return total, nil
		}
	}
}

