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
