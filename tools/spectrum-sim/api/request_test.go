package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
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

func TestWriteRequest(t *testing.T) {
	tests := []struct {
		title string
		typ   RequestType
		ch    uint32
		data  []byte
		out   []byte
		err   error
	}{
		{
			title: "Listen",
			typ:   Listen,
			ch:    258,
			out: []byte{8, 0, 0, 0,
				1, 0, 0, 0,
				2, 1, 0, 0},
		},
	}
	for _, tt := range tests {
		req := &Request{
			Type:    tt.typ,
			Channel: tt.ch,
			Data:    tt.data,
		}
		out, err := WriteRequest(req)
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", tt.err) {
			t.Errorf("%s: WriteRequest: unexpected err: %v, want: %v", tt.title, err, tt.err)
			continue
		}
		if err != nil {
			continue
		}
		if !bytes.Equal(out, tt.out) {
			t.Errorf("%s: unexpected out:\n%v\nwant:\n%v", out, tt.out)
			continue
		}
	}
}
