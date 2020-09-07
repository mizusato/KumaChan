package util

import (
	"io"
	"fmt"
)


func WellBehavedScanLine(f io.Reader) ([]rune, error) {
	// This function is a well-behaved substitution of fmt.Fscanln
	//   (fmt.Fscanln does not accept empty lines)
	// Note: ...[\n][EOF] and ...[EOF] are not distinguished
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

