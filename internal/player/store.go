package player

import (
	"context"
	"errors"
)

var (
	ErrCharacterExists   = errors.New("character already exists")
	ErrCharacterNotFound = errors.New("character not found")
)

type Character struct {
	ID          int64  `json:"id"`
	PlayerID    string `json:"player_id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Exp         int64  `json:"exp"`
	CreatedAtMS int64  `json:"created_at_ms"`
	UpdatedAtMS int64  `json:"updated_at_ms"`
}

type Store interface {
	CreateCharacter(ctx context.Context, playerID string, name string) (Character, error)
	GetCharacter(ctx context.Context, playerID string) (Character, error)
	Close() error
}
