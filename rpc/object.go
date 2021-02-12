package rpc

import (
	"io"
	"compress/gzip"
	"kumachan/rpc/kmd"
)


func receiveObject(t *kmd.Type, conn io.Reader, limit uint, api KmdApi) (kmd.Object, error) {
	var decompressed, gz_err = gzip.NewReader(conn)
	if gz_err != nil { panic(gz_err) }
	var limited = &LimitedReader {
		Underlying: decompressed,
		SizeLimit:  limit,
	}
	return api.DeserializeFromStream(t, limited)
}

func sendObject(obj kmd.Object, t *kmd.Type, conn io.Writer, api KmdApi) error {
	var compressed = gzip.NewWriter(conn)
	return api.SerializeToStream(obj, t, compressed)
}

