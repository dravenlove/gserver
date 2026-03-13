package handlers

import (
	"context"
	"log"
	"time"

	"openclaw-go/internal/messages"
	"openclaw-go/internal/protocol"
	"openclaw-go/internal/router"
)

type PingReq struct {
	ClientTime int64 `json:"client_time"`
}

type PingResp struct {
	ServerTime int64  `json:"server_time"`
	Message    string `json:"message"`
}

func RegisterPing(r *router.Router, logger *log.Logger) {
	r.Register(messages.MsgPingReq, func(ctx context.Context, session router.Session, packet protocol.Packet) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var req PingReq
		if len(packet.Payload) > 0 {
			if err := protocol.UnmarshalJSON(packet.Payload, &req); err != nil {
				logger.Printf("ping unmarshal failed session=%s err=%v", session.ID(), err)
				return nil
			}
		}

		respPayload, err := protocol.MarshalJSON(PingResp{
			ServerTime: time.Now().UnixMilli(),
			Message:    "pong",
		})
		if err != nil {
			return err
		}

		logger.Printf("ping request session=%s client_time=%d", session.ID(), req.ClientTime)
		return session.Send(protocol.Packet{
			MsgID:   messages.MsgPingResp,
			Payload: respPayload,
		})
	})
}
