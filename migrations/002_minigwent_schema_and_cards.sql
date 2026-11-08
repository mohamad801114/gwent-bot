-- Migration 002: Transform to Witcher 3 Mini-Gwent + Seed Cards
-- Run AFTER 001_initial_schema.sql

-- ============================================================================
-- STEP 1: SIMPLIFY SCHEMA FOR MINI-GWENT
-- ============================================================================

-- Remove armor from cards table
ALTER TABLE cards DROP COLUMN IF EXISTS armor;

-- Remove provisions from cards table (not used in mini-gwent)
ALTER TABLE cards DROP COLUMN IF EXISTS provision;

-- Add mini-gwent specific columns
ALTER TABLE cards ADD COLUMN IF NOT EXISTS is_hero BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE cards ADD COLUMN IF NOT EXISTS faction VARCHAR(50);        -- northern_realms, nilfgaard, scoiatael, monsters, neutral
ALTER TABLE cards ADD COLUMN IF NOT EXISTS card_group VARCHAR(50);     -- close, ranged, siege, special, hero

-- Simplify players table for mini-gwent progression
ALTER TABLE players DROP COLUMN IF EXISTS scrap;
ALTER TABLE players DROP COLUMN IF EXISTS meteorite_powder;
ALTER TABLE players DROP COLUMN IF EXISTS mine_unlocked;
ALTER TABLE players DROP COLUMN IF EXISTS mine_extracts_today;
ALTER TABLE players DROP COLUMN IF EXISTS last_mine_extract_at;
ALTER TABLE players DROP COLUMN IF EXISTS axe_durability;
ALTER TABLE players DROP COLUMN IF EXISTS axe_claimed_at;
ALTER TABLE players DROP COLUMN IF EXISTS daily_coin_claimed_at;
ALTER TABLE players DROP COLUMN IF EXISTS leader_id;

-- Add mini-gwent progression columns
ALTER TABLE players ADD COLUMN IF NOT EXISTS leader_ability VARCHAR(100);  -- passive leader ability
ALTER TABLE players ADD COLUMN IF NOT EXISTS cards_won INT NOT NULL DEFAULT 0;
ALTER TABLE players ADD COLUMN IF NOT EXISTS opponents_defeated JSONB NOT NULL DEFAULT '[]'::jsonb;

-- Update existing battle data: remove armor from all card instances in JSONB
UPDATE battles SET
    player1_hand = (SELECT jsonb_agg(jsonb_strip_nulls(elem - 'armor')) FROM jsonb_array_elements(player1_hand) elem),
    player2_hand = (SELECT jsonb_agg(jsonb_strip_nulls(elem - 'armor')) FROM jsonb_array_elements(player2_hand) elem),
    player1_deck = (SELECT jsonb_agg(jsonb_strip_nulls(elem - 'armor')) FROM jsonb_array_elements(player1_deck) elem),
    player2_deck = (SELECT jsonb_agg(jsonb_strip_nulls(elem - 'armor')) FROM jsonb_array_elements(player2_deck) elem),
    player1_graveyard = (SELECT jsonb_agg(jsonb_strip_nulls(elem - 'armor')) FROM jsonb_array_elements(player1_graveyard) elem),
    player2_graveyard = (SELECT jsonb_agg(jsonb_strip_nulls(elem - 'armor')) FROM jsonb_array_elements(player2_graveyard) elem),
    player1_board = jsonb_set(player1_board, '{melee}', (SELECT jsonb_agg(jsonb_strip_nulls(elem - 'armor')) FROM jsonb_array_elements(player1_board->'melee') elem)),
    player1_board = jsonb_set(player1_board, '{ranged}', (SELECT jsonb_agg(jsonb_strip_nulls(elem - 'armor')) FROM jsonb_array_elements(player1_board->'ranged') elem)),
    player1_board = jsonb_set(player1_board, '{siege}', (SELECT jsonb_agg(jsonb_strip_nulls(elem - 'armor')) FROM jsonb_array_elements(player1_board->'siege') elem)),
    player2_board = jsonb_set(player2_board, '{melee}', (SELECT jsonb_agg(jsonb_strip_nulls(elem - 'armor')) FROM jsonb_array_elements(player2_board->'melee') elem)),
    player2_board = jsonb_set(player2_board, '{ranged}', (SELECT jsonb_agg(jsonb_strip_nulls(elem - 'armor')) FROM jsonb_array_elements(player2_board->'ranged') elem)),
    player2_board = jsonb_set(player2_board, '{siege}', (SELECT jsonb_agg(jsonb_strip_nulls(elem - 'armor')) FROM jsonb_array_elements(player2_board->'siege') elem));

