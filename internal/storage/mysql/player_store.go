package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"

	"openclaw-go/internal/player"
)

type PlayerStore struct {
	db *sql.DB
}

func NewPlayerStore(dsn string) (*PlayerStore, error) {
	if dsn == "" {
		return nil, errors.New("mysql dsn is empty")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &PlayerStore{db: db}, nil
}

func (s *PlayerStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *PlayerStore) CreateCharacter(ctx context.Context, playerID string, name string) (player.Character, error) {
	nowMS := time.Now().UnixMilli()
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO player_characters(player_id, name, level, exp, created_at_ms, updated_at_ms)
		 VALUES(?, ?, 1, 0, ?, ?)`,
		playerID, name, nowMS, nowMS,
	)
	if err != nil {
		if isDuplicateEntry(err) {
			return player.Character{}, player.ErrCharacterExists
		}
		return player.Character{}, fmt.Errorf("insert character: %w", err)
	}

	return s.GetCharacter(ctx, playerID)
}

func (s *PlayerStore) GetCharacter(ctx context.Context, playerID string) (player.Character, error) {
	const q = `SELECT id, player_id, name, level, exp, created_at_ms, updated_at_ms
			   FROM player_characters WHERE player_id = ?`

	var c player.Character
	if err := s.db.QueryRowContext(ctx, q, playerID).Scan(
		&c.ID,
		&c.PlayerID,
		&c.Name,
		&c.Level,
		&c.Exp,
		&c.CreatedAtMS,
		&c.UpdatedAtMS,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return player.Character{}, player.ErrCharacterNotFound
		}
		return player.Character{}, fmt.Errorf("query character: %w", err)
	}
	return c, nil
}

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS player_characters (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			player_id VARCHAR(64) NOT NULL UNIQUE,
			name VARCHAR(64) NOT NULL,
			level INT NOT NULL DEFAULT 1,
			exp BIGINT NOT NULL DEFAULT 0,
			created_at_ms BIGINT NOT NULL,
			updated_at_ms BIGINT NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE INDEX idx_player_characters_player_id
			ON player_characters(player_id);`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			// Ignore duplicate index creation to keep startup idempotent.
			if !isDuplicateKeyName(err) {
				return fmt.Errorf("migrate mysql: %w", err)
			}
		}
	}
	return nil
}

func isDuplicateEntry(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func isDuplicateKeyName(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1061
}
