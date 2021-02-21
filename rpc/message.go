package rpc

import (
	"io"
	"fmt"
	"bytes"
	"strconv"
	"strings"
)


const MsgKindWidth = 8
const MsgCallIdWidth = 16
const MsgPayloadLengthWidth = 8
const MsgPayloadLengthMax = 99999999

const MSG_SERVICE = "service"
const MSG_CREATED = "created"
const MSG_CALL = "call"
const MSG_CALL_MULTI = "call*"
const MSG_VALUE = "value"
const MSG_ERROR = "error"
const MSG_COMPLETE = "complete"

func writeMessageHeaderField(content string, width int, w io.Writer) error {
	if len(content) > width {
		panic(fmt.Sprintf("field content width exceeded maximum (%d)", width))
	}
	var buf = make([] byte, width)
	for i := 0; i < width; i += 1 {
		if i < len(content) {
			buf[i] = content[i]
		} else {
			buf[i] = ' '
		}
	}
	_, err := w.Write(buf)
	return err
}

func readMessageHeaderField(width int, r io.Reader) (string, error) {
	var buf = make([] byte, width)
	_, err := io.ReadFull(r, buf)
	if err != nil { return "", err }
	var raw_str = string(buf)
	var str = strings.TrimRight(raw_str, " ")
	return str, nil
}

func sendMessage(kind string, id uint64, payload ([] byte), conn io.Writer) error {
	if len(payload) > MsgPayloadLengthMax {
		return fmt.Errorf("message payload length exceeded maximum (%d)", MsgPayloadLengthMax)
	}
	var buf bytes.Buffer
	err := writeMessageHeaderField(kind, MsgKindWidth, &buf)
	if err != nil { return err }
	id_string := strconv.FormatUint(id, 16)
	err = writeMessageHeaderField(id_string, MsgCallIdWidth, &buf)
	if err != nil { return err }
	length := strconv.Itoa(len(payload))
	err = writeMessageHeaderField(length, MsgPayloadLengthWidth, &buf)
	if err != nil { return err }
	_, err = buf.Write(payload)
	if err != nil { return err }
	_, err = conn.Write(buf.Bytes())
	if err != nil { return err }
	return nil
}

func receiveMessage(conn io.Reader) (string, uint64, ([] byte), error) {
	kind, err := readMessageHeaderField(MsgKindWidth, conn)
	if err != nil { return "", ^uint64(0), nil, err }
	id_string, err := readMessageHeaderField(MsgCallIdWidth, conn)
	if err != nil { return kind, ^uint64(0), nil, err }
	id, err := strconv.ParseUint(id_string, 16, 64)
	if err != nil { return kind, ^uint64(0), nil, fmt.Errorf("invalid call id: %w", err) }
	length_string, err := readMessageHeaderField(MsgPayloadLengthWidth, conn)
	if err != nil { return kind, id, nil, err }
	length, err := strconv.Atoi(length_string)
	if err != nil { return kind, id, nil, fmt.Errorf("invalid payload length: %w", err) }
	buf := make([] byte, length)
	_, err = io.ReadFull(conn, buf)
	if err != nil { return kind, id, nil, err }
	return kind, id, buf, nil
}

func sendError(e error, id uint64, conn io.Writer) error {
	var desc = e.Error()
	if len(desc) > MsgPayloadLengthMax {
		desc = desc[:MsgPayloadLengthMax]
	}
	return sendMessage(MSG_ERROR, id, ([] byte)(desc), conn)
}

