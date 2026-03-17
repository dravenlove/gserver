package protocol_test

import (
	"bytes"
	"errors"
	"testing"

	"openclaw-go/internal/protocol"
)

func TestEncodeDecodePacket(t *testing.T) {
	origin := protocol.Packet{
		MsgID:   1001,
		Payload: []byte(`{"hello":"world"}`),
	}

	data, err := protocol.EncodePacket(origin)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	got, err := protocol.DecodePacket(bytes.NewReader(data))
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
	_, err := protocol.EncodePacket(protocol.Packet{
		MsgID:   1,
		Payload: make([]byte, protocol.MaxPayload+1),
	})
	if !errors.Is(err, protocol.ErrPayloadTooLarge) {
		t.Fatalf("unexpected err: %v", err)
	}
}
