# Gwent Bot - Debug & Progress Document
**Last Updated:** 2026-11-17  
**Session:** Database connection issues and button callback failures  
**Environment:** Docker on Linux VPS (planned)

---

## CRITICAL ISSUES IDENTIFIED

### Issue 1: Buttons Not Working (Callbacks Fail Silently)
**Symptoms:** 
- Clicking Buy Cards, Admin Cards, or any inline button does nothing
- No error messages, no response
- Cards exist in DB but not displayed

**Root Causes Found:**
1. Callback queries not being answered consistently - answerCallback() is called in some handlers but NOT in OnCallback router for all paths
2. Telegram requires callback query response within 30s - if not answered, button stays in loading state
3. Missing c.Bot().Respond() calls in several callback handlers
4. Error handling swallows errors - c.Edit() failures fallback to c.Send() but errors not logged properly

**Evidence in Code:**
In OnCallback - answers callback but then routes to handlers that DONT answer
func (h *Handlers) OnCallback(c tele.Context) error {
    cb := c.Callback()
    if cb != nil {
        // Answers HERE but then handlers below dont answer their own callbacks
        if err := c.Bot().Respond(cb, &tele.CallbackResponse{Text: })); err != nil {
            log.Printf(ERROR: Failed to answer callback query: %v, err)
        }
        // ... routes to handleBattleCallback, handleShopCallback, etc.
    }
}

But handleShopCallback calls OnShopCards which calls c.Edit() - if edit fails, tries c.Send()
NEITHER answers the callback query again!

### Issue 2: Database Queries Return Empty Results
**Symptoms:**
- /admin_cards shows empty list
- Buy Cards shows empty shop
- But migration seeds 85 cards + 11 shop items

**Potential Causes:**
1. Migration not applied - Docker volume may have old schema without seeds
2. Transaction isolation - Migration runs in tx but handlers use different connection
3. Context timeout - context.Background() no timeout, could hang
4. Column mismatch - Models expect rarity, provision, armor but migration 002 dropped them (per PROGRESS.md)

**Evidence in PROGRESS.md:**
Known Issues
- Migration 002 drops provision and armor columns but new models expect them
- Need to either update migration 002 or adjust models to work without them

### Issue 3: No Health Check Endpoint
**Docker Healthcheck fails:**
HEALTHCHECK CMD wget --spider http://localhost:8080/health || exit 1
But NO /health HTTP handler exists in main.go!

---

## DEBUG SYSTEM TO IMPLEMENT

### 1. Structured Logging Wrapper
internal/debug/logger.go
type DebugLogger struct {
    enabled bool
    prefix  string
}

func (d *DebugLogger) Logf(format string, args ...any) {
    if d.enabled {
        log.Printf([DEBUG:%s] +format, append([]any{d.prefix}, args...)...)
    }
}

func (d *DebugLogger) Errorf(format string, args ...any) {
    log.Printf([ERROR:%s] +format, append([]any{d.prefix}, args...)...)
}

### 2. Database Query Debugging
- Log every SQL query with parameters
- Log row counts returned
- Log query duration
- Add ?debug=1 query param to admin endpoints

### 3. Callback Debug Middleware
func (h *Handlers) DebugCallbackMiddleware(next tele.HandlerFunc) tele.HandlerFunc {
    return func(c tele.Context) error {
        if cb := c.Callback(); cb != nil {
            log.Printf(CALLBACK: data=%q from_user=%d chat=%d, 
                cb.Data, cb.Sender.ID, cb.Message.Chat.ID)
            // ALWAYS answer immediately
            _ = c.Bot().Respond(cb, &tele.CallbackResponse{Text: Processing...})
        }
        return next(c)
    }
}

### 4. Request/Response Logging
- Log all incoming updates (message, callback, inline query)
- Log all outgoing messages (success/failure)
- Track handler execution time

### 5. Database Connection Health
func (h *Handlers) checkDB(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    return h.db.Pool.Ping(ctx)
}

---

## FIXES TO APPLY (Priority Order)

### P0 - Critical (Blocks all functionality)
1. Add HTTP health endpoint in main.go (/health returning 200 OK)
2. Fix callback answering - ensure EVERY callback path calls c.Bot().Respond()
3. Verify migration applied - check schema_migrations table has 000_complete_schema

### P1 - High (Fixes empty results)
4. Verify database schema matches models - check columns: rarity, provision, armor, nation
5. Add debug logging to GetAllCards, GetShopItems - log row count, first few results
6. Fix context usage - use context.WithTimeout(context.Background(), 10*time.Second) instead of bare Background()

### P2 - Medium (Improves reliability)
7. Add structured debug logger package
8. Add request/response logging middleware for telebot
9. Fix error handling - dont swallow errors in c.Edit/c.Send fallbacks
10. Add admin debug command /debug showing DB stats, config, active battles

### P3 - Low (Nice to have)
11. Add metrics (battle count, active users, errors)
12. Add structured JSON logging for production

---

## DOCKER DEPLOYMENT CHECKLIST

### Pre-deployment (Local Verification)
- [ ] go build ./cmd/bot - compiles without errors
- [ ] docker-compose up --build - builds successfully
- [ ] docker-compose logs postgres - shows database system is ready to accept connections
- [ ] docker-compose exec postgres psql -U gwent_user -d gwent_bot -c SELECT version FROM schema_migrations;
- [ ] docker-compose exec postgres psql -U gwent_user -d gwent_bot -c SELECT count(*) FROM cards;
- [ ] docker-compose exec postgres psql -U gwent_user -d gwent_bot -c SELECT count(*) FROM shop_items;
- [ ] curl http://localhost:8080/health - returns 200 OK

