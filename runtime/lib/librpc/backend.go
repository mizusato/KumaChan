package librpc

import "net"


type ServerBackend interface {
	Serve() (net.Listener, error)
}
type ClientBackend interface {
	Access() (net.Conn, error)
}

type ServerCleartextNet struct {
	Network  string
	Address  string
}
type ClientCleartextNet struct {
	Network  string
	Address  string
}
func (backend ServerCleartextNet) Serve() (net.Listener, error) {
	return net.Listen(backend.Network, backend.Address)
}
func (backend ClientCleartextNet) Access() (net.Conn, error) {
	return net.Dial(backend.Network, backend.Address)
}

