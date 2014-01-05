package api

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"
)

const tsval = 0x1234567823456789

var timestamp = time.Unix(tsval/int64(1E9), tsval%int64(1E9))

func TestWriteResponse(t *testing.T) {
	tests := []struct {
		title string
		ch    uint32
		ts    time.Time
		data  []byte
		out   []byte
		err   error
	}{
		{
			title: "Too large channel",
			ch:    1 << 31,
			err:   errors.New("too large channel: 2147483648. Max value: 2147483647"),
		},
		{
			title: "Too long data",
			ch:    23,
			ts:    timestamp,
			data:  make([]byte, 200),
			err:   errors.New("too large data size: 200. Max data size: 116"),
		},
		{
			title: "Some data",
			ch:    38,
			ts:    timestamp,
			data:  []byte("Hello"),
			out: []byte{17, 0, 0, 0,
				38, 0, 0, 0,
				0x89, 0x67, 0x45, 0x23, 0x78, 0x56, 0x34, 0x12,
				'H', 'e', 'l', 'l', 'o',
			},
		},
	}

	for _, tt := range tests {
		out, err := WriteResponse(tt.ch, tt.ts, tt.data)
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", tt.err) {
			t.Errorf("%s: Write: unexpected err: %v, want: %v", tt.title, err, tt.err)
			continue
		}
		if err != nil {
			continue
		}
		if !bytes.Equal(out, tt.out) {
			t.Errorf("%s: Write:\n%v\nwant:\n%v", tt.title, out, tt.out)
			continue
		}
	}
}
