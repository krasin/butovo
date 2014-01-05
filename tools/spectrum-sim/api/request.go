package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

type RequestType int

const (
	Send   RequestType = 0
	Listen RequestType = 1
)

const (
	// Max data length
	MaxSize = 128

	MaxChannel = math.MaxInt32
)

type Request struct {
	Type    RequestType
	Channel uint32
	Data    []byte
}

// ReadRequest reads a request from the reader.
// The format is:
// <uint32 size> <uint32 cmd> <uint32 ch> <data>
// where all multi-byte fields are low-endian.
// Size is the number of bytes in the rest of the message.
// Channel must not exceeed MaxChannel.
// Data length must be not greater than MaxSize.
func ReadRequest(r io.Reader) (cmd *Request, err error) {
	var size uint32
	if err = binary.Read(r, binary.LittleEndian, &size); err != nil {
		err = fmt.Errorf("could not read command body size: %v", err)
		return
	}
	if size > MaxSize+8 {
		err = fmt.Errorf("command body size too large: %d. Max packet size: %d", size, MaxSize+8)
		return
	}
	if size < 8 {
		err = fmt.Errorf("command body size too small: %d. Min packet size: 8", size)
		return
	}
	data := make([]byte, size)
	if _, err = io.ReadFull(r, data); err != nil {
		data = nil
		err = fmt.Errorf("could not read command body (size: %d): %v", size, err)
		return
	}
	typ := RequestType(uint32(data[0]) + uint32(data[1])<<8 + uint32(data[2])<<16 + uint32(data[3])<<24)
	ch := uint32(data[4]) + uint32(data[5])<<8 + uint32(data[6])<<16 + uint32(data[7])<<24
	if ch > MaxChannel {
		data = nil
		err = fmt.Errorf("channel is too large: %d. Max channel: %d", ch, MaxChannel)
		return
	}
	return &Request{
		Type:    typ,
		Channel: ch,
		Data:    data[8:],
	}, nil
}

// WriteRequest converts a command into the request data.
// The format is described in the documentation to ReadRequest.
func WriteRequest(req *Request) ([]byte, error) {
	var buf bytes.Buffer

	if len(req.Data) > MaxSize {
		return nil, fmt.Errorf("too large data size: %d. Max data size: %d", len(req.Data), MaxSize)
	}
	size := 8 + len(req.Data)
	if err := binary.Write(&buf, binary.LittleEndian, uint32(size)); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, uint32(req.Type)); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, uint32(req.Channel)); err != nil {
		return nil, err
	}
	if _, err := buf.Write(req.Data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
