package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

func WriteResponse(ch uint32, ts time.Time, data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if ch > math.MaxInt32 {
		return nil, fmt.Errorf("too large channel: %d. Max value: %d", ch, math.MaxInt32)
	}
	if len(data) > MaxSize-8 {
		return nil, fmt.Errorf("too large data size: %d. Max data size: %d", len(data), MaxSize-8)
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
