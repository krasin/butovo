package sim

import (
	"bytes"
	"net"
	"testing"

	"github.com/krasin/butovo/tools/spectrum-sim/api"
)

func TestServer(t *testing.T) {
	errChan := make(chan error, 1)
	s := NewServer(errChan)

	// Subscribe to listen a channel
	var ch uint32 = 37
	lis1, lis2 := net.Pipe()
	go s.Handle(lis2)

	reqData, err := api.WriteRequest(&api.Request{Channel: ch, Type: api.Listen})
	if err != nil {
		t.Fatal("WriteRequest: ", err)
	}
	if _, err := lis1.Write(reqData); err != nil {
		t.Fatal("Write(reqData): ", err)
	}

	// Send payload to the channel
	send1, send2 := net.Pipe()
	go s.Handle(send2)
	payload := []byte("foo")
	if reqData, err = api.WriteRequest(&api.Request{Channel: ch, Type: api.Send, Data: payload}); err != nil {
		t.Fatal("WriteRequest: ", err)
	}
	if _, err := send1.Write(reqData); err != nil {
		t.Fatal("Write(reqData): ", err)
	}
	if err = send1.Close(); err != nil {
		t.Fatal("send1.Close(): ", err)
	}

	// Read response to the listener
	resp, err := api.ReadResponse(lis1)
	if err != nil {
		t.Fatal("ReadResponse: ", err)
	}
	if err = lis1.Close(); err != nil {
		t.Fatal("lis1.Close()")
	}

	// Check that the response matches
	if resp.Channel != ch {
		t.Fatalf("Unexpected channel: %d, want: %d", resp.Channel, ch)
	}
	if !bytes.Equal(resp.Data, payload) {
		t.Fatalf("Unexpected payload: %v, want: %v", resp.Data, payload)
	}

	// Make sure, there was no errors
	select {
	case err := <-errChan:
		t.Fatal("Error reported via errChan:", err)
	default:
	}
}
