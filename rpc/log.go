package rpc

import (
	"io"
	"net"
	"fmt"
)


type ServerLogger struct {
	LocalAddr   net.Addr
	RemoteAddr  net.Addr
	Output      io.Writer
}
func (l ServerLogger) LogError(err error) {
	if l.Output != nil {
		fmt.Fprintf(l.Output, "[RPC] [Server %s] client %s: Error: %s\n",
			l.LocalAddr, l.RemoteAddr, err.Error())
	}
}

type ClientLogger struct {
	LocalAddr   net.Addr
	RemoteAddr  net.Addr
	Output      io.Writer
}
func (l ClientLogger) LogError(err error) {
	if l.Output != nil {
		fmt.Fprintf(l.Output, "[RPC] [Client %s] server %s: Error: %s\n",
			l.LocalAddr, l.RemoteAddr, err.Error())
	}
}

