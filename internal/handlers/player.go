package handlers

import (
	"context"
	"errors"
	"log"
	"strings"

	"openclaw-go/internal/messages"
	"openclaw-go/internal/player"
	"openclaw-go/internal/protocol"
	"openclaw-go/internal/router"
)

type CreateCharacterReq struct {
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
}

type CreateCharacterResp struct {
	Success   bool              `json:"success"`
	Error     string            `json:"error,omitempty"`
	Character *player.Character `json:"character,omitempty"`
}

type GetCharacterReq struct {
	PlayerID string `json:"player_id"`
}

type GetCharacterResp struct {
	Success   bool              `json:"success"`
	Error     string            `json:"error,omitempty"`
	Character *player.Character `json:"character,omitempty"`
}

func RegisterPlayer(r *router.Router, logger *log.Logger, store player.Store) {
	r.Register(messages.MsgCreateCharacterReq, func(ctx context.Context, session router.Session, packet protocol.Packet) error {
		var req CreateCharacterReq
		if err := protocol.UnmarshalJSON(packet.Payload, &req); err != nil {
			logger.Printf("create character unmarshal failed session=%s err=%v", session.ID(), err)
			return sendCreateCharacterResp(session, CreateCharacterResp{Success: false, Error: "invalid request payload"})
		}

		req.PlayerID = strings.TrimSpace(req.PlayerID)
		req.Name = strings.TrimSpace(req.Name)
		if req.PlayerID == "" || req.Name == "" {
			return sendCreateCharacterResp(session, CreateCharacterResp{Success: false, Error: "player_id and name are required"})
		}

		character, err := store.CreateCharacter(ctx, req.PlayerID, req.Name)
		if err != nil {
			switch {
			case errors.Is(err, player.ErrCharacterExists):
				return sendCreateCharacterResp(session, CreateCharacterResp{Success: false, Error: "character already exists"})
			default:
				logger.Printf("create character failed session=%s player_id=%s err=%v", session.ID(), req.PlayerID, err)
				return sendCreateCharacterResp(session, CreateCharacterResp{Success: false, Error: "internal error"})
			}
		}

		return sendCreateCharacterResp(session, CreateCharacterResp{
			Success:   true,
			Character: &character,
		})
	})

	r.Register(messages.MsgGetCharacterReq, func(ctx context.Context, session router.Session, packet protocol.Packet) error {
		var req GetCharacterReq
		if err := protocol.UnmarshalJSON(packet.Payload, &req); err != nil {
			logger.Printf("get character unmarshal failed session=%s err=%v", session.ID(), err)
			return sendGetCharacterResp(session, GetCharacterResp{Success: false, Error: "invalid request payload"})
		}

		req.PlayerID = strings.TrimSpace(req.PlayerID)
		if req.PlayerID == "" {
			return sendGetCharacterResp(session, GetCharacterResp{Success: false, Error: "player_id is required"})
		}

		character, err := store.GetCharacter(ctx, req.PlayerID)
		if err != nil {
			switch {
			case errors.Is(err, player.ErrCharacterNotFound):
				return sendGetCharacterResp(session, GetCharacterResp{Success: false, Error: "character not found"})
			default:
				logger.Printf("get character failed session=%s player_id=%s err=%v", session.ID(), req.PlayerID, err)
				return sendGetCharacterResp(session, GetCharacterResp{Success: false, Error: "internal error"})
			}
		}

		return sendGetCharacterResp(session, GetCharacterResp{
			Success:   true,
			Character: &character,
		})
	})
}

func sendCreateCharacterResp(session router.Session, resp CreateCharacterResp) error {
	payload, err := protocol.MarshalJSON(resp)
	if err != nil {
		return err
	}
	return session.Send(protocol.Packet{
		MsgID:   messages.MsgCreateCharacterResp,
		Payload: payload,
	})
}

func sendGetCharacterResp(session router.Session, resp GetCharacterResp) error {
	payload, err := protocol.MarshalJSON(resp)
	if err != nil {
		return err
	}
	return session.Send(protocol.Packet{
		MsgID:   messages.MsgGetCharacterResp,
		Payload: payload,
	})
}
