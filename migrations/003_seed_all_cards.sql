-- Migration 003: Seed ALL Mini-Gwent Cards from Assets (Witcher 3 Style)
-- Run AFTER 001 and 002
-- Based on images in assets/images/ with GitHub URLs

-- Clear existing cards
DELETE FROM cards;

-- ============================================================================
-- NORTHERN REALMS (20 cards)
-- ============================================================================

-- Heroes (Leader cards - high power, is_hero=true)
INSERT INTO cards (id, name, description, faction, type, row, power, is_hero, abilities, art_url, flavor_text, card_group) VALUES
('nr_foltest_siegemaster', 'Foltest, The Siegemaster', 'Leader: +1 power to all Siege units', 'northern_realms', 'hero', 'close', 6, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Esterad Thyssen.jpg', 'The North remembers.', 'hero'),
('nr_foltest_commander', 'Foltest, King of Temeria', 'Leader: +1 power to all Close units', 'northern_realms', 'hero', 'close', 5, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Barclay Els.jpg', 'Temeria stands.', 'hero'),
('nr_foltest_royal', 'Foltest, The Royal', 'Leader: +1 power to all Ranged units', 'northern_realms', 'hero', 'close', 4, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Prince Stennis.jpg', 'By blood and steel.', 'hero'),
('nr_vernon_roche', 'Vernon Roche', 'Commander of the Blue Stripes', 'northern_realms', 'hero', 'close', 6, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Vernon Roche.jpg', 'Temeria will not fall while I draw breath.', 'hero'),
('nr_siegfried', 'Siegfried of Denesle', 'Order of the Flaming Rose', 'northern_realms', 'hero', 'close', 5, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Sieg fried of Denesle.jpg', 'By fire and sword.', 'hero'),

-- Gold/Special Cards (high power units)
('nr_esterad_thyssen', 'Esterad Thyssen', 'King of Kovir and Poviss', 'northern_realms', 'unit', 'close', 7, false, '["morale"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Esterad Thyssen.jpg', 'Wealth is power.', 'close'),
('nr_hjalmar', 'Hjalmar', 'Son of Crach an Craite', 'northern_realms', 'unit', 'close', 6, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Hjalmar.jpg', 'Glory awaits.', 'close'),
('nr_ciri', 'Cirilla', 'Child of Destiny', 'northern_realms', 'hero', 'close', 15, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Yeneffer.jpg', 'Time and space.', 'hero'),
('nr_yennefer', 'Yennefer', 'Sorceress of Vengerberg', 'northern_realms', 'hero', 'close', 7, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/philippa Eilhart.jpg', 'Power and beauty.', 'hero'),
('nr_triss', 'Triss Merigold', 'Sorceress', 'northern_realms', 'hero', 'close', 7, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Sabrina Glevissig.jpg', 'Fire and faith.', 'hero'),

-- Silver/Strong Units
('nr_blue_stripes_commando', 'Blue Stripes Commando', 'Elite Temerian unit', 'northern_realms', 'unit', 'close', 4, false, '["tight_bond"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Blue Stripes Commando.jpg', 'Discipline wins wars.', 'close'),
('nr_crinfrid_reaver', 'Crinfrid Reaver', 'Dragon hunters', 'northern_realms', 'unit', 'close', 4, false, '["tight_bond"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Crinfrid Reavers.jpg', 'Gold for scales.', 'close'),
('nr_kaedweni_soldier', 'Kaedweni Soldier', 'Sturdy spearman', 'northern_realms', 'unit', 'close', 3, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Kaed weni Sige.jpg', 'Steel and resolve.', 'close'),
('nr_dun_banner', 'Dun Banner Heavy Cavalry', 'Armored charge', 'northern_realms', 'unit', 'close', 5, false, '["morale"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Impra Brigade Guard.jpg', 'Charge!', 'close'),
('nr_redanian_elite', 'Redanian Elite', 'Armored crossbowman', 'northern_realms', 'unit', 'ranged', 5, false, '["morale"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Redanian Foot soldier.jpg', 'For Radovid.', 'ranged'),
('nr_blue_stripes_scout', 'Blue Stripes Scout', 'Recon unit', 'northern_realms', 'unit', 'ranged', 2, false, '["spy"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Etolian Auxiliary Archers.jpg', 'Eyes everywhere.', 'ranged'),
('nr_arbalist', 'Arbalist', 'Crossbow specialist', 'northern_realms', 'unit', 'ranged', 4, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Black Infantry Archer.jpg', 'Pierce armor.', 'ranged'),
('nr_archer', 'Archer', 'Longbow expert', 'northern_realms', 'unit', 'ranged', 3, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Elven Skirmisher.jpg', 'Aim true.', 'ranged'),
('nr_ballista', 'Ballista', 'Siege engine', 'northern_realms', 'unit', 'siege', 6, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Balista.jpg', 'Rain fire.', 'siege'),
('nr_catapult', 'Catapult', 'Heavy artillery', 'northern_realms', 'unit', 'siege', 5, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Catapult.jpg', 'Smash walls.', 'siege'),
('nr_trebuchet', 'Trebuchet', 'Long-range siege', 'northern_realms', 'unit', 'siege', 8, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Trebuchet.jpg', 'From afar.', 'siege'),

-- Special Cards
('nr_commanders_horn', 'Commander''s Horn', 'Boost all units on a row by 3', 'northern_realms', 'special', NULL, 0, false, '["commanders_horn"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'For the North!', 'special'),
('nr_clear_weather', 'Clear Weather', 'Remove all weather effects', 'northern_realms', 'special', NULL, 0, false, '["clear_weather"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Sun returns.', 'special'),
('nr_scorch', 'Scorch', 'Destroy highest power non-hero unit(s)', 'northern_realms', 'special', NULL, 0, false, '["scorch"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Burn them all.', 'special'),
('nr_decoy', 'Decoy', 'Return a unit from your board to hand', 'northern_realms', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Trickery.', 'special'),
('nr_mardroeme', 'Mardroeme', 'Transform a unit into a base copy of another', 'northern_realms', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Cursed flesh.', 'special'),

-- ============================================================================
-- NILFGAARD (18 cards)
-- ============================================================================

-- Heroes
('ng_emhyr', 'Emhyr var Emreis', 'Emperor of Nilfgaard', 'nilfgaard', 'hero', 'close', 1, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Morvran voorhis.jpg', 'Empire comes first.', 'hero'),
('ng_menno', 'Menno Coehoorn', 'Siege master', 'nilfgaard', 'hero', 'close', 10, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Menno Coehoorn.jpg', 'Engineer of war.', 'hero'),
('ng_vattier', 'Vattier de Rideaux', 'Spymaster', 'nilfgaard', 'hero', 'close', 1, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Vattier de Rideaux.jpg', 'Information is power.', 'hero'),
('ng_cahir', 'Cahir Mawr Dyffryn aep Ceallach', 'The Black Knight', 'nilfgaard', 'hero', 'close', 5, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Cahir Mawr.jpg', 'Destiny guides my blade.', 'hero'),
('ng_stefan_skellen', 'Stefan Skellen', 'Corvo Bianco coroner', 'nilfgaard', 'hero', 'close', 4, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Stefan Skellen.jpg', 'Death serves the Empire.', 'hero'),

-- Strong Units
('ng_imperial_guard', 'Imperial Guard', 'Elite protectors', 'nilfgaard', 'unit', 'close', 5, false, '["morale"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Impra Brigade Guard.jpg', 'Death before dishonor.', 'close'),
('ng_black_infantry', 'Black Infantry', 'Cursed soldiers', 'nilfgaard', 'unit', 'close', 4, false, '["tight_bond"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Black Infantry Archer.jpg', 'Shadows march.', 'close'),
('ng_imperial_pikeman', 'Imperial Pikeman', 'Disciplined formation', 'nilfgaard', 'unit', 'close', 3, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Redanian Foot soldier.jpg', 'The Empire endures.', 'close'),
('ng_assassin', 'Assassin', 'Silent killer', 'nilfgaard', 'unit', 'close', 4, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Sheldon Skaggs.jpg', 'One strike.', 'close'),
('ng_emissary', 'Emissary', 'Diplomat spy', 'nilfgaard', 'unit', 'ranged', 1, false, '["spy"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Yang Emissary.jpg', 'Words as weapons.', 'ranged'),
('ng_spy', 'Imperial Spy', 'Eyes everywhere', 'nilfgaard', 'unit', 'close', 1, false, '["spy"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Sieg Technician.jpg', 'Know thy enemy.', 'close'),
('ng_medic', 'Medic', 'Battlefield healer', 'nilfgaard', 'unit', 'close', 3, false, '["medic"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Ves.jpg', 'Service to Empire.', 'close'),

-- Special Cards
('ng_commanders_horn', 'Commander''s Horn', 'Boost all units on a row by 3', 'nilfgaard', 'special', NULL, 0, false, '["commanders_horn"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'By decree.', 'special'),
('ng_clear_weather', 'Clear Weather', 'Remove all weather effects', 'nilfgaard', 'special', NULL, 0, false, '["clear_weather"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Sun returns.', 'special'),
('ng_scorch', 'Scorch', 'Destroy highest power non-hero unit(s)', 'nilfgaard', 'special', NULL, 0, false, '["scorch"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Burn.', 'special'),
('ng_decoy', 'Decoy', 'Return a unit from your board to hand', 'nilfgaard', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Trickery.', 'special'),
('ng_mardroeme', 'Mardroeme', 'Transform a unit', 'nilfgaard', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Cursed.', 'special'),

-- ============================================================================
-- SCOIA'TAEL (16 cards)
-- ============================================================================

-- Heroes
('st_francesca', 'Francesca Findabair', 'Daisy of the Valley', 'scoiatael', 'hero', 'close', 4, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Filavandrel aen.jpg', 'Daisy of the Valley.', 'hero'),
('st_eithne', 'Eithne', 'Queen of Dryads', 'scoiatael', 'hero', 'close', 1, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Ida Emean aep Sivney.jpg', 'Queen of Dryads.', 'hero'),
('st_isfil', 'Isengrim Faoiltiarna', 'Commander of Scoia''tael', 'scoiatael', 'hero', 'close', 5, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Iorweth.jpg', 'Commander of Scoia''tael.', 'hero'),
('st_toruviel', 'Toruviel', 'Elven commando', 'scoiatael', 'hero', 'close', 4, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Toruviel.jpg', 'Freedom or death.', 'hero'),
('st_yaevinn', 'Yaevinn', 'Commando leader', 'scoiatael', 'unit', 'close', 5, false, '["morale"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Yaevinn.jpg', 'We watch.', 'close'),

-- Units
('st_commando', 'Commando', 'Guerrilla fighter', 'scoiatael', 'unit', 'close', 4, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Commando.jpg', 'Strike and fade.', 'close'),
('st_dragon_hunter', 'Dragon Hunter', 'Beast slayer', 'scoiatael', 'unit', 'close', 5, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Dragon Hunter.jpg', 'For the glens.', 'close'),
('st_elven_archer', 'Elven Archer', 'Deadly aim', 'scoiatael', 'unit', 'ranged', 4, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Elven Skirmisher.jpg', 'We wait.', 'ranged'),
('st_elven_scout', 'Elven Scout', 'Swift archer', 'scoiatael', 'unit', 'ranged', 2, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Elven Skirmisher.jpg', 'We watch.', 'ranged'),
('st_trapper', 'Trapper', 'Forest survivor', 'scoiatael', 'unit', 'ranged', 3, false, '["spy"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Contrebandier havekar.jpg', 'Nature provides.', 'ranged'),
('st_vrihedd_brigade', 'Vrihedd Brigade', 'Elite elven archers', 'scoiatael', 'unit', 'ranged', 4, false, '["tight_bond"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Vrihedd Brigade.jpg', 'Precision.', 'ranged'),
('st_dol_blathanna', 'Dol Blathanna Archer', 'Aen Seidhe marksman', 'scoiatael', 'unit', 'ranged', 3, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Dol Blathanna.jpg', 'From the glens.', 'ranged'),
('st_medic', 'Medic', 'Healer', 'scoiatael', 'unit', 'close', 3, false, '["medic"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Vesemir.jpg', 'Life finds a way.', 'close'),

-- Special Cards
('st_commanders_horn', 'Commander''s Horn', 'Boost all units on a row by 3', 'scoiatael', 'special', NULL, 0, false, '["commanders_horn"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'For the glens!', 'special'),
('st_clear_weather', 'Clear Weather', 'Remove all weather effects', 'scoiatael', 'special', NULL, 0, false, '["clear_weather"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Sun returns.', 'special'),
('st_scorch', 'Scorch', 'Destroy highest power non-hero unit(s)', 'scoiatael', 'special', NULL, 0, false, '["scorch"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Burn.', 'special'),
('st_decoy', 'Decoy', 'Return a unit from your board to hand', 'scoiatael', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Trickery.', 'special'),
('st_mardroeme', 'Mardroeme', 'Transform a unit', 'scoiatael', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Cursed.', 'special'),

-- ============================================================================
-- MONSTERS (17 cards)
-- ============================================================================

-- Heroes
('mo_eredin', 'Eredin Bréacc Glas', 'King of the Wild Hunt', 'monsters', 'hero', 'close', 1, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Hemdall.jpg', 'King of the Wild Hunt.', 'hero'),
('mo_glustyworp', 'Glustyworp', 'Nest keeper', 'monsters', 'hero', 'close', 1, true, '["leader"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Grave Hag.jpg', 'Nest keeper.', 'hero'),
('mo_crone', 'Crone', 'Ancient evil', 'monsters', 'unit', 'ranged', 5, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Celaeno Harpy.jpg', 'Weave fate.', 'ranged'),

-- Units with Muster
('mo_nekker', 'Nekker', 'Forest creature', 'monsters', 'unit', 'close', 2, false, '["muster"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Nekker.jpg', 'Never alone.', 'close'),
('mo_endrega', 'Endrega', 'Insectoid horror', 'monsters', 'unit', 'close', 3, false, '["muster"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Forktail.jpg', 'Swarm.', 'close'),
('mo_foglet', 'Foglet', 'Mist wraith', 'monsters', 'unit', 'ranged', 3, false, '["muster"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Foglet.jpg', 'In the mist.', 'ranged'),
('mo_arachas', 'Arachas', 'Giant scorpion', 'monsters', 'unit', 'close', 5, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Ice Giant.jpg', 'Chitin and venom.', 'close'),
('mo_ekimmara', 'Ekimmara', 'Vampire spawn', 'monsters', 'unit', 'close', 4, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/VampireEkimmara.jpg', 'Blood hunger.', 'close'),
('mo_vampire_bruxa', 'Vampire Bruxa', 'Higher vampire', 'monsters', 'unit', 'close', 6, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Vampire Bruxa.jpg', 'Night hunter.', 'close'),
('mo_grave_hag', 'Grave Hag', 'Carrion feeder', 'monsters', 'unit', 'ranged', 4, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Grave Hag.jpg', 'Death''s harvest.', 'ranged'),
('mo_celaeno_harpy', 'Celaeno Harpy', 'Mountain predator', 'monsters', 'unit', 'ranged', 3, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Celaeno Harpy.jpg', 'Screech and strike.', 'ranged'),

-- Special Cards
('mo_commanders_horn', 'Commander''s Horn', 'Boost all units on a row by 3', 'monsters', 'special', NULL, 0, false, '["commanders_horn"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Ride forth!', 'special'),
('mo_clear_weather', 'Clear Weather', 'Remove all weather effects', 'monsters', 'special', NULL, 0, false, '["clear_weather"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Sun returns.', 'special'),
('mo_scorch', 'Scorch', 'Destroy highest power non-hero unit(s)', 'monsters', 'special', NULL, 0, false, '["scorch"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Burn.', 'special'),
('mo_decoy', 'Decoy', 'Return a unit from your board to hand', 'monsters', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Trickery.', 'special'),
('mo_mardroeme', 'Mardroeme', 'Transform a unit', 'monsters', 'special', NULL, 0, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Clear Wheater.jpg', 'Cursed.', 'special'),

-- ============================================================================
-- NEUTRAL (14 cards)
-- ============================================================================

-- Heroes
('neut_geralt', 'Geralt of Rivia', 'White Wolf', 'neutral', 'hero', 'close', 15, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Sieg fried of Denesle.jpg', 'Evil is evil.', 'hero'),
('neut_ciri', 'Cirilla', 'Child of Destiny', 'neutral', 'hero', 'close', 15, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Yeneffer.jpg', 'Time and space.', 'hero'),
('neut_yennefer', 'Yennefer', 'Sorceress', 'neutral', 'hero', 'close', 7, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/philippa Eilhart.jpg', 'Power and beauty.', 'hero'),
('neut_triss', 'Triss Merigold', 'Sorceress', 'neutral', 'hero', 'close', 7, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Sabrina Glevissig.jpg', 'Fire and faith.', 'hero'),
('neut_zoltan', 'Zoltan Chivay', 'Dwarf warrior', 'neutral', 'hero', 'close', 6, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Yarpen Zigrin.jpg', 'Honor among thieves.', 'hero'),
('neut_dandelion', 'Dandelion', 'Bard', 'neutral', 'hero', 'close', 2, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Schirru.jpg', 'Silver tongue.', 'hero'),
('neut_vesemir', 'Vesemir', 'Old Witcher', 'neutral', 'hero', 'close', 6, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Vesemir.jpg', 'Experience.', 'hero'),
('neut_keira', 'Keira Metz', 'Sorceress', 'neutral', 'hero', 'close', 5, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Keira Metz.jpg', 'Secrets.', 'hero'),
('neut_thaler', 'Thaler', 'Spy', 'neutral', 'unit', 'close', 1, false, '["spy"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Sieg Technician.jpg', 'Information.', 'close'),
('neut_yarpen', 'Yarpen Zigrin', 'Dwarf commander', 'neutral', 'hero', 'close', 4, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Yarpen Zigrin.jpg', 'Leader.', 'hero'),
('neut_shani', 'Shani', 'Medic', 'neutral', 'unit', 'close', 3, false, '["medic"]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Ves.jpg', 'Healing hands.', 'close'),
('neut_gaunter', 'Gaunter O''Dimm', 'Master Mirror', 'neutral', 'hero', 'close', 1, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Gaunter O Dimm.jpg', 'Time is a flat circle.', 'hero'),
('neut_kambi', 'Kambi', 'Reckless warrior', 'neutral', 'hero', 'close', 12, true, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/Kambi.jpg', 'No fear.', 'hero'),
('neut_poor_infantry', 'Poor Fucking Infantry', 'Basic foot soldier', 'neutral', 'unit', 'close', 1, false, '[]', 'https://github.com/mohamad801114/gwent-bot/assets/images/poor fucking infantry.jpg', 'Every army needs fodder.', 'close');

-- ============================================================================
-- KEYWORDS REFERENCE (ensure they exist)
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