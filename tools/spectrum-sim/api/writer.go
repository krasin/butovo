package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

// WriteRequest converts a command into the request data.
// The format is described in the documentation to ReadRequest.
func WriteRequest(cmd *Command) ([]byte, error) {
	var buf bytes.Buffer

	if len(cmd.Data) > MaxSize {
		return nil, fmt.Errorf("too large data size: %d. Max data size: %d", len(cmd.Data), MaxSize)
	}
	size := 8 + len(cmd.Data)
	if err := binary.Write(&buf, binary.LittleEndian, uint32(size)); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, uint32(cmd.Cmd)); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, uint32(cmd.Channel)); err != nil {
		return nil, err
	}
	if _, err := buf.Write(cmd.Data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WriteResponse converts channel, timestamp and packet data into the response data.
// The format is described in the documentation to ReadResponse.
func WriteResponse(ch uint32, ts time.Time, data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if ch > math.MaxInt32 {
		return nil, fmt.Errorf("too large channel: %d. Max value: %d", ch, math.MaxInt32)
	}
	if len(data) > MaxSize {
		return nil, fmt.Errorf("too large data size: %d. Max data size: %d", len(data), MaxSize)
	}

	size := 12 + len(data)
	if err := binary.Write(&buf, binary.LittleEndian, uint32(size)); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, ch); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, ts.UnixNano()); err != nil {
		return nil, err
	}
	if _, err := buf.Write(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
