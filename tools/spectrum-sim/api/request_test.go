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
		typ   RequestType
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
			err:   errors.New("command body size too small: 2. Min packet size: 8"),
		},
		{
			title: "Too big",
			in:    []byte{200, 0, 0, 0},
			err:   errors.New("command body size too large: 200. Max packet size: 136"),
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
			title: "Send req",
			in:    []byte{13, 0, 0, 0, 0, 0, 0, 0, 37, 0, 0, 0, 'h', 'e', 'l', 'l', 'o'},
			typ:   Send,
			ch:    37,
			data:  []byte("hello"),
		},
		{
			title: "Send max data len",
			in:    append([]byte{136, 0, 0, 0, 0, 0, 0, 0, 37, 0, 0, 0}, make([]byte, 128)...),
			typ:   Send,
			ch:    37,
			data:  make([]byte, 128),
		},
		{
			title: "Listen req",
			in:    []byte{8, 0, 0, 0, 1, 0, 0, 0, 37, 0, 0, 0},
			typ:   Listen,
			ch:    37,
		},
	}

	for _, tt := range tests {
		req, err := ReadRequest(bytes.NewBuffer(tt.in))
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", tt.err) {
			t.Errorf("%q: unexpected error: %v, want: %v", tt.title, err, tt.err)
			continue
		}
		if err != nil {
			continue
		}
		if req.Type != tt.typ {
			t.Errorf("%q: typ: %d, want: %d", tt.title, req.Type, tt.typ)
			continue
		}
		if req.Channel != tt.ch {
			t.Errorf("%q: ch: %d, want: %d", tt.title, req.Channel, tt.ch)
			continue
		}
		if !bytes.Equal(req.Data, tt.data) {
			t.Errorf("%q, data: %v, want: %v", tt.title, req.Data, tt.data)
			continue
		}
		out, err := WriteRequest(req)
		if err != nil {
			t.Errorf("%q: WriteRequest(%+v): %v", tt.title, req, err)
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
