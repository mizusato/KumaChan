package rpc

import "io"


type CallStatus int
const (
	CALL_RECV_ARG CallStatus = iota
	CALL_RUNNING
	CALL_ERROR
	CALL_COMPLETE
)
type Call struct {
	Method     string
	Error      error
	ArgReader  *io.PipeReader
	ArgWriter  *io.PipeWriter
}
type Calls  map[uint64] *Call