-- Keywords reference table
CREATE TABLE IF NOT EXISTS minigwent_keywords (
    keyword VARCHAR(30) PRIMARY KEY,
    description TEXT
);

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
('clear_weather', 'Removes all weather effects')
ON CONFLICT (keyword) DO NOTHING;

-- ============================================================================
-- STEP 2: SEED MINI-GWENT CARDS (Witcher 3 Base Game)
-- ============================================================================

DELETE FROM cards;

-- NORTHERN REALMS
INSERT INTO cards (id, name, description, faction, type, row, power, is_hero, abilities, art_url, flavor_text, card_group) VALUES
('nr_poor_infantry', 'Poor Infantry', 'Basic foot soldier', 'northern_realms', 'unit', 'close', 1, false, '[]', 'https://cdn.example.com/cards/nr_poor_infantry.webp', 'Every army needs fodder.', 'close'),
('nr_kaedweni_soldier', 'Kaedweni Soldier', 'Sturdy spearman', 'northern_realms', 'unit', 'close', 3, false, '[]', 'https://cdn.example.com/cards/nr_kaedweni_soldier.webp', 'Steel and resolve.', 'close'),
('nr_blue_stripes_commando', 'Blue Stripes Commando', 'Elite Temerian unit', 'northern_realms', 'unit', 'close', 4, false, '["tight_bond"]', 'https://cdn.example.com/cards/nr_blue_stripes_commando.webp', 'Discipline wins wars.', 'close'),
('nr_dun_banner', 'Dun Banner Heavy Cavalry', 'Armored charge', 'northern_realms', 'unit', 'close', 5, false, '["morale"]', 'https://cdn.example.com/cards/nr_dun_banner.webp', 'Charge!', 'close'),
('nr_crinfrid_reaver', 'Crinfrid Reaver', 'Dragon hunters', 'northern_realms', 'unit', 'close', 4, false, '["tight_bond"]', 'https://cdn.example.com/cards/nr_crinfrid_reaver.webp', 'Gold for scales.', 'close'),
('nr_ballista', 'Ballista', 'Siege engine', 'northern_realms', 'unit', 'siege', 6, false, '[]', 'https://cdn.example.com/cards/nr_ballista.webp', 'Rain fire.', 'siege'),
('nr_catapult', 'Catapult', 'Heavy artillery', 'northern_realms', 'unit', 'siege', 5, false, '[]', 'https://cdn.example.com/cards/nr_catapult.webp', 'Smash walls.', 'siege'),
('nr_trebuchet', 'Trebuchet', 'Long-range siege', 'northern_realms', 'unit', 'siege', 8, false, '[]', 'https://cdn.example.com/cards/nr_trebuchet.webp', 'From afar.', 'siege'),
('nr_archer', 'Archer', 'Longbow expert', 'northern_realms', 'unit', 'ranged', 3, false, '[]', 'https://cdn.example.com/cards/nr_archer.webp', 'Aim true.', 'ranged'),
('nr_arbalist', 'Arbalist', 'Crossbow specialist', 'northern_realms', 'unit', 'ranged', 4, false, '[]', 'https://cdn.example.com/cards/nr_arbalist.webp', 'Pierce armor.', 'ranged'),
('nr_redanian_elite', 'Redanian Elite', 'Armored crossbowman', 'northern_realms', 'unit', 'ranged', 5, false, '["morale"]', 'https://cdn.example.com/cards/nr_redanian_elite.webp', 'For Radovid.', 'ranged'),
('nr_blue_stripes_scout', 'Blue Stripes Scout', 'Recon unit', 'northern_realms', 'unit', 'ranged', 2, false, '["spy"]', 'https://cdn.example.com/cards/nr_blue_stripes_scout.webp', 'Eyes everywhere.', 'ranged'),
('nr_medic', 'Medic', 'Battlefield healer', 'northern_realms', 'unit', 'close', 3, false, '["medic"]', 'https://cdn.example.com/cards/nr_medic.webp', 'Life preserved.', 'close'),
('nr_foltest_siegemaster', 'Foltest, The Siegemaster', 'Leader: +1 siege', 'northern_realms', 'hero', 'close', 6, true, '["leader"]', 'https://cdn.example.com/cards/nr_foltest_siegemaster.webp', 'The North remembers.', 'hero'),
('nr_foltest_commander', 'Foltest, King of Temeria', 'Leader: +1 close', 'northern_realms', 'hero', 'close', 5, true, '["leader"]', 'https://cdn.example.com/cards/nr_foltest_commander.webp', 'Temeria stands.', 'hero'),
('nr_foltest_royal', 'Foltest, The Royal', 'Leader: +1 ranged', 'northern_realms', 'hero', 'close', 4, true, '["leader"]', 'https://cdn.example.com/cards/nr_foltest_royal.webp', 'By blood and steel.', 'hero'),
('nr_commanders_horn', 'Commander''s Horn', 'Boost row by 3', 'northern_realms', 'special', NULL, 0, false, '["commanders_horn"]', 'https://cdn.example.com/cards/nr_commanders_horn.webp', 'For the North!', 'special'),
('nr_mardroeme', 'Mardroeme', 'Transform unit to 1 power', 'northern_realms', 'special', NULL, 0, false, '[]', 'https://cdn.example.com/cards/nr_mardroeme.webp', 'Cursed flesh.', 'special'),
('nr_clear_weather', 'Clear Weather', 'Remove weather', 'northern_realms', 'special', NULL, 0, false, '["clear_weather"]', 'https://cdn.example.com/cards/nr_clear_weather.webp', 'Sun returns.', 'special'),
('nr_scorch', 'Scorch', 'Destroy highest unit', 'northern_realms', 'special', NULL, 0, false, '["scorch"]', 'https://cdn.example.com/cards/nr_scorch.webp', 'Burn them all.', 'special'),
('nr_decoy', 'Decoy', 'Return unit to hand', 'northern_realms', 'special', NULL, 0, false, '[]', 'https://cdn.example.com/cards/nr_decoy.webp', 'Trickery.', 'special'),