### VPS Deployment
- [ ] Copy .env.example to .env with real values
- [ ] docker-compose -f docker-compose.yml up -d --build
- [ ] docker-compose logs -f bot - verify Gwent Bot started! message
- [ ] Test /start in Telegram
- [ ] Test Buy Cards button
- [ ] Test /admin_cards command

### Environment Variables Required
BOT_TOKEN=<from BotFather>
BOT_MODE=polling
DATABASE_URL=postgres://gwent_user:gwent_password@postgres:5432/gwent_bot?sslmode=disable
ADMIN_USER_IDS=<your_telegram_id>
... rest from .env.example

---

## CURRENT CODE STATE ANALYSIS

### Files Needing Immediate Fixes

| File | Issue | Fix |
|------|-------|-----|
| cmd/bot/main.go | No /health endpoint | Add HTTP server with /health |
| internal/telegram/handlers/handlers.go | Callback not answered in all paths | Add answerCallback to ALL callback handlers |
| internal/telegram/handlers/handlers.go | context.Background() no timeout | Use context.WithTimeout |
| internal/database/card_repo.go | No debug logging | Add debug logs to GetAllCards, GetShopItems |
| migrations/000_complete_schema.sql | May not match models | Verify columns: rarity, provision, armor |

### Database Schema vs Models Check

**Models expect (pkg/models/models.go):**
type Card struct {
    ID          string
    Name        string
    Description string
    Faction     Faction
    Nation      Nation        // alias for Faction
    Type        CardType
    Row         RowType
    Power       int
    Armor       int           // Migration 002 dropped this!
    Provision   int           // Migration 002 dropped this!
    IsHero      bool
    Abilities   []string
    ArtURL      string
    FlavorText  string
    CardGroup   string
    Rarity      Rarity        // Migration 002 dropped this!
    CostGold    int
    CostScrap   int
}

**Migration 000 has:**
CREATE TABLE cards (
    -- ...
    power SMALLINT NOT NULL DEFAULT 0,
    is_hero BOOLEAN NOT NULL DEFAULT FALSE,
    abilities JSONB NOT NULL DEFAULT []::jsonb,
    art_url VARCHAR(500),
    flavor_text TEXT,
    card_group VARCHAR(50),
    rarity VARCHAR(20),           -- EXISTS
    cost_gold BIGINT NOT NULL DEFAULT 0,
    cost_scrap BIGINT NOT NULL DEFAULT 0,
    -- ...
);
-- Legacy column for backward compatibility
ALTER TABLE cards ADD COLUMN IF NOT EXISTS nation VARCHAR(50);

**MISSING COLUMNS:** armor, provision - Migration 000 has power but not armor/provision!

But models expect Armor and Provision fields!

---

## IMMEDIATE ACTION PLAN

### Step 1: Fix Database Schema (Run on VPS after deploy)
ALTER TABLE cards ADD COLUMN IF NOT EXISTS armor SMALLINT NOT NULL DEFAULT 0;
ALTER TABLE cards ADD COLUMN IF NOT EXISTS provision SMALLINT NOT NULL DEFAULT 0;

UPDATE cards SET 
    armor = CASE WHEN is_hero THEN power ELSE 0 END,
    provision = CASE 
        WHEN rarity = legendary THEN 10
        WHEN rarity = epic THEN 8
        WHEN rarity = rare THEN 6
        ELSE 4
    END
WHERE armor = 0 OR provision = 0;

### Step 2: Add Health Endpoint to main.go
// Add before bot.Start()
go func() {
    http.HandleFunc(/health, func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(OK))
    })
    log.Fatal(http.ListenAndServe(:8080, nil))
}()

### Step 3: Fix Callback Answering in handlers.go
- Add h.answerCallback(c) at START of every callback handler
- Ensure OnCallback router doesnt double-answer

### Step 4: Add Debug Logging
- Create internal/debug/logger.go
- Add to all repository methods
- Add to all handler entry points

### Step 5: Test and Deploy
- Build locally
- Deploy to VPS
- Verify with debug commands

---

## NEXT SESSION PRIORITIES (If Terminal Closes)

1. Apply schema fix on VPS PostgreSQL
2. Add /health endpoint to main.go
3. Fix callback answering throughout handlers.go
4. Add debug logger package
5. Instrument repository methods with debug logging
6. Deploy and test with /debug command
7. Verify all buttons work: Buy Cards, Admin Cards, Battle, Mine, Chest, Leader

---

## DEBUG COMMANDS TO ADD

// /debug - shows system status
func (h *Handlers) OnDebug(c tele.Context) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // DB health
    dbOk := h.db.Pool.Ping(ctx) == nil
    
    // Count cards
    var cardCount int
    h.db.Pool.QueryRow(ctx, SELECT count(*) FROM cards).Scan(&cardCount)
    
    // Count shop items
    var shopCount int
    h.db.Pool.QueryRow(ctx, SELECT count(*) FROM shop_items).Scan(&shopCount)
    
    // Active battles
    var battleCount int
    h.db.Pool.QueryRow(ctx, SELECT count(*) FROM battles WHERE phase != finished).Scan(&battleCount)
    
    msg := fmt.Sprintf(DEBUG Info
    
DB Connected: %v
Cards in DB: %d
Shop Items: %d
Active Battles: %d
Bot Mode: %s
Admin IDs: %v
, dbOk, cardCount, shopCount, battleCount, h.cfg.BotMode, h.cfg.AdminUserIDs)
    
    return c.Send(msg, tele.ModeHTML)
}

---

This document serves as the single source of truth for debugging. If terminal closes, resume from IMMEDIATE ACTION PLAN section.
