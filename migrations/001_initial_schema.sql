-- Gwent Bot Database Schema
-- Migration 001: Initial Schema

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- PLAYERS TABLE
-- ============================================================================
CREATE TABLE players (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    level SMALLINT NOT NULL DEFAULT 1,
    experience BIGINT NOT NULL DEFAULT 0,
    gold BIGINT NOT NULL DEFAULT 100,
    scrap BIGINT NOT NULL DEFAULT 0,
    meteorite_powder BIGINT NOT NULL DEFAULT 0,
    collection JSONB NOT NULL DEFAULT '[]'::jsonb,
    deck JSONB NOT NULL DEFAULT '[]'::jsonb,
    leader_id VARCHAR(100),
    daily_coin_claimed_at TIMESTAMPTZ,
    axe_claimed_at TIMESTAMPTZ,
    axe_durability SMALLINT NOT NULL DEFAULT 0,
    mine_unlocked BOOLEAN NOT NULL DEFAULT FALSE,
    mine_extracts_today SMALLINT NOT NULL DEFAULT 0,
    last_mine_extract_at TIMESTAMPTZ,
    games_played INT NOT NULL DEFAULT 0,
    games_won INT NOT NULL DEFAULT 0,
    games_lost INT NOT NULL DEFAULT 0,
    highest_rank INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_players_telegram_id ON players(telegram_id);
CREATE INDEX idx_players_level ON players(level);

-- ============================================================================
-- CARDS TABLE
-- ============================================================================
CREATE TABLE cards (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    nation VARCHAR(50) NOT NULL,
    rarity VARCHAR(20) NOT NULL,
    type VARCHAR(20) NOT NULL,
    row VARCHAR(20),
    provision SMALLINT NOT NULL,
    power SMALLINT NOT NULL,
    armor SMALLINT NOT NULL DEFAULT 0,
    abilities JSONB NOT NULL DEFAULT '[]'::jsonb,
    art_url VARCHAR(500),
    flavor_text TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cards_nation ON cards(nation);
CREATE INDEX idx_cards_rarity ON cards(rarity);
CREATE INDEX idx_cards_type ON cards(type);
CREATE INDEX idx_cards_provision ON cards(provision);

-- ============================================================================
-- BATTLES TABLE
-- ============================================================================
CREATE TABLE battles (
    id VARCHAR(100) PRIMARY KEY,
    type VARCHAR(20) NOT NULL,
    phase VARCHAR(20) NOT NULL DEFAULT 'waiting',
    current_round SMALLINT NOT NULL DEFAULT 1,
    current_turn BIGINT,
    player1_id BIGINT NOT NULL REFERENCES players(id),
    player2_id BIGINT REFERENCES players(id),
    winner_id BIGINT REFERENCES players(id),
    player1_board JSONB NOT NULL DEFAULT '{"melee":[],"ranged":[],"siege":[],"total_power":0}'::jsonb,
    player2_board JSONB NOT NULL DEFAULT '{"melee":[],"ranged":[],"siege":[],"total_power":0}'::jsonb,
    player1_hand JSONB NOT NULL DEFAULT '[]'::jsonb,
    player2_hand JSONB NOT NULL DEFAULT '[]'::jsonb,
    player1_deck JSONB NOT NULL DEFAULT '[]'::jsonb,
    player2_deck JSONB NOT NULL DEFAULT '[]'::jsonb,
    player1_graveyard JSONB NOT NULL DEFAULT '[]'::jsonb,
    player2_graveyard JSONB NOT NULL DEFAULT '[]'::jsonb,
    player1_leader_used BOOLEAN NOT NULL DEFAULT FALSE,
    player2_leader_used BOOLEAN NOT NULL DEFAULT FALSE,
    player1_passed BOOLEAN NOT NULL DEFAULT FALSE,
    player2_passed BOOLEAN NOT NULL DEFAULT FALSE,
    weather JSONB NOT NULL DEFAULT '{"active":"clear","duration":0,"row":null}'::jsonb,
    bot_difficulty VARCHAR(20),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX idx_battles_player1 ON battles(player1_id);
CREATE INDEX idx_battles_player2 ON battles(player2_id);
CREATE INDEX idx_battles_phase ON battles(phase);

-- ============================================================================
-- CHESTS TABLE
-- ============================================================================
CREATE TABLE chests (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    level_required SMALLINT NOT NULL,
    cards JSONB NOT NULL DEFAULT '[]'::jsonb,
    gold BIGINT NOT NULL DEFAULT 0,
    cooldown INTERVAL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- SHOP ITEMS TABLE
-- ============================================================================
CREATE TABLE shop_items (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    card_id VARCHAR(100) REFERENCES cards(id),
    cost_gold BIGINT NOT NULL DEFAULT 0,
    cost_scrap BIGINT NOT NULL DEFAULT 0,
    stock INT NOT NULL DEFAULT -1,
    level_required SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shop_level ON shop_items(level_required);

-- ============================================================================
-- GAME SETTINGS TABLE
-- ============================================================================
CREATE TABLE game_settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- PLAYER INVENTORY (for tracking card quantities)
-- ============================================================================
CREATE TABLE player_cards (
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    card_id VARCHAR(100) NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 1,
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (player_id, card_id)
);

CREATE INDEX idx_player_cards_player ON player_cards(player_id);

-- ============================================================================
-- BATTLE LOG (for replay/history)
-- ============================================================================
CREATE TABLE battle_logs (
    id BIGSERIAL PRIMARY KEY,
    battle_id VARCHAR(100) NOT NULL REFERENCES battles(id) ON DELETE CASCADE,
    round_num SMALLINT NOT NULL,
    turn_player_id BIGINT NOT NULL REFERENCES players(id),
    action_type VARCHAR(50) NOT NULL,
    action_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_battle_logs_battle ON battle_logs(battle_id);

-- ============================================================================
-- FRIEND INVITES (for PvP with friends)
-- ============================================================================
CREATE TABLE friend_invites (
    id BIGSERIAL PRIMARY KEY,
    from_player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    to_player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    battle_id VARCHAR(100) REFERENCES battles(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    UNIQUE(from_player_id, to_player_id, status)
);

CREATE INDEX idx_friend_invites_to ON friend_invites(to_player_id, status);
CREATE INDEX idx_friend_invites_from ON friend_invites(from_player_id, status);