-- NILFGAARD
('ng_imperial_pikeman', 'Imperial Pikeman', 'Disciplined formation', 'nilfgaard', 'unit', 'close', 3, false, '[]', 'https://cdn.example.com/cards/ng_imperial_pikeman.webp', 'The Empire endures.', 'close'),
('ng_imperial_guard', 'Imperial Guard', 'Elite protectors', 'nilfgaard', 'unit', 'close', 5, false, '["morale"]', 'https://cdn.example.com/cards/ng_imperial_guard.webp', 'Death before dishonor.', 'close'),
('ng_black_infantry', 'Black Infantry', 'Cursed soldiers', 'nilfgaard', 'unit', 'close', 4, false, '["tight_bond"]', 'https://cdn.example.com/cards/ng_black_infantry.webp', 'Shadows march.', 'close'),
('ng_assassin', 'Assassin', 'Silent killer', 'nilfgaard', 'unit', 'close', 4, false, '[]', 'https://cdn.example.com/cards/ng_assassin.webp', 'One strike.', 'close'),
('ng_emissary', 'Emissary', 'Diplomat spy', 'nilfgaard', 'unit', 'ranged', 1, false, '["spy"]', 'https://cdn.example.com/cards/ng_emissary.webp', 'Words as weapons.', 'ranged'),
('ng_spy', 'Imperial Spy', 'Eyes everywhere', 'nilfgaard', 'unit', 'close', 1, false, '["spy"]', 'https://cdn.example.com/cards/ng_spy.webp', 'Know thy enemy.', 'close'),
('ng_medic', 'Medic', 'Battlefield healer', 'nilfgaard', 'unit', 'close', 3, false, '["medic"]', 'https://cdn.example.com/cards/ng_medic.webp', 'Service to Empire.', 'close'),
('ng_emhyr', 'Emhyr var Emreis', 'Emperor: draw 1', 'nilfgaard', 'hero', 'close', 1, true, '["leader"]', 'https://cdn.example.com/cards/ng_emhyr.webp', 'Empire comes first.', 'hero'),
('ng_menno', 'Menno Coehoorn', 'Siege master', 'nilfgaard', 'hero', 'close', 10, true, '["leader"]', 'https://cdn.example.com/cards/ng_menno.webp', 'Engineer of war.', 'hero'),
('ng_vattier', 'Vattier de Rideaux', 'Spymaster', 'nilfgaard', 'hero', 'close', 1, true, '["leader"]', 'https://cdn.example.com/cards/ng_vattier.webp', 'Information is power.', 'hero'),
('ng_commanders_horn', 'Commander''s Horn', 'Boost row by 3', 'nilfgaard', 'special', NULL, 0, false, '["commanders_horn"]', 'https://cdn.example.com/cards/ng_commanders_horn.webp', 'By decree.', 'special'),
('ng_mardroeme', 'Mardroeme', 'Transform unit', 'nilfgaard', 'special', NULL, 0, false, '[]', 'https://cdn.example.com/cards/ng_mardroeme.webp', 'Cursed.', 'special'),
('ng_clear_weather', 'Clear Weather', 'Remove weather', 'nilfgaard', 'special', NULL, 0, false, '["clear_weather"]', 'https://cdn.example.com/cards/ng_clear_weather.webp', 'Sun returns.', 'special'),
('ng_scorch', 'Scorch', 'Destroy highest', 'nilfgaard', 'special', NULL, 0, false, '["scorch"]', 'https://cdn.example.com/cards/ng_scorch.webp', 'Burn.', 'special'),
('ng_decoy', 'Decoy', 'Return to hand', 'nilfgaard', 'special', NULL, 0, false, '[]', 'https://cdn.example.com/cards/ng_decoy.webp', 'Trickery.', 'special'),

