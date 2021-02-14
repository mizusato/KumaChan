package rpc

import (
	"time"
)


type Limits struct {
	SendTimeout        time.Duration
	RecvTimeout        time.Duration
	RecvInterval       time.Duration
	RecvMaxObjectSize  uint
}

