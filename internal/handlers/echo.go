package handlers

import (
	"context"
	"log"

	"openclaw-go/internal/messages"
	"openclaw-go/internal/protocol"
	"openclaw-go/internal/router"
)

type EchoReq struct {
	Text string `json:"text"`
}

type EchoResp struct {
	Text string `json:"text"`
}

func RegisterEcho(r *router.Router, logger *log.Logger) {
	r.Register(messages.MsgEchoReq, func(ctx context.Context, session router.Session, packet protocol.Packet) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var req EchoReq
		if err := protocol.UnmarshalJSON(packet.Payload, &req); err != nil {
			logger.Printf("echo unmarshal failed session=%s err=%v", session.ID(), err)
			return nil
		}

		respPayload, err := protocol.MarshalJSON(EchoResp{Text: req.Text})
		if err != nil {
			return err
		}

		return session.Send(protocol.Packet{
			MsgID:   messages.MsgEchoResp,
			Payload: respPayload,
		})
	})
}