-- SCOIA'TAEL
('st_elven_scout', 'Elven Scout', 'Swift archer', 'scoiatael', 'unit', 'ranged', 2, false, '[]', 'https://cdn.example.com/cards/st_elven_scout.webp', 'We watch.', 'ranged'),
('st_elven_archer', 'Elven Archer', 'Deadly aim', 'scoiatael', 'unit', 'ranged', 4, false, '[]', 'https://cdn.example.com/cards/st_elven_archer.webp', 'We wait.', 'ranged'),
('st_yaevinn', 'Yaevinn', 'Commando leader', 'scoiatael', 'unit', 'close', 5, false, '["morale"]', 'https://cdn.example.com/cards/st_yaevinn.webp', 'Freedom or death.', 'close'),
('st_commando', 'Commando', 'Guerrilla fighter', 'scoiatael', 'unit', 'close', 4, false, '[]', 'https://cdn.example.com/cards/st_commando.webp', 'Strike and fade.', 'close'),
('st_dragon_hunter', 'Dragon Hunter', 'Beast slayer', 'scoiatael', 'unit', 'close', 5, false, '[]', 'https://cdn.example.com/cards/st_dragon_hunter.webp', 'For the glens.', 'close'),
('st_trapper', 'Trapper', 'Forest survivor', 'scoiatael', 'unit', 'ranged', 3, false, '["spy"]', 'https://cdn.example.com/cards/st_trapper.webp', 'Nature provides.', 'ranged'),
('st_medic', 'Medic', 'Healer', 'scoiatael', 'unit', 'close', 3, false, '["medic"]', 'https://cdn.example.com/cards/st_medic.webp', 'Life finds a way.', 'close'),
('st_francesca', 'Francesca Findabair', 'Leader: ranged+1', 'scoiatael', 'hero', 'close', 4, true, '["leader"]', 'https://cdn.example.com/cards/st_francesca.webp', 'Daisy of the Valley.', 'hero'),
('st_eithne', 'Eithne', 'Leader: draw 1', 'scoiatael', 'hero', 'close', 1, true, '["leader"]', 'https://cdn.example.com/cards/st_eithne.webp', 'Queen of Dryads.', 'hero'),
('st_isfil', 'Isengrim Faoiltiarna', 'Leader: close+1', 'scoiatael', 'hero', 'close', 5, true, '["leader"]', 'https://cdn.example.com/cards/st_isfil.webp', 'Commander of Scoia''tael.', 'hero'),
('st_commanders_horn', 'Commander''s Horn', 'Boost row by 3', 'scoiatael', 'special', NULL, 0, false, '["commanders_horn"]', 'https://cdn.example.com/cards/st_commanders_horn.webp', 'For the glens!', 'special'),
('st_mardroeme', 'Mardroeme', 'Transform unit', 'scoiatael', 'special', NULL, 0, false, '[]', 'https://cdn.example.com/cards/st_mardroeme.webp', 'Cursed.', 'special'),
('st_clear_weather', 'Clear Weather', 'Remove weather', 'scoiatael', 'special', NULL, 0, false, '["clear_weather"]', 'https://cdn.example.com/cards/st_clear_weather.webp', 'Sun returns.', 'special'),
('st_scorch', 'Scorch', 'Destroy highest', 'scoiatael', 'special', NULL, 0, false, '["scorch"]', 'https://cdn.example.com/cards/st_scorch.webp', 'Burn.', 'special'),
('st_decoy', 'Decoy', 'Return to hand', 'scoiatael', 'special', NULL, 0, false, '[]', 'https://cdn.example.com/cards/st_decoy.webp', 'Trickery.', 'special'),

