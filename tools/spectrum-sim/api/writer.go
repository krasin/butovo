package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

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
