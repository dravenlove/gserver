package protocol

import (
	"bytes"
	"errors"
	"testing"
)

func TestEncodeDecodePacket(t *testing.T) {
	origin := Packet{
		MsgID:   1001,
		Payload: []byte(`{"hello":"world"}`),
	}

	data, err := EncodePacket(origin)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	got, err := DecodePacket(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if got.MsgID != origin.MsgID {
		t.Fatalf("msg_id mismatch got=%d want=%d", got.MsgID, origin.MsgID)
	}
	if !bytes.Equal(got.Payload, origin.Payload) {
		t.Fatalf("payload mismatch got=%q want=%q", got.Payload, origin.Payload)
	}
}

func TestEncodePacketPayloadTooLarge(t *testing.T) {
	_, err := EncodePacket(Packet{
		MsgID:   1,
		Payload: make([]byte, MaxPayload+1),
	})
	if !errors.Is(err, ErrPayloadTooLarge) {
		t.Fatalf("unexpected err: %v", err)
	}
}
