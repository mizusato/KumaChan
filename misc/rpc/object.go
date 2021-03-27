package rpc

import (
	"io"
	"bytes"
	"errors"
	"compress/gzip"
	"encoding/binary"
	"kumachan/misc/rpc/kmd"
)


type KmdApi interface {
	SerializeToStream(v kmd.Object, t *kmd.Type, stream io.Writer) error
	DeserializeFromStream(t *kmd.Type, stream io.Reader) (kmd.Object, error)
}

func receiveObject(t *kmd.Type, conn io.Reader, limit uint, api KmdApi) (kmd.Object, error) {
	var length uint64
	err := binary.Read(conn, binary.BigEndian, &length)
	if err != nil { return nil, err }
	if limit != 0 && length > uint64(limit) {
		return nil, errors.New("object size limit exceeded")
	}
	var buf = make([] byte, length)
	_, err = io.ReadFull(conn, buf)
	if err != nil { return nil, err }
	var decompressed, gz_err = gzip.NewReader(bytes.NewReader(buf))
	if gz_err != nil { panic(gz_err) }
	return api.DeserializeFromStream(t, decompressed)
}

func sendObject(obj kmd.Object, t *kmd.Type, conn io.Writer, api KmdApi) error {
	var buf bytes.Buffer
	var compressed = gzip.NewWriter(&buf)
	err := api.SerializeToStream(obj, t, compressed)
	if err != nil { return err }
	err = compressed.Close()
	if err != nil { return err }
	var bin = buf.Bytes()
	err = binary.Write(conn, binary.BigEndian, uint64(len(bin)))
	if err != nil { return err }
	_, err = conn.Write(bin)
	return err
}

