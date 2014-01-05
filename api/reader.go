package api

import (
	"encoding/binary"
	"fmt"
	"io"
)

type CommandType int

const (
	Send   CommandType = 0
	Listen CommandType = 1

	MinSize    = 8
	MaxSize    = 128
	MaxChannel = 1<<31 - 1
)

type Command struct {
	Cmd     CommandType
	Channel uint32
	Data    []byte
}

// ReadRequest reads a request from the reader.
// The format is:
// <uint32 size> <uint32 cmd> <uint32 ch> <data>
// where all multi-byte fields are low-endian.
// Size is the number of bytes in the rest of the message. It must not exceed MaxSize.
// Channel must not exceeed MaxChannel.
func ReadRequest(r io.Reader) (cmd *Command, err error) {
	var size uint32
	if err = binary.Read(r, binary.LittleEndian, &size); err != nil {
		err = fmt.Errorf("could not read command body size: %v", err)
		return
	}
	if size > MaxSize {
		err = fmt.Errorf("command body size too large: %d. Max packet size: %d", size, MaxSize)
		return
	}
	if size < MinSize {
		err = fmt.Errorf("command body size too small: %d. Min packet size: %d", size, MinSize)
		return
	}
	data := make([]byte, size)
	if _, err = io.ReadFull(r, data); err != nil {
		data = nil
		err = fmt.Errorf("could not read command body (size: %d): %v", size, err)
		return
	}
	typ := CommandType(uint32(data[0]) + uint32(data[1])<<8 + uint32(data[2])<<16 + uint32(data[3])<<24)
	ch := uint32(data[4]) + uint32(data[5])<<8 + uint32(data[6])<<16 + uint32(data[7])<<24
	if ch > MaxChannel {
		data = nil
		err = fmt.Errorf("channel is too large: %d. Max channel: %d", ch, MaxChannel)
		return
	}
	return &Command{
		Cmd:     typ,
		Channel: ch,
		Data:    data[8:],
	}, nil
}
