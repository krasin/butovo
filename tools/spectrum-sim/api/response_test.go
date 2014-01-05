package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"
)

const tsval = 0x1234567823456789

var timestamp = time.Unix(tsval/int64(1E9), tsval%int64(1E9))

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
			title: "Empty",
			err:   fmt.Errorf("could not read response body size: %v", io.EOF),
		},
		{
			title: "Too short",
			in:    []byte{100, 0, 0},
			err:   fmt.Errorf("could not read response body size: %v", io.ErrUnexpectedEOF),
		},
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
		{
			title: "max data len",
			in: append([]byte{140, 0, 0, 0, 4, 3, 2, 1,
				0x89, 0x67, 0x45, 0x23, 0x78, 0x56, 0x34, 0x12},
				make([]byte, 128)...),
			ch:   0x01020304,
			ts:   timestamp,
			data: make([]byte, 128),
		},
		{
			title: "too long",
			in:    []byte{200, 0, 0, 0},
			err:   errors.New("response body size too large: 200. Max packet size: 140"),
		},
	}
	for _, tt := range tests {
		resp, err := ReadResponse(bytes.NewBuffer(tt.in))
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", tt.err) {
			t.Errorf("%s: ReadResponse: unexpected err: %v, want: %v", err, tt.err)
			continue
		}
		if err != nil {
			continue
		}
		if resp.Channel != tt.ch {
			t.Errorf("%s: ch: %d, want: %d", tt.title, resp.Channel, tt.ch)
			continue
		}
		if resp.Timestamp != tt.ts {
			t.Errorf("%s: ts: %v, want: %v", tt.title, resp.Timestamp, tt.ts)
			continue
		}
		if !bytes.Equal(resp.Data, tt.data) {
			t.Errorf("%s: data:\n%v\nwant:\n%v", tt.title, resp.Data, tt.data)
			continue
		}
		out, err := WriteResponse(resp)
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
			err:   errors.New("too large data size: 200. Max data size: 128"),
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
		resp := &Response{
			Channel:   tt.ch,
			Timestamp: tt.ts,
			Data:      tt.data,
		}
		out, err := WriteResponse(resp)
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
