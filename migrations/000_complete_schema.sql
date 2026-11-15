-- Gwent Bot - Complete Mini-Gwent Database Schema (Witcher 3 Style)
-- Single consolidated migration replacing 001, 002, 003
-- Run this on fresh database only

-- ============================================================================
-- EXTENSIONS
-- ============================================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- PLAYERS TABLE (Mini-Gwent simplified)
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
    leader_id VARCHAR(100),
    leader_ability VARCHAR(100),           -- passive leader ability
    cards_won INT NOT NULL DEFAULT 0,
    opponents_defeated JSONB NOT NULL DEFAULT '[]'::jsonb,
    deck JSONB NOT NULL DEFAULT '[]'::jsonb,
    collection JSONB NOT NULL DEFAULT '[]'::jsonb,
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
-- CARDS TABLE (Mini-Gwent schema)
-- ============================================================================
CREATE TABLE cards (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    faction VARCHAR(50) NOT NULL,          -- northern_realms, nilfgaard, scoiatael, monsters, neutral
    type VARCHAR(20) NOT NULL,             -- unit, special, hero
    row VARCHAR(20),                       -- close/melee, ranged, siege, NULL for specials
    power SMALLINT NOT NULL DEFAULT 0,
    is_hero BOOLEAN NOT NULL DEFAULT FALSE,
    abilities JSONB NOT NULL DEFAULT '[]'::jsonb,
    art_url VARCHAR(500),
    flavor_text TEXT,
    card_group VARCHAR(50),                -- close, ranged, siege, special, hero
    rarity VARCHAR(20),                    -- common, rare, epic, legendary (nullable for Mini-Gwent)
    cost_gold BIGINT NOT NULL DEFAULT 0,
    cost_scrap BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Legacy column for backward compatibility (code sets Nation = Faction after reading)
ALTER TABLE cards ADD COLUMN IF NOT EXISTS nation VARCHAR(50);

CREATE INDEX idx_cards_faction ON cards(faction);
CREATE INDEX idx_cards_type ON cards(type);
CREATE INDEX idx_cards_card_group ON cards(card_group);
CREATE INDEX idx_cards_rarity ON cards(rarity);

-- ============================================================================
-- BATTLES TABLE
-- ============================================================================
CREATE TABLE battles (
    id VARCHAR(100) PRIMARY KEY,
    type VARCHAR(20) NOT NULL,             -- pve, pvp, friend
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
-- SHOP ITEMS TABLE (no FK to cards - cards seeded after)
-- ============================================================================
CREATE TABLE shop_items (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    card_id VARCHAR(100),                  -- references cards(id) but no FK constraint
    cost_gold BIGINT NOT NULL DEFAULT 0,
    cost_scrap BIGINT NOT NULL DEFAULT 0,
    stock INT NOT NULL DEFAULT -1,         -- -1 = unlimited
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
-- PLAYER CARDS (for tracking card quantities in collection)
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
-- BATTLE LOGS (for replay/history)
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

-- ============================================================================
-- KEYWORDS REFERENCE TABLE
-- ============================================================================
CREATE TABLE minigwent_keywords (
    keyword VARCHAR(30) PRIMARY KEY,
    description TEXT
);

-- ============================================================================
-- SEED KEYWORDS
-- ============================================================================
INSERT INTO minigwent_keywords (keyword, description) VALUES
('spy', 'Place on opponent''s side, draw 2 cards'),
('medic', 'Revive a non-hero unit from your graveyard'),
('morale', 'Boost all non-hero units on this row by 1'),
('tight_bond', 'Boost all units with same name by their base power'),
('muster', 'Play all units with same name from deck'),
('commanders_horn', 'Boost all non-hero units on a row by 3'),
('scorch', 'Destroy highest power non-hero unit(s) on board'),
('hero', 'Immune to weather, scorch, and special cards'),
('weather_frost', 'Sets melee row power to 1 for both sides'),
('weather_fog', 'Sets ranged row power to 1 for both sides'),
('weather_rain', 'Sets siege row power to 1 for both sides'),
('clear_weather', 'Removes all weather effects'),
('leader', 'Leader ability - once per battle')
ON CONFLICT (keyword) DO NOTHING;

-- ============================================================================
-- SEED CARDS (85 cards from Mini-Gwent Witcher 3)
-- ============================================================================

-- NORTHERN REALMS (20 cards)
INSERT INTO cards (id, name, description, faction, type, row, power, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap) VALUES
-- Heroes (Leader cards)
('nr_foltest_siegemaster', 'Foltest, The Siegemaster', 'Leader: +1 power to all Siege units', 'northern_realms', 'hero', 'close', 6, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Esterad Thyssen.jpg', 'The North remembers.', 'hero', 'legendary', 0, 0),
('nr_foltest_commander', 'Foltest, King of Temeria', 'Leader: +1 power to all Close units', 'northern_realms', 'hero', 'close', 5, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Barclay Els.jpg', 'Temeria stands.', 'hero', 'legendary', 0, 0),
('nr_foltest_royal', 'Foltest, The Royal', 'Leader: +1 power to all Ranged units', 'northern_realms', 'hero', 'close', 4, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Prince Stennis.jpg', 'By blood and steel.', 'hero', 'legendary', 0, 0),
('nr_vernon_roche', 'Vernon Roche', 'Commander of the Blue Stripes', 'northern_realms', 'hero', 'close', 6, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Vernon Roche.jpg', 'Temeria will not fall while I draw breath.', 'hero', 'legendary', 0, 0),
('nr_siegfried', 'Siegfried of Denesle', 'Order of the Flaming Rose', 'northern_realms', 'hero', 'close', 5, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Sieg fried of Denesle.jpg', 'By fire and sword.', 'hero', 'legendary', 0, 0),

-- Gold/Special Cards (high power units)
('nr_esterad_thyssen', 'Esterad Thyssen', 'King of Kovir and Poviss', 'northern_realms', 'unit', 'close', 7, false, '["morale"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Esterad Thyssen.jpg', 'Wealth is power.', 'close', 'legendary', 0, 0),
('nr_hjalmar', 'Hjalmar', 'Son of Crach an Craite', 'northern_realms', 'unit', 'close', 6, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Hjalmar.jpg', 'Glory awaits.', 'close', 'epic', 0, 0),
('nr_ciri', 'Cirilla', 'Child of Destiny', 'northern_realms', 'hero', 'close', 15, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Yeneffer.jpg', 'Time and space.', 'hero', 'legendary', 0, 0),
('nr_yennefer', 'Yennefer', 'Sorceress of Vengerberg', 'northern_realms', 'hero', 'close', 7, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/philippa Eilhart.jpg', 'Power and beauty.', 'hero', 'legendary', 0, 0),
('nr_triss', 'Triss Merigold', 'Sorceress', 'northern_realms', 'hero', 'close', 7, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Sabrina Glevissig.jpg', 'Fire and faith.', 'hero', 'legendary', 0, 0),

-- Silver/Strong Units
('nr_blue_stripes_commando', 'Blue Stripes Commando', 'Elite Temerian unit', 'northern_realms', 'unit', 'close', 4, false, '["tight_bond"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Blue Stripes Commando.jpg', 'Discipline wins wars.', 'close', 'rare', 100, 20),
('nr_crinfrid_reaver', 'Crinfrid Reaver', 'Dragon hunters', 'northern_realms', 'unit', 'close', 4, false, '["tight_bond"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Crinfrid Reavers.jpg', 'Gold for scales.', 'close', 'rare', 80, 15),
('nr_kaedweni_soldier', 'Kaedweni Soldier', 'Sturdy spearman', 'northern_realms', 'unit', 'close', 3, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Kaed weni Sige.jpg', 'Steel and resolve.', 'close', 'common', 50, 10),
('nr_dun_banner', 'Dun Banner Heavy Cavalry', 'Armored charge', 'northern_realms', 'unit', 'close', 5, false, '["morale"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Impra Brigade Guard.jpg', 'Charge!', 'close', 'rare', 120, 25),
('nr_redanian_elite', 'Redanian Elite', 'Armored crossbowman', 'northern_realms', 'unit', 'ranged', 5, false, '["morale"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Redanian Foot soldier.jpg', 'For Radovid.', 'ranged', 'rare', 100, 20),
('nr_blue_stripes_scout', 'Blue Stripes Scout', 'Recon unit', 'northern_realms', 'unit', 'ranged', 2, false, '["spy"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Etolian Auxiliary Archers.jpg', 'Eyes everywhere.', 'ranged', 'common', 80, 15),
('nr_arbalist', 'Arbalist', 'Crossbow specialist', 'northern_realms', 'unit', 'ranged', 4, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Black Infantry Archer.jpg', 'Pierce armor.', 'ranged', 'common', 60, 10),
('nr_archer', 'Archer', 'Longbow expert', 'northern_realms', 'unit', 'ranged', 3, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Elven Skirmisher.jpg', 'Aim true.', 'ranged', 'common', 50, 10),
('nr_ballista', 'Ballista', 'Siege engine', 'northern_realms', 'unit', 'siege', 6, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Balista.jpg', 'Rain fire.', 'siege', 'rare', 120, 25),
('nr_catapult', 'Catapult', 'Heavy artillery', 'northern_realms', 'unit', 'siege', 5, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Catapult.jpg', 'Smash walls.', 'siege', 'common', 80, 15),
('nr_trebuchet', 'Trebuchet', 'Long-range siege', 'northern_realms', 'unit', 'siege', 8, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Trebuchet.jpg', 'From afar.', 'siege', 'epic', 150, 30),

-- Special Cards
('nr_commanders_horn', 'Commander''s Horn', 'Boost all units on a row by 3', 'northern_realms', 'special', NULL, 0, false, '["commanders_horn"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'For the North!', 'special', 'rare', 0, 0),
('nr_clear_weather', 'Clear Weather', 'Remove all weather effects', 'northern_realms', 'special', NULL, 0, false, '["clear_weather"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Sun returns.', 'special', 'common', 0, 0),
('nr_scorch', 'Scorch', 'Destroy highest power non-hero unit(s)', 'northern_realms', 'special', NULL, 0, false, '["scorch"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Burn them all.', 'special', 'rare', 0, 0),
('nr_decoy', 'Decoy', 'Return a unit from your board to hand', 'northern_realms', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Trickery.', 'special', 'common', 0, 0),
('nr_mardroeme', 'Mardroeme', 'Transform a unit into a base copy of another', 'northern_realms', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Cursed flesh.', 'special', 'rare', 0, 0);

-- NILFGAARD (18 cards)
INSERT INTO cards (id, name, description, faction, type, row, power, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap) VALUES
-- Heroes
('ng_emhyr', 'Emhyr var Emreis', 'Emperor of Nilfgaard', 'nilfgaard', 'hero', 'close', 1, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Morvran voorhis.jpg', 'Empire comes first.', 'hero', 'legendary', 0, 0),
('ng_menno', 'Menno Coehoorn', 'Siege master', 'nilfgaard', 'hero', 'close', 10, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Menno Coehoorn.jpg', 'Engineer of war.', 'hero', 'legendary', 0, 0),
('ng_vattier', 'Vattier de Rideaux', 'Spymaster', 'nilfgaard', 'hero', 'close', 1, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Vattier de Rideaux.jpg', 'Information is power.', 'hero', 'legendary', 0, 0),
('ng_cahir', 'Cahir Mawr Dyffryn aep Ceallach', 'The Black Knight', 'nilfgaard', 'hero', 'close', 5, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Cahir Mawr.jpg', 'Destiny guides my blade.', 'hero', 'legendary', 0, 0),
('ng_stefan_skellen', 'Stefan Skellen', 'Corvo Bianco coroner', 'nilfgaard', 'hero', 'close', 4, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Stefan Skellen.jpg', 'Death serves the Empire.', 'hero', 'legendary', 0, 0),

-- Strong Units
('ng_imperial_guard', 'Imperial Guard', 'Elite protectors', 'nilfgaard', 'unit', 'close', 5, false, '["morale"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Impra Brigade Guard.jpg', 'Death before dishonor.', 'close', 'rare', 100, 20),
('ng_black_infantry', 'Black Infantry', 'Cursed soldiers', 'nilfgaard', 'unit', 'close', 4, false, '["tight_bond"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Black Infantry Archer.jpg', 'Shadows march.', 'close', 'rare', 80, 15),
('ng_imperial_pikeman', 'Imperial Pikeman', 'Disciplined formation', 'nilfgaard', 'unit', 'close', 3, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Redanian Foot soldier.jpg', 'The Empire endures.', 'close', 'common', 50, 10),
('ng_assassin', 'Assassin', 'Silent killer', 'nilfgaard', 'unit', 'close', 4, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Sheldon Skaggs.jpg', 'One strike.', 'close', 'common', 60, 10),
('ng_emissary', 'Emissary', 'Diplomat spy', 'nilfgaard', 'unit', 'ranged', 1, false, '["spy"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Yang Emissary.jpg', 'Words as weapons.', 'ranged', 'common', 40, 10),
('ng_spy', 'Imperial Spy', 'Eyes everywhere', 'nilfgaard', 'unit', 'close', 1, false, '["spy"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Sieg Technician.jpg', 'Know thy enemy.', 'close', 'common', 30, 5),
('ng_medic', 'Medic', 'Battlefield healer', 'nilfgaard', 'unit', 'close', 3, false, '["medic"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Ves.jpg', 'Service to Empire.', 'close', 'common', 50, 10),

-- Special Cards
('ng_commanders_horn', 'Commander''s Horn', 'Boost all units on a row by 3', 'nilfgaard', 'special', NULL, 0, false, '["commanders_horn"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'By decree.', 'special', 'rare', 0, 0),
('ng_clear_weather', 'Clear Weather', 'Remove all weather effects', 'nilfgaard', 'special', NULL, 0, false, '["clear_weather"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Sun returns.', 'special', 'common', 0, 0),
('ng_scorch', 'Scorch', 'Destroy highest power non-hero unit(s)', 'nilfgaard', 'special', NULL, 0, false, '["scorch"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Burn.', 'special', 'rare', 0, 0),
('ng_decoy', 'Decoy', 'Return a unit from your board to hand', 'nilfgaard', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Trickery.', 'special', 'common', 0, 0),
('ng_mardroeme', 'Mardroeme', 'Transform a unit', 'nilfgaard', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Cursed.', 'special', 'rare', 0, 0);

-- SCOIA'TAEL (16 cards)
INSERT INTO cards (id, name, description, faction, type, row, power, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap) VALUES
-- Heroes
('st_francesca', 'Francesca Findabair', 'Daisy of the Valley', 'scoiatael', 'hero', 'close', 4, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Filavandrel aen.jpg', 'Daisy of the Valley.', 'hero', 'legendary', 0, 0),
('st_eithne', 'Eithne', 'Queen of Dryads', 'scoiatael', 'hero', 'close', 1, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Ida Emean aep Sivney.jpg', 'Queen of Dryads.', 'hero', 'legendary', 0, 0),
('st_isfil', 'Isengrim Faoiltiarna', 'Commander of Scoia''tael', 'scoiatael', 'hero', 'close', 5, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Iorweth.jpg', 'Commander of Scoia''tael.', 'hero', 'legendary', 0, 0),
('st_toruviel', 'Toruviel', 'Elven commando', 'scoiatael', 'hero', 'close', 4, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Toruviel.jpg', 'Freedom or death.', 'hero', 'legendary', 0, 0),
('st_yaevinn', 'Yaevinn', 'Commando leader', 'scoiatael', 'unit', 'close', 5, false, '["morale"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Yaevinn.jpg', 'We watch.', 'close', 'epic', 100, 20),

-- Units
('st_commando', 'Commando', 'Guerrilla fighter', 'scoiatael', 'unit', 'close', 4, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Commando.jpg', 'Strike and fade.', 'close', 'rare', 80, 15),
('st_dragon_hunter', 'Dragon Hunter', 'Beast slayer', 'scoiatael', 'unit', 'close', 5, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Dragon Hunter.jpg', 'For the glens.', 'close', 'rare', 100, 20),
('st_elven_archer', 'Elven Archer', 'Deadly aim', 'scoiatael', 'unit', 'ranged', 4, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Elven Skirmisher.jpg', 'We wait.', 'ranged', 'common', 60, 10),
('st_elven_scout', 'Elven Scout', 'Swift archer', 'scoiatael', 'unit', 'ranged', 2, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Elven Skirmisher.jpg', 'We watch.', 'ranged', 'common', 40, 10),
('st_trapper', 'Trapper', 'Forest survivor', 'scoiatael', 'unit', 'ranged', 3, false, '["spy"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Contrebandier havekar.jpg', 'Nature provides.', 'ranged', 'common', 50, 10),
('st_vrihedd_brigade', 'Vrihedd Brigade', 'Elite elven archers', 'scoiatael', 'unit', 'ranged', 4, false, '["tight_bond"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Vrihedd Brigade.jpg', 'Precision.', 'ranged', 'rare', 80, 15),
('st_dol_blathanna', 'Dol Blathanna Archer', 'Aen Seidhe marksman', 'scoiatael', 'unit', 'ranged', 3, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Dol Blathanna.jpg', 'From the glens.', 'ranged', 'common', 50, 10),
('st_medic', 'Medic', 'Healer', 'scoiatael', 'unit', 'close', 3, false, '["medic"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Vesemir.jpg', 'Life finds a way.', 'close', 'common', 50, 10),

-- Special Cards
('st_commanders_horn', 'Commander''s Horn', 'Boost all units on a row by 3', 'scoiatael', 'special', NULL, 0, false, '["commanders_horn"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'For the glens!', 'special', 'rare', 0, 0),
('st_clear_weather', 'Clear Weather', 'Remove all weather effects', 'scoiatael', 'special', NULL, 0, false, '["clear_weather"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Sun returns.', 'special', 'common', 0, 0),
('st_scorch', 'Scorch', 'Destroy highest power non-hero unit(s)', 'scoiatael', 'special', NULL, 0, false, '["scorch"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Burn.', 'special', 'rare', 0, 0),
('st_decoy', 'Decoy', 'Return a unit from your board to hand', 'scoiatael', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Trickery.', 'special', 'common', 0, 0),
('st_mardroeme', 'Mardroeme', 'Transform a unit', 'scoiatael', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Cursed.', 'special', 'rare', 0, 0);

-- MONSTERS (17 cards)
INSERT INTO cards (id, name, description, faction, type, row, power, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap) VALUES
-- Heroes
('mo_eredin', 'Eredin Bréacc Glas', 'King of the Wild Hunt', 'monsters', 'hero', 'close', 1, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Hemdall.jpg', 'King of the Wild Hunt.', 'hero', 'legendary', 0, 0),
('mo_glustyworp', 'Glustyworp', 'Nest keeper', 'monsters', 'hero', 'close', 1, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Grave Hag.jpg', 'Nest keeper.', 'hero', 'legendary', 0, 0),
('mo_crone', 'Crone', 'Ancient evil', 'monsters', 'unit', 'ranged', 5, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Celaeno Harpy.jpg', 'Weave fate.', 'ranged', 'epic', 100, 20),

-- Units with Muster
('mo_nekker', 'Nekker', 'Forest creature', 'monsters', 'unit', 'close', 2, false, '["muster"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Nekker.jpg', 'Never alone.', 'close', 'common', 40, 10),
('mo_endrega', 'Endrega', 'Insectoid horror', 'monsters', 'unit', 'close', 3, false, '["muster"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Forktail.jpg', 'Swarm.', 'close', 'common', 50, 10),
('mo_foglet', 'Foglet', 'Mist wraith', 'monsters', 'unit', 'ranged', 3, false, '["muster"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Foglet.jpg', 'In the mist.', 'ranged', 'common', 50, 10),
('mo_arachas', 'Arachas', 'Giant scorpion', 'monsters', 'unit', 'close', 5, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Ice Giant.jpg', 'Chitin and venom.', 'close', 'rare', 100, 20),
('mo_ekimmara', 'Ekimmara', 'Vampire spawn', 'monsters', 'unit', 'close', 4, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/VampireEkimmara.jpg', 'Blood hunger.', 'close', 'common', 60, 10),
('mo_vampire_bruxa', 'Vampire Bruxa', 'Higher vampire', 'monsters', 'unit', 'close', 6, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Vampire Bruxa.jpg', 'Night hunter.', 'close', 'epic', 150, 30),
('mo_grave_hag', 'Grave Hag', 'Carrion feeder', 'monsters', 'unit', 'ranged', 4, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Grave Hag.jpg', 'Death''s harvest.', 'ranged', 'rare', 80, 15),
('mo_celaeno_harpy', 'Celaeno Harpy', 'Mountain predator', 'monsters', 'unit', 'ranged', 3, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Celaeno Harpy.jpg', 'Screech and strike.', 'ranged', 'common', 50, 10),

-- Special Cards
('mo_commanders_horn', 'Commander''s Horn', 'Boost all units on a row by 3', 'monsters', 'special', NULL, 0, false, '["commanders_horn"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Ride forth!', 'special', 'rare', 0, 0),
('mo_clear_weather', 'Clear Weather', 'Remove all weather effects', 'monsters', 'special', NULL, 0, false, '["clear_weather"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Sun returns.', 'special', 'common', 0, 0),
('mo_scorch', 'Scorch', 'Destroy highest power non-hero unit(s)', 'monsters', 'special', NULL, 0, false, '["scorch"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Burn.', 'special', 'rare', 0, 0),
('mo_decoy', 'Decoy', 'Return a unit from your board to hand', 'monsters', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Trickery.', 'special', 'common', 0, 0),
('mo_mardroeme', 'Mardroeme', 'Transform a unit', 'monsters', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Cursed.', 'special', 'rare', 0, 0);

-- NEUTRAL (14 cards)
INSERT INTO cards (id, name, description, faction, type, row, power, is_hero, abilities, art_url, flavor_text, card_group, rarity, cost_gold, cost_scrap) VALUES
-- Heroes
('neut_geralt', 'Geralt of Rivia', 'White Wolf', 'neutral', 'hero', 'close', 15, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Sieg fried of Denesle.jpg', 'Evil is evil.', 'hero', 'legendary', 0, 0),
('neut_ciri', 'Cirilla', 'Child of Destiny', 'neutral', 'hero', 'close', 15, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Yeneffer.jpg', 'Time and space.', 'hero', 'legendary', 0, 0),
('neut_yennefer', 'Yennefer', 'Sorceress', 'neutral', 'hero', 'close', 7, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/philippa Eilhart.jpg', 'Power and beauty.', 'hero', 'legendary', 0, 0),
('neut_triss', 'Triss Merigold', 'Sorceress', 'neutral', 'hero', 'close', 7, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Sabrina Glevissig.jpg', 'Fire and faith.', 'hero', 'legendary', 0, 0),
('neut_zoltan', 'Zoltan Chivay', 'Dwarf warrior', 'neutral', 'hero', 'close', 6, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Yarpen Zigrin.jpg', 'Honor among thieves.', 'hero', 'legendary', 0, 0),
('neut_dandelion', 'Dandelion', 'Bard', 'neutral', 'hero', 'close', 2, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Schirru.jpg', 'Silver tongue.', 'hero', 'legendary', 0, 0),
('neut_vesemir', 'Vesemir', 'Old Witcher', 'neutral', 'hero', 'close', 6, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Vesemir.jpg', 'Experience.', 'hero', 'legendary', 0, 0),
('neut_keira', 'Keira Metz', 'Sorceress', 'neutral', 'hero', 'close', 5, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Keira Metz.jpg', 'Secrets.', 'hero', 'legendary', 0, 0),
('neut_thaler', 'Thaler', 'Spy', 'neutral', 'unit', 'close', 1, false, '["spy"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Sieg Technician.jpg', 'Information.', 'close', 'common', 30, 5),
('neut_yarpen', 'Yarpen Zigrin', 'Dwarf commander', 'neutral', 'hero', 'close', 4, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Yarpen Zigrin.jpg', 'Leader.', 'hero', 'legendary', 0, 0),
('neut_shani', 'Shani', 'Medic', 'neutral', 'unit', 'close', 3, false, '["medic"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Ves.jpg', 'Healing hands.', 'close', 'common', 50, 10),
('neut_gaunter', 'Gaunter O''Dimm', 'Master Mirror', 'neutral', 'hero', 'close', 1, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Gaunter O Dimm.jpg', 'Time is a flat circle.', 'hero', 'legendary', 0, 0),
('neut_kambi', 'Kambi', 'Reckless warrior', 'neutral', 'hero', 'close', 12, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Kambi.jpg', 'No fear.', 'hero', 'legendary', 0, 0),
('neut_poor_infantry', 'Poor Fucking Infantry', 'Basic foot soldier', 'neutral', 'unit', 'close', 1, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/poor fucking infantry.jpg', 'Every army needs fodder.', 'close', 'common', 10, 2);

-- ============================================================================
-- SEED SHOP ITEMS (cards will exist now - no FK constraint)
-- ============================================================================
INSERT INTO shop_items (id, name, description, card_id, cost_gold, cost_scrap, stock, level_required) VALUES
('shop_card_nr_001', 'Vernon Roche', 'Temerian commander', 'nr_vernon_roche', 100, 20, -1, 1),
('shop_card_nr_002', 'Blue Stripes Commando', 'Elite Temerian soldiers', 'nr_blue_stripes_commando', 80, 15, -1, 1),
('shop_card_nr_003', 'Reinforced Ballista', 'Siege engine', 'nr_ballista', 120, 25, -1, 2),
('shop_card_ng_001', 'Imperial Golem', 'Nilfgaardian construct', 'ng_imperial_guard', 100, 20, -1, 1),
('shop_card_ng_002', 'Viper Witcher', 'Empire''s monster slayer', 'ng_assassin', 200, 50, -1, 3),
('shop_card_st_001', 'Elven Scout', 'Scoia''tael ranger', 'st_elven_scout', 80, 15, -1, 1),
('shop_card_mo_001', 'Ghoul', 'Scavenger of the dead', 'mo_nekker', 80, 15, -1, 1),
('shop_card_sy_001', 'Bounty Hunter', 'Coin for blood', 'neut_thaler', 80, 15, -1, 1),
('shop_chest_starter', 'Starter Chest', 'Basic cards for new players', NULL, 500, 0, -1, 1),
('shop_chest_novice', 'Novice Chest', 'Cards for level 1+ players', NULL, 1000, 0, -1, 1),
('shop_chest_adept', 'Adept Chest', 'Better cards for level 2+ players', NULL, 2000, 0, -1, 2),
('shop_axe', 'Mining Axe', 'Tool for extracting gold', NULL, 100, 0, -1, 1)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SEED CHESTS
-- ============================================================================
INSERT INTO chests (id, name, description, level_required, cards, gold, cooldown) VALUES
('starter', 'Starter Chest', 'Basic cards for new players', 1,
 '[{"card_id": "nr_kaedweni_soldier", "min_qty": 1, "max_qty": 2, "weight": 50}, {"card_id": "ng_imperial_pikeman", "min_qty": 1, "max_qty": 2, "weight": 50}]'::jsonb,
 50, '24 hours'),
('level_1', 'Novice Chest', 'Cards for level 1+ players', 1,
 '[{"card_id": "nr_blue_stripes_commando", "min_qty": 1, "max_qty": 1, "weight": 40}, {"card_id": "st_elven_scout", "min_qty": 1, "max_qty": 1, "weight": 30}, {"card_id": "mo_nekker", "min_qty": 1, "max_qty": 1, "weight": 30}]'::jsonb,
 100, '12 hours'),
('level_2', 'Adept Chest', 'Better cards for level 2+ players', 2,
 '[{"card_id": "nr_dun_banner", "min_qty": 1, "max_qty": 1, "weight": 30}, {"card_id": "st_vrihedd_brigade", "min_qty": 1, "max_qty": 1, "weight": 35}, {"card_id": "ng_black_infantry", "min_qty": 1, "max_qty": 1, "weight": 35}]'::jsonb,
 200, '8 hours')
ON CONFLICT (id) DO NOTHING;