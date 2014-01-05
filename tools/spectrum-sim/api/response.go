package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"time"
)

type Response struct {
	Channel   uint32
	Timestamp time.Time
	Data      []byte
}

// ReadResponse read a response from the writer.
// The format is:
// <uint32 size> <uint32 ch> <int64 ts> <data>
// where all multi-byte fields are low-endian.
// Size is the number of bytes in the rest of the message.
// Channel must not exceed MaxChannel.
// Timestamp is serialized as a number of nanoseconds since January 1, 1970 UTC.
// Data length must be not greater than MaxSize.
func ReadResponse(r io.Reader) (*Response, error) {
	var size uint32
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, fmt.Errorf("could not read response body size: %v", err)
	}
	if size > MaxSize+12 {
		return nil, fmt.Errorf("response body size too large: %d. Max packet size: %d", size, MaxSize+12)
	}
	if size < 12 {
		return nil, fmt.Errorf("response body size too small: %d. Min response size: 12", size)
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, fmt.Errorf("could not read response body (size: %d): %v", size, err)
	}
	ch := uint32(data[0]) + uint32(data[1])<<8 + uint32(data[2])<<16 + uint32(data[3])<<24
	if ch > MaxChannel {
		return nil, fmt.Errorf("channel is too large: %d. Max channel: %d", ch, MaxChannel)
	}
	v := int64(data[4]) + int64(data[5])<<8 + int64(data[6])<<16 + int64(data[7])<<24 +
		int64(data[8])<<32 + int64(data[9])<<40 + int64(data[10])<<48 + int64(data[11])<<56
	ts := time.Unix(v/int64(1E9), v%int64(1E9))
	data = data[12:]
	return &Response{
		Channel:   ch,
		Timestamp: ts,
		Data:      data,
	}, nil
}

// WriteResponse converts channel, timestamp and packet data into the response data.
// The format is described in the documentation to ReadResponse.
func WriteResponse(resp *Response) ([]byte, error) {
	var buf bytes.Buffer
	if resp.Channel > math.MaxInt32 {
		return nil, fmt.Errorf("too large channel: %d. Max value: %d", resp.Channel, math.MaxInt32)
	}
	if len(resp.Data) > MaxSize {
		return nil, fmt.Errorf("too large data size: %d. Max data size: %d", len(resp.Data), MaxSize)
	}

	size := 12 + len(resp.Data)
	if err := binary.Write(&buf, binary.LittleEndian, uint32(size)); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, resp.Channel); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, resp.Timestamp.UnixNano()); err != nil {
		return nil, err
	}
	if _, err := buf.Write(resp.Data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