-- MONSTERS
('mo_nekker', 'Nekker', 'Forest creature', 'monsters', 'unit', 'close', 2, false, '["muster"]', 'https://cdn.example.com/cards/mo_nekker.webp', 'Never alone.', 'close'),
('mo_endrega', 'Endrega', 'Insectoid horror', 'monsters', 'unit', 'close', 3, false, '["muster"]', 'https://cdn.example.com/cards/mo_endrega.webp', 'Swarm.', 'close'),
('mo_foglet', 'Foglet', 'Mist wraith', 'monsters', 'unit', 'ranged', 3, false, '["muster"]', 'https://cdn.example.com/cards/mo_foglet.webp', 'In the mist.', 'ranged'),
('mo_arachas', 'Arachas', 'Giant scorpion', 'monsters', 'unit', 'close', 5, false, '[]', 'https://cdn.example.com/cards/mo_arachas.webp', 'Chitin and venom.', 'close'),
('mo_ekimmara', 'Ekimmara', 'Vampire spawn', 'monsters', 'unit', 'close', 4, false, '[]', 'https://cdn.example.com/cards/mo_ekimmara.webp', 'Blood hunger.', 'close'),
('mo_grave_hag', 'Grave Hag', 'Carrion feeder', 'monsters', 'unit', 'ranged', 4, false, '[]', 'https://cdn.example.com/cards/mo_grave_hag.webp', 'Death''s harvest.', 'ranged'),
('mo_crone', 'Crone', 'Ancient evil', 'monsters', 'unit', 'ranged', 5, false, '[]', 'https://cdn.example.com/cards/mo_crone.webp', 'Weave fate.', 'ranged'),
('mo_eredin', 'Eredin', 'Leader: frost', 'monsters', 'hero', 'close', 1, true, '["leader"]', 'https://cdn.example.com/cards/mo_eredin.webp', 'King of the Wild Hunt.', 'hero'),
('mo_glustyworp', 'Glustyworp', 'Leader: muster', 'monsters', 'hero', 'close', 1, true, '["leader"]', 'https://cdn.example.com/cards/mo_glustyworp.webp', 'Nest keeper.', 'hero'),
('mo_commanders_horn', 'Commander''s Horn', 'Boost row by 3', 'monsters', 'special', NULL, 0, false, '["commanders_horn"]', 'https://cdn.example.com/cards/mo_commanders_horn.webp', 'Ride forth!', 'special'),
('mo_mardroeme', 'Mardroeme', 'Transform unit', 'monsters', 'special', NULL, 0, false, '[]', 'https://cdn.example.com/cards/mo_mardroeme.webp', 'Cursed.', 'special'),
('mo_clear_weather', 'Clear Weather', 'Remove weather', 'monsters', 'special', NULL, 0, false, '["clear_weather"]', 'https://cdn.example.com/cards/mo_clear_weather.webp', 'Sun returns.', 'special'),
('mo_scorch', 'Scorch', 'Destroy highest', 'monsters', 'special', NULL, 0, false, '["scorch"]', 'https://cdn.example.com/cards/mo_scorch.webp', 'Burn.', 'special'),
('mo_decoy', 'Decoy', 'Return to hand', 'monsters', 'special', NULL, 0, false, '[]', 'https://cdn.example.com/cards/mo_decoy.webp', 'Trickery.', 'special'),

