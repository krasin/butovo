package api

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestRead(t *testing.T) {
	tests := []struct {
		title string
		in    []byte
		err   error
		cmd   CommandType
		ch    int
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
			title: "Negative channel",
			in:    []byte{8, 0, 0, 0, 1, 0, 0, 0, 255, 255, 255, 255},
			err:   fmt.Errorf("channel is negative: -1"),
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
		cmd, err := Read(bytes.NewBuffer(tt.in))
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", tt.err) {
			t.Errorf("%q: unexpected error: %v, ch: %d, want: %v", tt.title, err, cmd.Channel, tt.err)
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
	}
}
