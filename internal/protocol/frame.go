package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	HeaderLen  = 6
	MaxPayload = 64 * 1024
)

var ErrPayloadTooLarge = errors.New("payload too large")

type Packet struct {
	MsgID   uint16
	Payload []byte
}

func EncodePacket(packet Packet) ([]byte, error) {
	if len(packet.Payload) > MaxPayload {
		return nil, fmt.Errorf("%w: %d > %d", ErrPayloadTooLarge, len(packet.Payload), MaxPayload)
	}

	buf := make([]byte, HeaderLen+len(packet.Payload))
	binary.BigEndian.PutUint32(buf[:4], uint32(len(packet.Payload)))
	binary.BigEndian.PutUint16(buf[4:6], packet.MsgID)
	copy(buf[6:], packet.Payload)
	return buf, nil
}

func DecodePacket(r io.Reader) (Packet, error) {
	header := make([]byte, HeaderLen)
	if _, err := io.ReadFull(r, header); err != nil {
		return Packet{}, err
	}

	payloadLen := binary.BigEndian.Uint32(header[:4])
	msgID := binary.BigEndian.Uint16(header[4:6])
	if payloadLen > MaxPayload {
		return Packet{}, fmt.Errorf("%w: %d > %d", ErrPayloadTooLarge, payloadLen, MaxPayload)
	}

	payload := make([]byte, payloadLen)
	if payloadLen > 0 {
		if _, err := io.ReadFull(r, payload); err != nil {
			return Packet{}, err
		}
	}

	return Packet{
		MsgID:   msgID,
		Payload: payload,
	}, nil
}
