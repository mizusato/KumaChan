package librpc

import (
	"os"
	"io"
	"kumachan/rpc"
)


type ServerOptions struct {
	CommonOptions
}
type ClientOptions struct {
	CommonOptions
}
type CommonOptions struct {
	Debug bool
	rpc.Limits
}
func (opts CommonOptions) GetDebugOutput() io.Writer {
	if opts.Debug {
		return os.Stderr
	} else {
		return nil
	}
}