-- NEUTRAL HEROES
('neut_geralt', 'Geralt of Rivia', 'White Wolf', 'neutral', 'hero', 'close', 15, true, '[]', 'https://cdn.example.com/cards/neut_geralt.webp', 'Evil is evil.', 'hero'),
('neut_ciri', 'Cirilla', 'Child of Destiny', 'neutral', 'hero', 'close', 15, true, '[]', 'https://cdn.example.com/cards/neut_ciri.webp', 'Time and space.', 'hero'),
('neut_yennefer', 'Yennefer', 'Sorceress', 'neutral', 'hero', 'close', 7, true, '[]', 'https://cdn.example.com/cards/neut_yennefer.webp', 'Power and beauty.', 'hero'),
('neut_triss', 'Triss Merigold', 'Sorceress', 'neutral', 'hero', 'close', 7, true, '[]', 'https://cdn.example.com/cards/neut_triss.webp', 'Fire and faith.', 'hero'),
('neut_zoltan', 'Zoltan Chivay', 'Dwarf warrior', 'neutral', 'hero', 'close', 6, true, '[]', 'https://cdn.example.com/cards/neut_zoltan.webp', 'Honor among thieves.', 'hero'),
('neut_dandelion', 'Dandelion', 'Bard', 'neutral', 'hero', 'close', 2, true, '[]', 'https://cdn.example.com/cards/neut_dandelion.webp', 'Silver tongue.', 'hero'),
('neut_vesemir', 'Vesemir', 'Old Witcher', 'neutral', 'hero', 'close', 6, true, '[]', 'https://cdn.example.com/cards/neut_vesemir.webp', 'Experience.', 'hero'),
('neut_keira', 'Keira Metz', 'Sorceress', 'neutral', 'hero', 'close', 5, true, '[]', 'https://cdn.example.com/cards/neut_keira.webp', 'Secrets.', 'hero'),
('neut_thaler', 'Thaler', 'Spy', 'neutral', 'unit', 'close', 1, false, '["spy"]', 'https://cdn.example.com/cards/neut_thaler.webp', 'Information.', 'close'),
('neut_yarpen', 'Yarpen Zigrin', 'Dwarf commander', 'neutral', 'hero', 'close', 4, true, '[]', 'https://cdn.example.com/cards/neut_yarpen.webp', 'Leader.', 'hero');