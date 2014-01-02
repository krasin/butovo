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

	MinSize = 8
	MaxSize = 128
)

type Command struct {
	Cmd     CommandType
	Channel int
	Data    []byte
}

func Read(r io.Reader) (cmd *Command, err error) {
	var size uint32
	if err = binary.Read(r, binary.LittleEndian, &size); err != nil {
		err = fmt.Errorf("Could not read command body size: %v", err)
		return
	}
	if size > MaxSize {
		err = fmt.Errorf("Command body size too large: %d. Max packet size: %d", size, MaxSize)
		return
	}
	if size < MinSize {
		err = fmt.Errorf("Command body size too small: %d. Min packet size: %d", size, MinSize)
		return
	}
	data := make([]byte, size)
	if _, err = io.ReadFull(r, data); err != nil {
		data = nil
		err = fmt.Errorf("Could not read command body (size: %d): %v", size, err)
		return
	}
	typ := CommandType(uint32(data[0]) + uint32(data[1])<<8 + uint32(data[2])<<16 + uint32(data[3])<<24)
	ch := int(data[4]) + int(data[5])<<8 + int(data[6])<<16 + int(data[7])<<24
	return &Command{
		Cmd:     typ,
		Channel: ch,
		Data:    data[8:],
	}, nil
}
