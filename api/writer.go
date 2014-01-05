package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

// WriteResponse converts channel, timestamp and packet data into the response data.
// The format is:
// <uint32 size> <uint32 ch> <int64 ts> <data>
// where all multi-byte fields are low-endian.
// Size is the number of bytes in the rest of the message.
// Channel must be less than 1^31 = 2147483648.
// Timestamp is serialized as a number of nanoseconds since January 1, 1970 UTC.
// Data length must be not greater than MaxSize - 12.
func WriteResponse(ch uint32, ts time.Time, data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if ch > math.MaxInt32 {
		return nil, fmt.Errorf("too large channel: %d. Max value: %d", ch, math.MaxInt32)
	}
	if len(data) > MaxSize-12 {
		return nil, fmt.Errorf("too large data size: %d. Max data size: %d", len(data), MaxSize-12)
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
