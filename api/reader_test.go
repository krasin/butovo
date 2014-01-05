package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"
)

func TestReadRequest(t *testing.T) {
	tests := []struct {
		title string
		in    []byte
		err   error
		cmd   CommandType
		ch    uint32
		data  []byte
	}{
		{
			title: "Empty",
			err:   fmt.Errorf("could not read command body size: %v", io.EOF),
		},
		{
			title: "Short size field",
			in:    []byte{100},
			err:   fmt.Errorf("could not read command body size: %v", io.ErrUnexpectedEOF),
		},
		{
			title: "Too small",
			in:    []byte{2, 0, 0, 0},
			err:   fmt.Errorf("command body size too small: %d. Min packet size: %d", 2, MinSize),
		},
		{
			title: "Too big",
			in:    []byte{200, 0, 0, 0},
			err:   fmt.Errorf("command body size too large: %d. Max packet size: %d", 200, MaxSize),
		},
		{
			title: "No command body",
			in:    []byte{100, 0, 0, 0},
			err:   fmt.Errorf("could not read command body (size: %d): %v", 100, io.EOF),
		},
		{
			title: "Too large channel",
			in:    []byte{8, 0, 0, 0, 1, 0, 0, 0, 255, 255, 255, 255},
			err:   errors.New("channel is too large: 4294967295. Max channel: 2147483647"),
		},
		{
			title: "Send cmd",
			in:    []byte{13, 0, 0, 0, 0, 0, 0, 0, 37, 0, 0, 0, 'h', 'e', 'l', 'l', 'o'},
			cmd:   Send,
			ch:    37,
			data:  []byte("hello"),
		},
		{
			title: "Listen cmd",
			in:    []byte{8, 0, 0, 0, 1, 0, 0, 0, 37, 0, 0, 0},
			cmd:   Listen,
			ch:    37,
		},
	}

	for _, tt := range tests {
		cmd, err := ReadRequest(bytes.NewBuffer(tt.in))
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", tt.err) {
			t.Errorf("%q: unexpected error: %v, want: %v", tt.title, err, tt.err)
			continue
		}
		if err != nil {
			continue
		}
		if cmd.Cmd != tt.cmd {
			t.Errorf("%q: cmd: %d, want: %d", tt.title, cmd.Cmd, tt.cmd)
			continue
		}
		if cmd.Channel != tt.ch {
			t.Errorf("%q: ch: %d, want: %d", tt.title, cmd.Channel, tt.ch)
			continue
		}
		if !bytes.Equal(cmd.Data, tt.data) {
			t.Errorf("%q, data: %v, want: %v", tt.title, cmd.Data, tt.data)
			continue
		}
		out, err := WriteRequest(cmd)
		if err != nil {
			t.Errorf("%q: WriteRequest(%+v): %v", tt.title, cmd, err)
			continue
		}
		if !bytes.Equal(out, tt.in) {
			t.Errorf("%q: WriteRequest(%+v): %v, want: %v", tt.title, out, tt.in)
			continue
		}
	}
}

func TestReadResponse(t *testing.T) {
	tests := []struct {
		title string
		in    []byte
		ch    uint32
		ts    time.Time
		data  []byte
		err   error
	}{
		{
			title: "hello",
			in: []byte{17, 0, 0, 0,
				4, 1, 0, 0,
				0x89, 0x67, 0x45, 0x23, 0x78, 0x56, 0x34, 0x12,
				'h', 'e', 'l', 'l', 'o'},
			ch:   260,
			ts:   timestamp,
			data: []byte("hello"),
		},
	}
	for _, tt := range tests {
		ch, ts, data, err := ReadResponse(bytes.NewBuffer(tt.in))
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", tt.err) {
			t.Errorf("%s: ReadResponse: unexpected err: %v, want: %v", err, tt.err)
			continue
		}
		if ch != tt.ch {
			t.Errorf("%s: ch: %d, want: %d", tt.title, ch, tt.ch)
			continue
		}
		if ts != tt.ts {
			t.Errorf("%s: ts: %v, want: %v", tt.title, ts, tt.ts)
			continue
		}
		if !bytes.Equal(data, tt.data) {
			t.Errorf("%s: data:\n%v\nwant:\n%v", tt.title, data, tt.data)
			continue
		}
		out, err := WriteResponse(ch, ts, data)
		if err != nil {
			t.Errorf("%s: WriteResponse: %v", tt.title, err)
			continue
		}
		if !bytes.Equal(out, tt.in) {
			t.Errorf("%s: WriteResponse: %v, want: %v", tt.title, out, tt.in)
			continue
		}
	}
}
