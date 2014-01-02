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
		cmd   Command
		ch    int
		data  []byte
		err   error
	}{
		{
			title: "Empty",
			err:   fmt.Errorf("Could not read command body size: %v", io.EOF),
		},
		{
			title: "Short size field",
			in:    []byte{100},
			err:   fmt.Errorf("Could not read command body size: %v", io.ErrUnexpectedEOF),
		},
		{
			title: "Too small",
			in:    []byte{2, 0, 0, 0},
			err:   fmt.Errorf("Command body size too small: %d. Min packet size: %d", 2, MinSize),
		},
		{
			title: "Too big",
			in:    []byte{200, 0, 0, 0},
			err:   fmt.Errorf("Command body size too large: %d. Max packet size: %d", 200, MaxSize),
		},
		{
			title: "No command body",
			in:    []byte{100, 0, 0, 0},
			err:   fmt.Errorf("Could not read command body (size: %d): %v", 100, io.EOF),
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
		cmd, ch, data, err := Read(bytes.NewBuffer(tt.in))
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", tt.err) {
			t.Errorf("%q: unexpected error: %v, want: %v", tt.title, err, tt.err)
			continue
		}
		if err != nil {
			continue
		}
		if cmd != tt.cmd {
			t.Errorf("%q: cmd: %d, want: %d", tt.title, cmd, tt.cmd)
			continue
		}
		if ch != tt.ch {
			t.Errorf("%q: ch: %d, want: %d", tt.title, ch, tt.ch)
			continue
		}
		if !bytes.Equal(data, tt.data) {
			t.Errorf("%q, data: %v, want: %v", tt.title, data, tt.data)
			continue
		}
	}
}
