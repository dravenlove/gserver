CREATE DATABASE IF NOT EXISTS gserver CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE gserver;

CREATE TABLE IF NOT EXISTS player_characters (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  player_id VARCHAR(64) NOT NULL UNIQUE,
  name VARCHAR(64) NOT NULL,
  level INT NOT NULL DEFAULT 1,
  exp BIGINT NOT NULL DEFAULT 0,
  created_at_ms BIGINT NOT NULL,
  updated_at_ms BIGINT NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_player_characters_player_id ON player_characters(player_id);
