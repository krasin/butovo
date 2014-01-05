package api

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"time"
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
	Cmd     RequestType
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
		Cmd:     typ,
		Channel: ch,
		Data:    data[8:],
	}, nil
}

// ReadResponse read a response from the writer.
// The format is:
// <uint32 size> <uint32 ch> <int64 ts> <data>
// where all multi-byte fields are low-endian.
// Size is the number of bytes in the rest of the message.
// Channel must not exceed MaxChannel.
// Timestamp is serialized as a number of nanoseconds since January 1, 1970 UTC.
// Data length must be not greater than MaxSize.
func ReadResponse(r io.Reader) (ch uint32, ts time.Time, data []byte, err error) {
	var size uint32
	if err = binary.Read(r, binary.LittleEndian, &size); err != nil {
		err = fmt.Errorf("could not read response body size: %v", err)
		return
	}
	if size > MaxSize+12 {
		err = fmt.Errorf("response body size too large: %d. Max packet size: %d", size, MaxSize+12)
		return
	}
	if size < 12 {
		err = fmt.Errorf("response body size too small: %d. Min response size: 12", size)
	}
	data = make([]byte, size)
	if _, err = io.ReadFull(r, data); err != nil {
		data = nil
		err = fmt.Errorf("could not read response body (size: %d): %v", size, err)
		return
	}
	ch = uint32(data[0]) + uint32(data[1])<<8 + uint32(data[2])<<16 + uint32(data[3])<<24
	if ch > MaxChannel {
		data = nil
		err = fmt.Errorf("channel is too large: %d. Max channel: %d", ch, MaxChannel)
		return
	}
	v := int64(data[4]) + int64(data[5])<<8 + int64(data[6])<<16 + int64(data[7])<<24 +
		int64(data[8])<<32 + int64(data[9])<<40 + int64(data[10])<<48 + int64(data[11])<<56
	ts = time.Unix(v/int64(1E9), v%int64(1E9))
	data = data[12:]
	return
}
