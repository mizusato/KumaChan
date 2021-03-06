package librpc

import (
	"os"
	"io"
	"kumachan/standalone/rpc"
)


type ServerOptions struct {
	CommonOptions
}
type ClientOptions struct {
	CommonOptions
}
type CommonOptions struct {
	LogEnabled  bool
	rpc.Limits
}
func (opts CommonOptions) GetDebugOutput() io.Writer {
	if opts.LogEnabled {
		return os.Stderr
	} else {
		return nil
	}
}

