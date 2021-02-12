package rpc

import (
	"io"
	"time"
	"errors"
)


type Limits struct {
	SendTimeout        time.Duration
	RecvTimeout        time.Duration
	RecvInterval       time.Duration
	RecvMaxObjectSize  uint
}

type LimitedReader struct {
	Underlying   io.Reader
	CurrentRead  uint
	SizeLimit    uint
}
func (l *LimitedReader) Read(buf ([] byte)) (int, error) {
	var n, err = l.Underlying.Read(buf)
	l.CurrentRead += uint(n)
	if l.SizeLimit != 0 && l.CurrentRead > l.SizeLimit {
		return n, errors.New("object size limit exceeded")
	}
	return n, err
}

