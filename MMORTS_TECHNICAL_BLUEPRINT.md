# MMORTS Game Technical Blueprint
## Space Empire Builder - "Galactic Dominion"

---

## Table of Contents
1. [Executive Summary](#executive-summary)
2. [Core Concept](#core-concept)
3. [Technology Stack](#technology-stack)
4. [System Architecture](#system-architecture)
5. [Complex Features - Deep Dive](#complex-features-deep-dive)
6. [Database Architecture](#database-architecture)
7. [Server-Client Synchronization Model](#server-client-synchronization-model)
8. [Performance Optimization Strategies](#performance-optimization-strategies)
9. [Security & Anti-Cheat](#security-anti-cheat)
10. [Scalability Considerations](#scalability-considerations)
11. [Development Roadmap](#development-roadmap)

---

## 1. Executive Summary

**Galactic Dominion** is a browser-based MMORTS game where players build space colonies, manage complex economies, engage in real-time combat, and form powerful alliances. The game leverages cutting-edge web technologies to deliver a seamless, high-performance experience capable of supporting tens of thousands of concurrent players.

**Key Technical Pillars:**
- **Performance**: WebAssembly for computation-heavy operations
- **Real-time Communication**: WebSocket-based event streaming with delta compression
- **Scalability**: Microservices architecture with horizontal scaling
- **Rendering**: Babylon.js for 3D space visualization with LOD optimization
- **Data Persistence**: Hybrid SQL/NoSQL architecture for different data patterns

---

## 2. Core Concept

### Game Loop Overview
Players start with a single colony on a resource-rich planet. Through resource extraction, research, and diplomacy, they expand their empire across the galaxy. The game operates on multiple time scales:

- **Real-time**: Combat, market transactions, diplomacy
- **Near-real-time** (tick-based): Resource production, construction, research (5-second ticks)
- **Asynchronous**: Alliance governance, long-term strategic planning

### Core Gameplay Pillars

#### Colony Management
- Multiple resource types: Energy, Minerals, Rare Elements, Population
- Building construction with dependencies and upgrade paths
- Research trees affecting combat, economy, and expansion

#### Economic Warfare
- Player-driven market with supply/demand dynamics
- Trade routes and blockade mechanics
- Resource scarcity based on galactic regions (core vs. rim)

#### Military Operations
- Real-time fleet combat with physics-based movement
- Unit composition strategies (rock-paper-scissors mechanics)
- Territory control and resource denial

#### Alliance Politics
- Democratic governance with voting systems
- Role-based permissions (Commander, Diplomat, Treasurer, Member)
- Shared resources and coordinated war declarations

---

## 3. Technology Stack

### Frontend

#### Core Technologies
```javascript
{
  "engine": "Babylon.js 7.x",
  "rationale": "Superior 3D rendering, physics engine, WebGPU support",
  "alternatives_considered": ["Three.js", "Phaser (2D-focused)"]
}
```

**Babylon.js Benefits:**
- Built-in physics engine (Havok/Cannon.js integration)
- Optimized for large scene rendering (crucial for world map)
- WebGPU support for future-proofing
- Strong TypeScript support

#### UI Framework
```javascript
{
  "framework": "React 18+",
  "state_management": "Zustand with Immer",
  "ui_library": "Mantine UI + Custom components",
  "rationale": "React's ecosystem + Zustand's simplicity for game state"
}
```

#### High-Performance Computing
```javascript
{
  "wasm_framework": "AssemblyScript / Rust (wasm-bindgen)",
  "use_cases": [
    "Pathfinding algorithms (A* for 100k+ nodes)",
    "Combat calculations (projectile physics)",
    "Market simulation (supply/demand calculation)",
    "Map chunk compression/decompression"
  ]
}
```

**WebAssembly Module Structure:**
```
wasm_modules/
├── pathfinding.wasm      // A* implementation
├── combat_sim.wasm       // Damage calculation, projectile physics
├── economy_engine.wasm   // Price calculation, supply-demand curves
└── compression.wasm      // LZ4 compression for map data
```

#### Communication Layer
```javascript
{
  "primary": "WebSocket (Socket.IO for fallback)",
  "protocol": "Binary (MessagePack/Protobuf)",
  "rationale": "~60% smaller payload vs JSON, faster parsing"
}
```

### Backend

#### Core Server Architecture
```javascript
{
  "language": "Node.js (TypeScript) + Rust microservices",
  "framework": "NestJS for API, Custom WS server",
  "rationale": "NestJS provides structure, Rust for performance-critical services"
}
```

**Microservices Breakdown:**
```
services/
├── api_gateway/          // Node.js (NestJS) - HTTP REST API
├── websocket_gateway/    // Node.js (Socket.IO) - Real-time events
├── game_world_service/   // Rust - Tick processing, map state
├── combat_service/       // Rust - Combat resolution
├── economy_service/      // Node.js - Market calculations
├── alliance_service/     // Node.js - Governance, voting
└── notification_service/ // Node.js - Push notifications, emails
```

#### Game Server (Rust - Performance Critical)
```rust
// High-level structure
use tokio; // Async runtime
use actix_web; // Web framework

struct GameWorldServer {
    tick_rate: u16, // 200ms per tick (5 ticks/sec)
    spatial_hash: SpatialHashMap, // O(1) neighbor queries
    entity_registry: EntityComponentSystem,
}

// Handles 10k+ entities per tick
impl GameWorldServer {
    async fn process_tick(&mut self) {
        // 1. Process resource production (parallel)
        // 2. Update building construction
        // 3. Process fleet movements
        // 4. Trigger combat events
        // 5. Broadcast delta updates
    }
}
```

### Database Architecture

#### Hybrid Approach
```javascript
{
  "sql": "PostgreSQL 16+",
  "nosql": "Redis + MongoDB",
  "time_series": "TimescaleDB (PostgreSQL extension)",
  "search": "Elasticsearch"
}
```

**Data Distribution Strategy:**

| Data Type | Database | Reasoning |
|-----------|----------|-----------|
| User accounts, auth | PostgreSQL | ACID compliance, relational integrity |
| Game state (colonies, fleets) | PostgreSQL + Redis cache | Structured data with caching layer |
| Real-time sessions | Redis | In-memory speed for active players |
| Historical events, logs | TimescaleDB | Efficient time-series queries |
| Market transactions | MongoDB | High write throughput, flexible schema |
| Alliance chat | MongoDB | Document-based, append-heavy |
| Player search | Elasticsearch | Full-text search, geo-queries |

#### Database Schema (PostgreSQL - Core Tables)

```sql
-- Users and Authentication
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(32) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    last_login TIMESTAMP,
    is_banned BOOLEAN DEFAULT false
);

-- Player Profile and Resources
CREATE TABLE players (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    display_name VARCHAR(64),
    faction VARCHAR(32),
    total_power BIGINT DEFAULT 0, -- Calculated score
    alliance_id BIGINT REFERENCES alliances(id) ON DELETE SET NULL,

    -- Resources (updated each tick)
    energy BIGINT DEFAULT 1000,
    minerals BIGINT DEFAULT 1000,
    rare_elements BIGINT DEFAULT 100,
    population BIGINT DEFAULT 500,

    -- Resource production rates (per tick)
    energy_rate INTEGER DEFAULT 10,
    minerals_rate INTEGER DEFAULT 10,
    rare_elements_rate INTEGER DEFAULT 1,
    population_rate INTEGER DEFAULT 2,

    last_tick_processed TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id)
);

-- Colonies (Players can have multiple)
CREATE TABLE colonies (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT REFERENCES players(id) ON DELETE CASCADE,
    name VARCHAR(64) NOT NULL,

    -- Location in galaxy
    galaxy_x INTEGER NOT NULL,
    galaxy_y INTEGER NOT NULL,
    sector_id INTEGER, -- For spatial partitioning

    -- Colony attributes
    planet_type VARCHAR(32), -- terrestrial, gas_giant, asteroid
    resource_bonus JSONB, -- {"minerals": 1.5, "energy": 0.8}

    founded_at TIMESTAMP DEFAULT NOW(),
    defense_rating INTEGER DEFAULT 100,

    UNIQUE(galaxy_x, galaxy_y)
);
CREATE INDEX idx_colonies_location ON colonies(sector_id, galaxy_x, galaxy_y);
CREATE INDEX idx_colonies_player ON colonies(player_id);

-- Buildings (many-to-one with colonies)
CREATE TABLE buildings (
    id BIGSERIAL PRIMARY KEY,
    colony_id BIGINT REFERENCES colonies(id) ON DELETE CASCADE,
    building_type VARCHAR(32) NOT NULL, -- mine, power_plant, shipyard
    level INTEGER DEFAULT 1,

    -- Construction state
    is_constructing BOOLEAN DEFAULT false,
    construction_started_at TIMESTAMP,
    construction_complete_at TIMESTAMP,

    -- Production contribution
    resource_output JSONB, -- {"minerals": 50}

    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_buildings_colony ON buildings(colony_id);

-- Fleets (groups of ships)
CREATE TABLE fleets (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT REFERENCES players(id) ON DELETE CASCADE,
    name VARCHAR(64),

    -- Current location
    current_x DECIMAL(10,2) NOT NULL, -- Sub-grid precision for movement
    current_y DECIMAL(10,2) NOT NULL,

    -- Movement state
    is_moving BOOLEAN DEFAULT false,
    destination_x DECIMAL(10,2),
    destination_y DECIMAL(10,2),
    arrival_time TIMESTAMP,

    -- Fleet composition (JSON for flexibility)
    ships JSONB, -- {"fighter": 50, "cruiser": 10, "battleship": 2}

    -- Combat stats (cached calculation)
    total_attack INTEGER DEFAULT 0,
    total_defense INTEGER DEFAULT 0,
    total_speed INTEGER DEFAULT 100,

    status VARCHAR(32) DEFAULT 'idle', -- idle, moving, in_combat, returning

    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_fleets_location ON fleets(current_x, current_y);
CREATE INDEX idx_fleets_player ON fleets(player_id);

-- Alliances
CREATE TABLE alliances (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(64) UNIQUE NOT NULL,
    tag VARCHAR(6) UNIQUE NOT NULL, -- [TAG]
    founder_id BIGINT REFERENCES players(id),

    description TEXT,
    member_count INTEGER DEFAULT 1,
    total_power BIGINT DEFAULT 0,

    -- Governance settings
    governance_type VARCHAR(32) DEFAULT 'democracy', -- autocracy, democracy, council
    voting_threshold DECIMAL(3,2) DEFAULT 0.51, -- 51% majority

    created_at TIMESTAMP DEFAULT NOW()
);

-- Alliance Membership with Roles
CREATE TABLE alliance_members (
    id BIGSERIAL PRIMARY KEY,
    alliance_id BIGINT REFERENCES alliances(id) ON DELETE CASCADE,
    player_id BIGINT REFERENCES players(id) ON DELETE CASCADE,

    role VARCHAR(32) DEFAULT 'member', -- leader, commander, diplomat, treasurer, member
    permissions JSONB, -- {"can_invite": true, "can_declare_war": false}

    joined_at TIMESTAMP DEFAULT NOW(),
    contribution_points INTEGER DEFAULT 0,

    UNIQUE(alliance_id, player_id)
);
CREATE INDEX idx_alliance_members_alliance ON alliance_members(alliance_id);
CREATE INDEX idx_alliance_members_player ON alliance_members(player_id);

-- Alliance Voting System
CREATE TABLE alliance_votes (
    id BIGSERIAL PRIMARY KEY,
    alliance_id BIGINT REFERENCES alliances(id) ON DELETE CASCADE,
    created_by BIGINT REFERENCES players(id),

    vote_type VARCHAR(32) NOT NULL, -- declare_war, accept_member, promote_member, change_governance
    description TEXT,
    target_data JSONB, -- Flexible data for vote target

    votes_for INTEGER DEFAULT 0,
    votes_against INTEGER DEFAULT 0,
    required_votes INTEGER, -- Calculated at creation

    status VARCHAR(32) DEFAULT 'active', -- active, passed, failed, expired
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Individual Vote Records
CREATE TABLE vote_records (
    id BIGSERIAL PRIMARY KEY,
    vote_id BIGINT REFERENCES alliance_votes(id) ON DELETE CASCADE,
    player_id BIGINT REFERENCES players(id) ON DELETE CASCADE,

    vote_choice BOOLEAN, -- true = for, false = against
    voted_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(vote_id, player_id)
);

-- Market Orders
CREATE TABLE market_orders (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT REFERENCES players(id) ON DELETE CASCADE,

    order_type VARCHAR(4) NOT NULL, -- BUY, SELL
    resource_type VARCHAR(32) NOT NULL,
    quantity BIGINT NOT NULL,
    price_per_unit DECIMAL(10,2) NOT NULL, -- In "Credits" (virtual currency)

    filled_quantity BIGINT DEFAULT 0,
    status VARCHAR(32) DEFAULT 'active', -- active, partially_filled, filled, cancelled

    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP
);
CREATE INDEX idx_market_orders_resource ON market_orders(resource_type, order_type, status);
CREATE INDEX idx_market_orders_player ON market_orders(player_id);

-- Market Transaction History (TimescaleDB hypertable)
CREATE TABLE market_transactions (
    id BIGSERIAL,
    buyer_id BIGINT REFERENCES players(id),
    seller_id BIGINT REFERENCES players(id),

    resource_type VARCHAR(32) NOT NULL,
    quantity BIGINT NOT NULL,
    price_per_unit DECIMAL(10,2) NOT NULL,
    total_value DECIMAL(15,2) NOT NULL,

    transaction_time TIMESTAMP NOT NULL DEFAULT NOW()
);
-- Convert to hypertable for time-series optimization
SELECT create_hypertable('market_transactions', 'transaction_time');
CREATE INDEX idx_market_tx_resource ON market_transactions(resource_type, transaction_time DESC);

-- Combat Events
CREATE TABLE combat_events (
    id BIGSERIAL PRIMARY KEY,
    attacker_fleet_id BIGINT REFERENCES fleets(id),
    defender_fleet_id BIGINT REFERENCES fleets(id),
    defender_colony_id BIGINT REFERENCES colonies(id),

    location_x INTEGER NOT NULL,
    location_y INTEGER NOT NULL,

    combat_log JSONB, -- Detailed turn-by-turn results
    victor VARCHAR(32), -- attacker, defender, draw

    attacker_losses JSONB, -- Ship losses
    defender_losses JSONB,
    resources_looted JSONB,

    started_at TIMESTAMP DEFAULT NOW(),
    ended_at TIMESTAMP
);
CREATE INDEX idx_combat_events_time ON combat_events(started_at DESC);

-- Research Progress
CREATE TABLE research (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT REFERENCES players(id) ON DELETE CASCADE,
    tech_id VARCHAR(32) NOT NULL, -- Corresponds to game's tech tree

    level INTEGER DEFAULT 0,
    is_researching BOOLEAN DEFAULT false,
    research_started_at TIMESTAMP,
    research_complete_at TIMESTAMP,

    UNIQUE(player_id, tech_id)
);
CREATE INDEX idx_research_player ON research(player_id);

-- Notification Queue
CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT REFERENCES players(id) ON DELETE CASCADE,

    notification_type VARCHAR(32) NOT NULL, -- combat_report, trade_completed, alliance_invite
    title VARCHAR(128),
    message TEXT,
    data JSONB, -- Additional context

    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_notifications_player ON notifications(player_id, is_read, created_at DESC);
```

#### Redis Data Structures (Caching & Real-time State)

```javascript
// Active player sessions
redis.hset(`session:${userId}`, {
    socketId: 'abc123',
    lastSeen: Date.now(),
    currentView: 'galaxy_map', // For targeted updates
    visibleSector: 42
});

// Map sector cache (expires after 60s)
redis.setex(`sector:${sectorId}:entities`, 60, JSON.stringify({
    colonies: [...],
    fleets: [...]
}));

// Real-time market snapshot (updated each tick)
redis.hset('market:prices', 'minerals', 15.47);
redis.hset('market:volume', 'minerals', 1500000);

// Combat lock (prevent double-processing)
redis.setex(`combat:lock:fleet:${fleetId}`, 30, '1');

// Leaderboard (sorted sets)
redis.zadd('leaderboard:power', playerPower, playerId);
redis.zrevrange('leaderboard:power', 0, 99); // Top 100
```

#### MongoDB Collections (Flexible/High-Throughput Data)

```javascript
// Alliance chat messages
db.alliance_chat.insertOne({
    alliance_id: 123,
    player_id: 456,
    player_name: "CommanderX",
    message: "Attack coordinates (1500, 2300) at 14:00 UTC",
    timestamp: ISODate("2025-11-03T12:00:00Z"),
    mentions: [789], // Player IDs mentioned
});
db.alliance_chat.createIndex({alliance_id: 1, timestamp: -1});

// Player activity logs (analytics)
db.player_activity.insertOne({
    player_id: 456,
    event_type: "fleet_deployed",
    event_data: {
        fleet_id: 789,
        destination: {x: 1500, y: 2300}
    },
    timestamp: ISODate("2025-11-03T12:00:00Z")
});
db.player_activity.createIndex({player_id: 1, timestamp: -1});
db.player_activity.createIndex({event_type: 1, timestamp: -1});

// Game world snapshots (daily backups)
db.world_snapshots.insertOne({
    snapshot_date: ISODate("2025-11-03T00:00:00Z"),
    total_players: 50000,
    total_colonies: 125000,
    total_fleets: 85000,
    sector_data: [...] // Compressed binary data
});
```

---

## 4. System Architecture

### High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Babylon.js │  │   React UI   │  │ WASM Modules │         │
│  │  (Rendering) │  │   (HUD/UI)   │  │  (Compute)   │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
│         │                  │                  │                  │
│         └──────────────────┼──────────────────┘                  │
│                            │                                     │
│                   ┌────────▼─────────┐                          │
│                   │ State Management │                          │
│                   │    (Zustand)     │                          │
│                   └────────┬─────────┘                          │
└────────────────────────────┼──────────────────────────────────┘
                             │
                   ┌─────────▼──────────┐
                   │   WebSocket (WSS)  │
                   │   + MessagePack    │
                   └─────────┬──────────┘
                             │
┌────────────────────────────┼──────────────────────────────────┐
│                     GATEWAY LAYER                              │
│              ┌──────────────┴────────────┐                     │
│              │                           │                     │
│     ┌────────▼─────────┐     ┌──────────▼────────┐           │
│     │  API Gateway     │     │ WebSocket Gateway │           │
│     │  (NestJS/REST)   │     │   (Socket.IO)     │           │
│     └────────┬─────────┘     └──────────┬────────┘           │
└──────────────┼────────────────────────────┼──────────────────┘
               │                            │
        ┌──────┴────────────────────────────┴──────┐
        │         Message Bus (Redis Pub/Sub)      │
        │              + Event Queue               │
        └──────┬────────────────────────────┬──────┘
               │                            │
┌──────────────┼────────────────────────────┼──────────────────┐
│                    SERVICE LAYER                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ Game World  │  │   Combat    │  │   Economy   │          │
│  │  Service    │  │   Service   │  │   Service   │          │
│  │   (Rust)    │  │   (Rust)    │  │  (Node.js)  │          │
│  └─────┬───────┘  └─────┬───────┘  └─────┬───────┘          │
│        │                │                 │                   │
│  ┌─────▼───────┐  ┌─────▼───────┐  ┌─────▼───────┐          │
│  │  Alliance   │  │Notification │  │  Analytics  │          │
│  │   Service   │  │   Service   │  │   Service   │          │
│  │  (Node.js)  │  │  (Node.js)  │  │  (Node.js)  │          │
│  └─────┬───────┘  └─────┬───────┘  └─────┬───────┘          │
└────────┼────────────────┼─────────────────┼──────────────────┘
         │                │                 │
         └────────────────┼─────────────────┘
                          │
┌─────────────────────────┼──────────────────────────────────┐
│                   DATA LAYER                                │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐           │
│  │ PostgreSQL │  │   Redis    │  │  MongoDB   │           │
│  │  (Master)  │  │  (Cache +  │  │  (Logs +   │           │
│  │     +      │  │   Queue)   │  │   Chat)    │           │
│  │  Replicas  │  │            │  │            │           │
│  └────────────┘  └────────────┘  └────────────┘           │
│                                                             │
│  ┌────────────┐  ┌────────────┐                           │
│  │TimescaleDB │  │Elasticsearch│                           │
│  │(Time Series│  │  (Search)   │                           │
│  └────────────┘  └────────────┘                           │
└─────────────────────────────────────────────────────────────┘
```

### Service Communication Patterns

#### Synchronous Communication (REST)
- Client authentication
- Static data queries (tech tree, ship stats)
- Administrative operations

#### Asynchronous Communication (WebSocket + Message Bus)
- Real-time game state updates
- Combat events
- Chat messages
- Market price changes

#### Event-Driven Architecture

```javascript
// Example: Fleet reaches destination
// 1. Game World Service detects arrival (tick processing)
gameWorldService.emit('fleet.arrived', {
    fleetId: 123,
    playerId: 456,
    location: {x: 1500, y: 2300},
    destinationEntity: {type: 'colony', id: 789}
});

// 2. Combat Service subscribes and checks for hostilities
combatService.on('fleet.arrived', async (data) => {
    const enemies = await findHostileEntities(data.location);
    if (enemies.length > 0) {
        await initiateCombat(data.fleetId, enemies);
    }
});

// 3. Notification Service alerts player
notificationService.on('fleet.arrived', async (data) => {
    await sendNotification(data.playerId, {
        type: 'fleet_arrived',
        message: `Your fleet has reached (${data.location.x}, ${data.location.y})`
    });
});

// 4. WebSocket Gateway broadcasts to nearby players
wsGateway.on('fleet.arrived', async (data) => {
    const nearbyPlayers = await getPlayersInSector(data.location);
    nearbyPlayers.forEach(playerId => {
        wsGateway.sendToPlayer(playerId, 'entity.updated', {
            type: 'fleet',
            id: data.fleetId,
            position: data.location
        });
    });
});
```

---

## 5. Complex Features - Deep Dive

### 5.1 Intricate Combat System

#### Design Principles
1. **Server Authority**: All combat calculations happen server-side
2. **Deterministic**: Same inputs → same outputs (for replay/verification)
3. **Real-time Feel**: Client prediction + server reconciliation
4. **Physics-Lite**: Balance between realism and performance

#### Combat Flow

```
┌──────────────┐
│ Fleet enters │
│ combat zone  │
└──────┬───────┘
       │
       ▼
┌──────────────────────┐
│ Server: Lock entities│ ◄─── Prevents race conditions
│ (Redis distributed   │
│  lock)               │
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│ Initialize Combat    │
│ Instance             │
│ - Load fleet data    │
│ - Calculate stats    │
│ - Set up formations  │
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│ Combat Loop          │ ◄────────┐
│ (100ms ticks)        │          │
│                      │          │
│ 1. Movement phase    │          │
│ 2. Targeting phase   │          │
│ 3. Attack phase      │          │
│ 4. Damage resolution │          │
│ 5. Victory check     │          │
└──────┬───────────────┘          │
       │                          │
       ├── Continue? ─────────────┘
       │
       ▼
┌──────────────────────┐
│ Combat Resolution    │
│ - Calculate losses   │
│ - Loot distribution  │
│ - Generate report    │
│ - Update databases   │
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│ Broadcast Results    │
│ - Send to            │
│   participants       │
│ - Update map state   │
│ - Release locks      │
└──────────────────────┘
```

#### Combat Engine (Rust Implementation)

```rust
use nalgebra::{Vector2, Point2}; // Linear algebra for physics

#[derive(Clone)]
struct Ship {
    id: u64,
    ship_type: ShipType,
    position: Point2<f32>,
    velocity: Vector2<f32>,
    heading: f32, // Radians

    // Stats
    max_hp: u32,
    current_hp: u32,
    attack: u32,
    defense: u32,
    speed: f32,
    range: f32,

    // State
    target_id: Option<u64>,
    last_fired: u64, // Tick number
    fire_rate: u64, // Ticks between shots
}

struct CombatInstance {
    id: u64,
    attackers: Vec<Ship>,
    defenders: Vec<Ship>,
    tick_count: u64,
    max_ticks: u64, // Timeout (e.g., 6000 = 10 minutes)

    combat_log: Vec<CombatEvent>,
}

impl CombatInstance {
    // Called every 100ms
    fn process_tick(&mut self) {
        self.tick_count += 1;

        // Phase 1: Movement with steering behaviors
        self.update_movement();

        // Phase 2: Intelligent targeting
        self.update_targeting();

        // Phase 3: Fire weapons
        self.process_attacks();

        // Phase 4: Check victory condition
        if self.is_combat_over() {
            self.finalize_combat();
        }
    }

    fn update_movement(&mut self) {
        let all_ships = self.attackers.iter_mut()
            .chain(self.defenders.iter_mut());

        for ship in all_ships {
            if ship.current_hp == 0 { continue; }

            // Calculate desired velocity (steering toward target)
            let desired_velocity = if let Some(target_id) = ship.target_id {
                self.calculate_pursuit_vector(ship, target_id)
            } else {
                Vector2::zeros()
            };

            // Apply acceleration (simplified physics)
            ship.velocity += (desired_velocity - ship.velocity) * 0.1; // Smoothing
            ship.velocity = ship.velocity.cap_magnitude(ship.speed);

            // Update position
            ship.position += ship.velocity * 0.1; // 100ms = 0.1s

            // Update heading
            if ship.velocity.magnitude() > 0.1 {
                ship.heading = ship.velocity.y.atan2(ship.velocity.x);
            }
        }
    }

    fn calculate_pursuit_vector(&self, ship: &Ship, target_id: u64) -> Vector2<f32> {
        // Find target
        let target = self.find_ship(target_id);
        if target.is_none() || target.unwrap().current_hp == 0 {
            return Vector2::zeros();
        }
        let target = target.unwrap();

        // Calculate interception point (predictive targeting)
        let to_target = target.position - ship.position;
        let distance = to_target.magnitude();

        // Simple pursuit for now (could add lead calculation)
        if distance > ship.range * 0.7 {
            // Move closer
            to_target.normalize() * ship.speed
        } else if distance < ship.range * 0.4 {
            // Maintain distance (kiting)
            -to_target.normalize() * ship.speed * 0.5
        } else {
            // Circle strafe (perpendicular movement)
            Vector2::new(-to_target.y, to_target.x).normalize() * ship.speed * 0.7
        }
    }

    fn update_targeting(&mut self) {
        // For each ship, find optimal target using threat assessment
        let all_ships: Vec<Ship> = self.attackers.iter()
            .chain(self.defenders.iter())
            .cloned()
            .collect();

        for ship in self.attackers.iter_mut().chain(self.defenders.iter_mut()) {
            if ship.current_hp == 0 { continue; }

            // Find enemies
            let enemies: Vec<&Ship> = all_ships.iter()
                .filter(|s| s.id != ship.id && self.are_enemies(ship.id, s.id))
                .filter(|s| s.current_hp > 0)
                .collect();

            if enemies.is_empty() {
                ship.target_id = None;
                continue;
            }

            // Target selection: prioritize low HP, close range, high threat
            let best_target = enemies.iter()
                .map(|enemy| {
                    let distance = (enemy.position - ship.position).magnitude();
                    let in_range = distance <= ship.range;

                    // Threat score (higher = better target)
                    let threat_score =
                        (1000.0 - enemy.current_hp as f32) * 0.3 + // Low HP
                        (1000.0 - distance) * 0.5 + // Close range
                        (enemy.attack as f32) * 0.2; // High threat

                    (enemy.id, threat_score, in_range)
                })
                .filter(|(_, _, in_range)| *in_range) // Only in-range targets
                .max_by(|a, b| a.1.partial_cmp(&b.1).unwrap());

            ship.target_id = best_target.map(|(id, _, _)| id);
        }
    }

    fn process_attacks(&mut self) {
        let mut damage_queue: Vec<(u64, u32)> = Vec::new(); // (target_id, damage)

        // Collect all attacks (avoid double-borrow)
        let all_ships: Vec<Ship> = self.attackers.iter()
            .chain(self.defenders.iter())
            .cloned()
            .collect();

        for ship in &all_ships {
            if ship.current_hp == 0 { continue; }
            if ship.target_id.is_none() { continue; }

            // Check fire rate cooldown
            if self.tick_count - ship.last_fired < ship.fire_rate {
                continue;
            }

            let target_id = ship.target_id.unwrap();
            let target = self.find_ship(target_id);
            if target.is_none() || target.unwrap().current_hp == 0 {
                continue;
            }
            let target = target.unwrap();

            // Check range
            let distance = (target.position - ship.position).magnitude();
            if distance > ship.range {
                continue;
            }

            // Calculate damage (with variance)
            let base_damage = ship.attack;
            let defense_mitigation = target.defense / 2;
            let variance = (rand::random::<f32>() * 0.4 + 0.8); // 80-120%

            let final_damage = ((base_damage as f32 - defense_mitigation as f32) * variance)
                .max(1.0) as u32;

            damage_queue.push((target_id, final_damage));

            // Log event
            self.combat_log.push(CombatEvent::Attack {
                tick: self.tick_count,
                attacker_id: ship.id,
                target_id,
                damage: final_damage,
            });
        }

        // Apply damage
        for (target_id, damage) in damage_queue {
            if let Some(target) = self.find_ship_mut(target_id) {
                target.current_hp = target.current_hp.saturating_sub(damage);

                if target.current_hp == 0 {
                    self.combat_log.push(CombatEvent::ShipDestroyed {
                        tick: self.tick_count,
                        ship_id: target_id,
                    });
                }
            }
        }

        // Update last_fired for ships that attacked
        for ship in self.attackers.iter_mut().chain(self.defenders.iter_mut()) {
            if damage_queue.iter().any(|(tid, _)| *tid == ship.id) {
                ship.last_fired = self.tick_count;
            }
        }
    }

    fn is_combat_over(&self) -> bool {
        let attackers_alive = self.attackers.iter().any(|s| s.current_hp > 0);
        let defenders_alive = self.defenders.iter().any(|s| s.current_hp > 0);

        !attackers_alive || !defenders_alive || self.tick_count >= self.max_ticks
    }

    // Helper methods
    fn find_ship(&self, id: u64) -> Option<&Ship> {
        self.attackers.iter().chain(self.defenders.iter())
            .find(|s| s.id == id)
    }

    fn find_ship_mut(&mut self, id: u64) -> Option<&mut Ship> {
        self.attackers.iter_mut().chain(self.defenders.iter_mut())
            .find(|s| s.id == id)
    }

    fn are_enemies(&self, id1: u64, id2: u64) -> bool {
        let id1_is_attacker = self.attackers.iter().any(|s| s.id == id1);
        let id2_is_attacker = self.attackers.iter().any(|s| s.id == id2);
        id1_is_attacker != id2_is_attacker
    }
}

#[derive(Debug, Clone)]
enum CombatEvent {
    Attack { tick: u64, attacker_id: u64, target_id: u64, damage: u32 },
    ShipDestroyed { tick: u64, ship_id: u64 },
    CombatStart { tick: u64 },
    CombatEnd { tick: u64, victor: String },
}
```

#### Anti-Cheat Measures

1. **Server-Side Authority**: All game logic on server
2. **Input Validation**: Sanitize all client commands
3. **Rate Limiting**: Prevent command flooding
4. **Replay Verification**: Combat replays must match server logs
5. **Anomaly Detection**: Flag suspicious patterns (ML-based)

```javascript
// Example: Validate fleet movement command
async function validateMoveCommand(playerId, fleetId, destination) {
    // 1. Fleet ownership
    const fleet = await db.fleets.findOne({id: fleetId, player_id: playerId});
    if (!fleet) throw new Error('Unauthorized');

    // 2. Fleet availability
    if (fleet.status !== 'idle') throw new Error('Fleet busy');

    // 3. Distance check (anti-teleport)
    const distance = calculateDistance(fleet.position, destination);
    const maxDistance = fleet.total_speed * MAX_MOVEMENT_TIME;
    if (distance > maxDistance) throw new Error('Out of range');

    // 4. Rate limiting (Redis)
    const commandKey = `cmd:${playerId}:move`;
    const commandCount = await redis.incr(commandKey);
    await redis.expire(commandKey, 60); // 1 minute window
    if (commandCount > 10) throw new Error('Rate limit exceeded');

    // All checks passed
    return true;
}
```

### 5.2 Dynamic Economy System

#### Market Mechanics

**Core Principles:**
1. **Player-Driven**: No NPC vendors, all trades between players
2. **Regional Scarcity**: Resource availability varies by galaxy sector
3. **Supply/Demand**: Prices fluctuate based on order book
4. **Transaction Fees**: Sink mechanic to prevent inflation (2% fee)

#### Economic Simulation

```javascript
// Economy Service (Node.js)
class MarketSimulator {
    async processMarketTick() {
        const resources = ['energy', 'minerals', 'rare_elements'];

        for (const resource of resources) {
            // 1. Calculate current price based on order book
            const orderBook = await this.getOrderBook(resource);
            const newPrice = this.calculateMarketPrice(orderBook);

            // 2. Match buy/sell orders
            const matches = this.matchOrders(orderBook);
            await this.executeMatches(matches);

            // 3. Update price index
            await redis.hset('market:prices', resource, newPrice);

            // 4. Broadcast price update to clients
            wsGateway.broadcastToAll('market.price_update', {
                resource,
                price: newPrice,
                volume: orderBook.totalVolume
            });
        }
    }

    calculateMarketPrice(orderBook) {
        // Weighted average of top orders
        const topBuys = orderBook.buy_orders.slice(0, 10);
        const topSells = orderBook.sell_orders.slice(0, 10);

        if (topBuys.length === 0 || topSells.length === 0) {
            return orderBook.lastPrice || 10.0; // Fallback
        }

        const avgBuyPrice = topBuys.reduce((sum, o) => sum + o.price * o.quantity, 0) /
                            topBuys.reduce((sum, o) => sum + o.quantity, 0);

        const avgSellPrice = topSells.reduce((sum, o) => sum + o.price * o.quantity, 0) /
                             topSells.reduce((sum, o) => sum + o.quantity, 0);

        // Mid-price with slight ask bias (0.2%)
        return (avgBuyPrice + avgSellPrice) / 2 * 1.002;
    }

    matchOrders(orderBook) {
        const matches = [];
        let buyIndex = 0;
        let sellIndex = 0;

        // Sort: buys high-to-low, sells low-to-high
        const sortedBuys = [...orderBook.buy_orders].sort((a, b) => b.price - a.price);
        const sortedSells = [...orderBook.sell_orders].sort((a, b) => a.price - b.price);

        while (buyIndex < sortedBuys.length && sellIndex < sortedSells.length) {
            const buyOrder = sortedBuys[buyIndex];
            const sellOrder = sortedSells[sellIndex];

            // Can match if buy price >= sell price
            if (buyOrder.price >= sellOrder.price) {
                const matchPrice = (buyOrder.price + sellOrder.price) / 2; // Fair price
                const matchQuantity = Math.min(
                    buyOrder.quantity - buyOrder.filled_quantity,
                    sellOrder.quantity - sellOrder.filled_quantity
                );

                matches.push({
                    buyer_id: buyOrder.player_id,
                    seller_id: sellOrder.player_id,
                    buy_order_id: buyOrder.id,
                    sell_order_id: sellOrder.id,
                    price: matchPrice,
                    quantity: matchQuantity,
                    resource: orderBook.resource
                });

                // Update filled quantities
                buyOrder.filled_quantity += matchQuantity;
                sellOrder.filled_quantity += matchQuantity;

                // Move to next order if fully filled
                if (buyOrder.filled_quantity >= buyOrder.quantity) buyIndex++;
                if (sellOrder.filled_quantity >= sellOrder.quantity) sellIndex++;
            } else {
                break; // No more matches possible
            }
        }

        return matches;
    }

    async executeMatches(matches) {
        for (const match of matches) {
            const totalValue = match.price * match.quantity;
            const fee = totalValue * 0.02; // 2% transaction fee
            const sellerReceives = totalValue - fee;

            // Database transaction (ACID guarantees)
            await db.transaction(async (trx) => {
                // 1. Transfer resources
                await trx('players')
                    .where('id', match.buyer_id)
                    .increment(match.resource, match.quantity)
                    .decrement('credits', totalValue);

                await trx('players')
                    .where('id', match.seller_id)
                    .decrement(match.resource, match.quantity)
                    .increment('credits', sellerReceives);

                // 2. Update order status
                await trx('market_orders')
                    .whereIn('id', [match.buy_order_id, match.sell_order_id])
                    .increment('filled_quantity', match.quantity);

                // 3. Mark completed orders
                await trx('market_orders')
                    .whereRaw('quantity <= filled_quantity')
                    .update({status: 'filled'});

                // 4. Record transaction
                await trx('market_transactions').insert({
                    buyer_id: match.buyer_id,
                    seller_id: match.seller_id,
                    resource_type: match.resource,
                    quantity: match.quantity,
                    price_per_unit: match.price,
                    total_value: totalValue,
                    transaction_time: new Date()
                });
            });

            // 5. Notify players
            await notificationService.send(match.buyer_id, {
                type: 'trade_completed',
                message: `Bought ${match.quantity} ${match.resource} for ${totalValue} credits`
            });

            await notificationService.send(match.seller_id, {
                type: 'trade_completed',
                message: `Sold ${match.quantity} ${match.resource} for ${sellerReceives} credits`
            });
        }
    }
}

// Regional resource scarcity
const REGION_MODIFIERS = {
    'core': { energy: 1.2, minerals: 0.8, rare_elements: 0.5 },
    'mid_rim': { energy: 1.0, minerals: 1.0, rare_elements: 1.0 },
    'outer_rim': { energy: 0.7, minerals: 1.3, rare_elements: 1.5 },
    'deep_space': { energy: 0.5, minerals: 1.5, rare_elements: 2.0 }
};

function calculateResourceProduction(colony) {
    const region = getRegion(colony.galaxy_x, colony.galaxy_y);
    const modifiers = REGION_MODIFIERS[region];

    return {
        energy: colony.base_energy_rate * modifiers.energy,
        minerals: colony.base_minerals_rate * modifiers.minerals,
        rare_elements: colony.base_rare_elements_rate * modifiers.rare_elements
    };
}
```

#### Economic Analytics Dashboard

```javascript
// Real-time market statistics
GET /api/market/statistics

Response:
{
    "minerals": {
        "current_price": 15.47,
        "24h_change": "+2.3%",
        "volume_24h": 1500000,
        "buy_orders": 342,
        "sell_orders": 278,
        "price_history": [
            { "time": "2025-11-03T10:00:00Z", "price": 15.12 },
            { "time": "2025-11-03T11:00:00Z", "price": 15.28 },
            // ... hourly snapshots
        ]
    },
    // ... other resources
}
```

### 5.3 Advanced Alliance System

#### Governance Models

```javascript
const GOVERNANCE_TYPES = {
    AUTOCRACY: {
        name: 'Autocracy',
        description: 'Leader makes all decisions',
        voting_required: false,
        leadership: 'single'
    },
    DEMOCRACY: {
        name: 'Democracy',
        description: 'All members vote on major decisions',
        voting_required: true,
        voting_threshold: 0.51, // 51% majority
        leadership: 'elected'
    },
    COUNCIL: {
        name: 'Council',
        description: 'Officers vote on decisions',
        voting_required: true,
        voting_threshold: 0.66, // 66% supermajority
        leadership: 'council',
        voter_roles: ['leader', 'commander', 'diplomat']
    },
    MERITOCRACY: {
        name: 'Meritocracy',
        description: 'Vote weight based on contribution',
        voting_required: true,
        voting_threshold: 0.51,
        weighted_voting: true, // Vote power = contribution points
        leadership: 'elected'
    }
};
```

#### Voting System Implementation

```javascript
class AllianceVotingService {
    async createVote(allianceId, creatorId, voteData) {
        const alliance = await db.alliances.findOne({id: allianceId});
        const governance = GOVERNANCE_TYPES[alliance.governance_type];

        // Check if voting is required
        if (!governance.voting_required) {
            throw new Error('Voting not enabled for this governance type');
        }

        // Calculate required votes
        let totalEligibleVotes;
        if (governance.weighted_voting) {
            // Sum of contribution points
            totalEligibleVotes = await db.alliance_members
                .where('alliance_id', allianceId)
                .sum('contribution_points');
        } else if (governance.voter_roles) {
            // Only specific roles can vote
            totalEligibleVotes = await db.alliance_members
                .where('alliance_id', allianceId)
                .whereIn('role', governance.voter_roles)
                .count();
        } else {
            // All members vote
            totalEligibleVotes = alliance.member_count;
        }

        const requiredVotes = Math.ceil(totalEligibleVotes * governance.voting_threshold);

        // Create vote
        const vote = await db.alliance_votes.insert({
            alliance_id: allianceId,
            created_by: creatorId,
            vote_type: voteData.type,
            description: voteData.description,
            target_data: JSON.stringify(voteData.targetData),
            required_votes: requiredVotes,
            status: 'active',
            expires_at: new Date(Date.now() + 48 * 3600 * 1000), // 48 hours
            created_at: new Date()
        });

        // Notify eligible voters
        await this.notifyEligibleVoters(allianceId, vote.id, governance);

        return vote;
    }

    async castVote(voteId, playerId, choice) {
        const vote = await db.alliance_votes.findOne({id: voteId});
        if (vote.status !== 'active') {
            throw new Error('Vote is no longer active');
        }

        // Check if player is eligible
        const member = await db.alliance_members.findOne({
            alliance_id: vote.alliance_id,
            player_id: playerId
        });
        if (!member) throw new Error('Not an alliance member');

        const alliance = await db.alliances.findOne({id: vote.alliance_id});
        const governance = GOVERNANCE_TYPES[alliance.governance_type];

        // Check role restrictions
        if (governance.voter_roles && !governance.voter_roles.includes(member.role)) {
            throw new Error('Your role cannot vote on this');
        }

        // Calculate vote weight
        let voteWeight = 1;
        if (governance.weighted_voting) {
            voteWeight = member.contribution_points || 1;
        }

        // Record vote
        await db.vote_records.insert({
            vote_id: voteId,
            player_id: playerId,
            vote_choice: choice,
            voted_at: new Date()
        });

        // Update vote tallies
        if (choice) {
            await db.alliance_votes
                .where('id', voteId)
                .increment('votes_for', voteWeight);
        } else {
            await db.alliance_votes
                .where('id', voteId)
                .increment('votes_against', voteWeight);
        }

        // Check if vote passes
        await this.checkVoteCompletion(voteId);
    }

    async checkVoteCompletion(voteId) {
        const vote = await db.alliance_votes.findOne({id: voteId});

        if (vote.votes_for >= vote.required_votes) {
            // Vote passed!
            await db.alliance_votes
                .where('id', voteId)
                .update({status: 'passed'});

            // Execute vote action
            await this.executeVoteAction(vote);
        } else if (vote.votes_against > (vote.required_votes * 2 - vote.required_votes)) {
            // Vote failed (impossible to pass)
            await db.alliance_votes
                .where('id', voteId)
                .update({status: 'failed'});
        }
    }

    async executeVoteAction(vote) {
        const targetData = JSON.parse(vote.target_data);

        switch (vote.vote_type) {
            case 'declare_war':
                await this.declareWar(vote.alliance_id, targetData.target_alliance_id);
                break;

            case 'accept_member':
                await this.acceptMember(vote.alliance_id, targetData.player_id);
                break;

            case 'promote_member':
                await this.promoteMember(
                    vote.alliance_id,
                    targetData.player_id,
                    targetData.new_role
                );
                break;

            case 'change_governance':
                await db.alliances
                    .where('id', vote.alliance_id)
                    .update({
                        governance_type: targetData.new_governance_type,
                        voting_threshold: GOVERNANCE_TYPES[targetData.new_governance_type].voting_threshold
                    });
                break;

            case 'set_tax_rate':
                await db.alliances
                    .where('id', vote.alliance_id)
                    .update({tax_rate: targetData.tax_rate});
                break;
        }

        // Notify all members
        await this.notifyAllianceMembers(vote.alliance_id, {
            type: 'vote_passed',
            message: `Vote "${vote.description}" has passed and been executed`
        });
    }
}
```

#### Alliance Treasury & Taxation

```javascript
// Automatic tax collection (runs each tick)
async function collectAllianceTaxes() {
    const alliances = await db.alliances.findAll({
        where: { tax_rate: { $gt: 0 } }
    });

    for (const alliance of alliances) {
        const members = await db.alliance_members.findAll({
            where: { alliance_id: alliance.id }
        });

        for (const member of members) {
            const player = await db.players.findOne({id: member.player_id});

            // Calculate tax on resource production
            const taxAmount = {
                energy: Math.floor(player.energy_rate * alliance.tax_rate),
                minerals: Math.floor(player.minerals_rate * alliance.tax_rate),
                rare_elements: Math.floor(player.rare_elements_rate * alliance.tax_rate)
            };

            // Deduct from player, add to alliance treasury
            await db.players
                .where('id', player.id)
                .decrement({
                    energy: taxAmount.energy,
                    minerals: taxAmount.minerals,
                    rare_elements: taxAmount.rare_elements
                });

            await db.alliance_treasury
                .where('alliance_id', alliance.id)
                .increment({
                    energy: taxAmount.energy,
                    minerals: taxAmount.minerals,
                    rare_elements: taxAmount.rare_elements
                });
        }
    }
}
```

### 5.4 Massive World Map Rendering

#### Challenge: Display 10,000+ entities without performance loss

**Solution: Multi-tier optimization strategy**

#### Tier 1: Spatial Partitioning (Server-Side)

```javascript
// Divide galaxy into sectors (100x100 units each)
const SECTOR_SIZE = 100;

function getSectorId(x, y) {
    const sectorX = Math.floor(x / SECTOR_SIZE);
    const sectorY = Math.floor(y / SECTOR_SIZE);
    return `${sectorX},${sectorY}`;
}

// Client only receives data for visible sectors + 1 buffer zone
async function getVisibleEntities(viewportCenter, viewportSize) {
    const minX = viewportCenter.x - viewportSize.width / 2 - SECTOR_SIZE;
    const maxX = viewportCenter.x + viewportSize.width / 2 + SECTOR_SIZE;
    const minY = viewportCenter.y - viewportSize.height / 2 - SECTOR_SIZE;
    const maxY = viewportCenter.y + viewportSize.height / 2 + SECTOR_SIZE;

    const visibleSectors = [];
    for (let x = Math.floor(minX / SECTOR_SIZE); x <= Math.ceil(maxX / SECTOR_SIZE); x++) {
        for (let y = Math.floor(minY / SECTOR_SIZE); y <= Math.ceil(maxY / SECTOR_SIZE); y++) {
            visibleSectors.push(`${x},${y}`);
        }
    }

    // Query entities in these sectors only
    const entities = await db.query(`
        SELECT id, type, galaxy_x, galaxy_y, owner_id, status
        FROM (
            SELECT id, 'colony' as type, galaxy_x, galaxy_y, player_id as owner_id, 'active' as status,
                   FLOOR(galaxy_x / ${SECTOR_SIZE}) || ',' || FLOOR(galaxy_y / ${SECTOR_SIZE}) as sector_id
            FROM colonies
            UNION ALL
            SELECT id, 'fleet' as type, current_x, current_y, player_id, status,
                   FLOOR(current_x / ${SECTOR_SIZE}) || ',' || FLOOR(current_y / ${SECTOR_SIZE}) as sector_id
            FROM fleets
        ) entities
        WHERE sector_id = ANY($1)
    `, [visibleSectors]);

    return entities;
}
```

#### Tier 2: Level of Detail (LOD) System

```javascript
// Babylon.js LOD implementation
class GalaxyMapRenderer {
    constructor(scene) {
        this.scene = scene;
        this.entityMeshes = new Map(); // id -> mesh
        this.camera = scene.activeCamera;
    }

    createEntityMesh(entity) {
        const position = new BABYLON.Vector3(entity.galaxy_x, 0, entity.galaxy_y);

        // Create LOD levels
        const highDetailMesh = this.createHighDetailModel(entity);
        const mediumDetailMesh = this.createMediumDetailModel(entity);
        const lowDetailMesh = this.createLowDetailModel(entity);
        const iconMesh = this.createIconModel(entity);

        // Set LOD distances
        highDetailMesh.position = position;
        highDetailMesh.addLODLevel(50, mediumDetailMesh);   // Switch at 50 units
        highDetailMesh.addLODLevel(200, lowDetailMesh);     // Switch at 200 units
        highDetailMesh.addLODLevel(1000, iconMesh);         // Switch at 1000 units
        highDetailMesh.addLODLevel(5000, null);             // Cull beyond 5000 units

        this.entityMeshes.set(entity.id, highDetailMesh);
        return highDetailMesh;
    }

    createHighDetailModel(entity) {
        // Full 3D model with textures, animations
        if (entity.type === 'colony') {
            return BABYLON.MeshBuilder.CreateSphere('colony_' + entity.id, {
                diameter: 5,
                segments: 16
            }, this.scene);
        } else {
            return BABYLON.MeshBuilder.CreateBox('fleet_' + entity.id, {
                width: 2,
                height: 1,
                depth: 3
            }, this.scene);
        }
    }

    createMediumDetailModel(entity) {
        // Simplified geometry, basic textures
        return BABYLON.MeshBuilder.CreateSphere('colony_med_' + entity.id, {
            diameter: 5,
            segments: 8 // Half the segments
        }, this.scene);
    }

    createLowDetailModel(entity) {
        // Very simple shapes
        return BABYLON.MeshBuilder.CreateBox('entity_low_' + entity.id, {
            size: 3
        }, this.scene);
    }

    createIconModel(entity) {
        // 2D sprite/billboard
        const plane = BABYLON.MeshBuilder.CreatePlane('icon_' + entity.id, {
            size: 5
        }, this.scene);

        // Always face camera
        plane.billboardMode = BABYLON.Mesh.BILLBOARDMODE_ALL;

        return plane;
    }
}
```

#### Tier 3: Instanced Rendering for Common Objects

```javascript
// Use instanced meshes for repeated objects (huge performance gain)
class InstancedEntityRenderer {
    constructor(scene) {
        this.scene = scene;
        this.instances = new Map(); // entity_type -> [instances]

        // Create source meshes (rendered once, reused many times)
        this.sourceMeshes = {
            colony: this.createColonyMesh(),
            fleet_fighter: this.createFighterMesh(),
            fleet_cruiser: this.createCruiserMesh()
        };
    }

    addEntity(entity) {
        const sourceKey = this.getSourceMeshKey(entity);
        const sourceMesh = this.sourceMeshes[sourceKey];

        // Create instance (very cheap - just a transform matrix)
        const instance = sourceMesh.createInstance(entity.id);
        instance.position = new BABYLON.Vector3(entity.galaxy_x, 0, entity.galaxy_y);

        if (!this.instances.has(sourceKey)) {
            this.instances.set(sourceKey, []);
        }
        this.instances.get(sourceKey).push({
            id: entity.id,
            mesh: instance
        });
    }

    // Instancing can render 10k+ objects at 60fps vs 100-200 without
}
```

#### Tier 4: Occlusion Culling & Frustum Culling

```javascript
// Babylon.js automatically does frustum culling, but we can optimize further
scene.onBeforeRenderObservable.add(() => {
    const frustum = BABYLON.Frustum.GetPlanes(camera.getTransformationMatrix());

    entityMeshes.forEach((mesh, id) => {
        // Check if mesh bounding box intersects frustum
        const inFrustum = mesh.isInFrustum(frustum);
        mesh.isVisible = inFrustum;
    });
});
```

#### Tier 5: Delta Updates (Network Optimization)

```javascript
// Only send changed data to client
class MapStateManager {
    constructor() {
        this.clientStates = new Map(); // socketId -> lastSentState
    }

    sendMapUpdate(socket, currentMapState) {
        const lastState = this.clientStates.get(socket.id) || { entities: {} };
        const delta = this.calculateDelta(lastState, currentMapState);

        // Only send changes
        socket.emit('map.delta', {
            added: delta.added,       // New entities
            removed: delta.removed,   // Destroyed entities
            moved: delta.moved,       // Position changes
            updated: delta.updated    // Other property changes
        });

        this.clientStates.set(socket.id, currentMapState);
    }

    calculateDelta(oldState, newState) {
        const delta = { added: [], removed: [], moved: [], updated: [] };

        // Find added entities
        for (const [id, entity] of Object.entries(newState.entities)) {
            if (!oldState.entities[id]) {
                delta.added.push(entity);
            }
        }

        // Find removed entities
        for (const id of Object.keys(oldState.entities)) {
            if (!newState.entities[id]) {
                delta.removed.push(id);
            }
        }

        // Find moved/updated entities
        for (const [id, newEntity] of Object.entries(newState.entities)) {
            const oldEntity = oldState.entities[id];
            if (oldEntity) {
                if (oldEntity.x !== newEntity.x || oldEntity.y !== newEntity.y) {
                    delta.moved.push({
                        id,
                        x: newEntity.x,
                        y: newEntity.y
                    });
                } else if (JSON.stringify(oldEntity) !== JSON.stringify(newEntity)) {
                    delta.updated.push(newEntity);
                }
            }
        }

        return delta;
    }
}
```

#### Tier 6: Client-Side Prediction & Interpolation

```javascript
// Smooth movement despite network latency
class EntityInterpolator {
    constructor(entity) {
        this.entity = entity;
        this.targetPosition = entity.position.clone();
        this.interpolationSpeed = 0.1; // 10% per frame
    }

    update(deltaTime) {
        // Smoothly move towards target position
        this.entity.position.x += (this.targetPosition.x - this.entity.position.x) * this.interpolationSpeed;
        this.entity.position.z += (this.targetPosition.z - this.entity.position.z) * this.interpolationSpeed;
    }

    setTargetPosition(newPosition) {
        this.targetPosition = newPosition;
    }
}

// When server sends position update
socket.on('entity.moved', (data) => {
    const entity = entityManager.get(data.id);
    const interpolator = interpolators.get(data.id);

    if (interpolator) {
        interpolator.setTargetPosition(new BABYLON.Vector3(data.x, 0, data.y));
    } else {
        // Instant update if no interpolator
        entity.position.x = data.x;
        entity.position.z = data.y;
    }
});
```

#### Performance Benchmarks (Target)

| Entities Visible | FPS Target | Techniques Used |
|------------------|------------|-----------------|
| 0-100 | 60 | Full detail, no optimization needed |
| 100-500 | 60 | LOD + Frustum culling |
| 500-2000 | 60 | LOD + Instancing + Frustum culling |
| 2000-10000 | 60 | All techniques + Occlusion culling |
| 10000+ | 30-60 | All techniques + Aggressive LOD |

---

## 6. Server-Client Synchronization Model

### Core Principles

1. **Server Authority**: Server is source of truth
2. **Client Prediction**: Client predicts outcomes for immediate feedback
3. **Server Reconciliation**: Server corrects client when predictions are wrong
4. **Delta Compression**: Only send changes, not full state

### Synchronization Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    CLIENT SIDE                          │
│                                                         │
│  ┌─────────────┐         ┌──────────────┐            │
│  │   Input     │         │   Render     │            │
│  │   Handler   │────────▶│   Engine     │            │
│  └──────┬──────┘         └──────────────┘            │
│         │                                              │
│         │ 1. Send command                             │
│         │ 2. Predict outcome                          │
│         │ 3. Apply prediction                         │
│         ▼                                              │
│  ┌──────────────────────────────────┐                │
│  │    Local State Manager           │                │
│  │  - Predicted state               │                │
│  │  - Pending commands queue        │                │
│  │  - Last acknowledged sequence #  │                │
│  └──────────┬───────────────────────┘                │
└─────────────┼────────────────────────────────────────┘
              │
              │ WebSocket (Binary Protocol)
              │
┌─────────────┼────────────────────────────────────────┐
│             ▼          SERVER SIDE                    │
│  ┌──────────────────────────────────┐                │
│  │    Command Processor              │                │
│  │  1. Validate command              │                │
│  │  2. Execute on authoritative state│                │
│  │  3. Broadcast to affected clients │                │
│  └──────────┬───────────────────────┘                │
│             │                                          │
│             ▼                                          │
│  ┌──────────────────────────────────┐                │
│  │   Authoritative State             │                │
│  │   (Single Source of Truth)        │                │
│  └──────────┬───────────────────────┘                │
│             │                                          │
│             ▼                                          │
│  ┌──────────────────────────────────┐                │
│  │   Delta Generator                 │                │
│  │   (Compute minimal changes)       │                │
│  └──────────┬───────────────────────┘                │
└─────────────┼────────────────────────────────────────┘
              │
              ▼
       Broadcast to Clients
```

### Message Protocol (MessagePack Binary Format)

```javascript
// Message structure
{
    type: u8,           // Message type ID (0-255)
    sequence: u32,      // Sequence number for ordering
    timestamp: u64,     // Server timestamp
    payload: bytes      // Type-specific payload
}

// Message types
const MESSAGE_TYPES = {
    // Client -> Server
    CMD_MOVE_FLEET: 0x01,
    CMD_ATTACK: 0x02,
    CMD_BUILD: 0x03,
    CMD_TRADE: 0x04,

    // Server -> Client
    STATE_DELTA: 0x10,
    ENTITY_CREATED: 0x11,
    ENTITY_DESTROYED: 0x12,
    ENTITY_MOVED: 0x13,
    COMBAT_EVENT: 0x14,

    // Bidirectional
    PING: 0x20,
    PONG: 0x21,
    ACK: 0x22
};

// Example: Move fleet command
// Client -> Server
{
    type: MESSAGE_TYPES.CMD_MOVE_FLEET,
    sequence: 1234,
    timestamp: Date.now(),
    payload: msgpack.encode({
        fleet_id: 456,
        destination: { x: 1500, y: 2300 },
        predicted_arrival: Date.now() + 60000 // Client's prediction
    })
}

// Server -> Client (acknowledgment)
{
    type: MESSAGE_TYPES.ACK,
    sequence: 1234, // Same sequence
    timestamp: serverTime,
    payload: msgpack.encode({
        success: true,
        corrected_arrival: serverTime + 62000, // Server's calculation
        actual_position: { x: 100, y: 200 } // If client position was wrong
    })
}
```

### Latency Compensation Techniques

#### 1. Client-Side Prediction

```javascript
// Client predicts move outcome immediately
function handleMoveFleetCommand(fleetId, destination) {
    const fleet = localState.getFleet(fleetId);

    // Generate sequence number
    const sequence = localState.nextSequence++;

    // Predict outcome
    const travelTime = calculateTravelTime(fleet.position, destination, fleet.speed);
    const predictedArrival = Date.now() + travelTime;

    // Apply prediction locally (instant feedback)
    fleet.status = 'moving';
    fleet.destination = destination;
    fleet.eta = predictedArrival;

    // Start visual movement animation
    animateFleetMovement(fleet, destination, travelTime);

    // Add to pending commands queue
    localState.pendingCommands.set(sequence, {
        type: 'move_fleet',
        fleet_id: fleetId,
        destination,
        predicted_state: { ...fleet }
    });

    // Send to server
    socket.send(msgpack.encode({
        type: MESSAGE_TYPES.CMD_MOVE_FLEET,
        sequence,
        timestamp: Date.now(),
        payload: { fleet_id: fleetId, destination }
    }));
}
```

#### 2. Server Reconciliation

```javascript
// Server responds with authoritative result
socket.on('message', (data) => {
    const message = msgpack.decode(data);

    if (message.type === MESSAGE_TYPES.ACK) {
        const pendingCmd = localState.pendingCommands.get(message.sequence);

        if (!pendingCmd) return; // Already processed

        const serverPayload = msgpack.decode(message.payload);

        if (serverPayload.success) {
            // Prediction was correct (or close enough)
            if (serverPayload.corrected_arrival) {
                // Minor correction
                const fleet = localState.getFleet(pendingCmd.fleet_id);
                fleet.eta = serverPayload.corrected_arrival;

                // Adjust animation speed
                adjustAnimationSpeed(fleet, serverPayload.corrected_arrival - Date.now());
            }

            // Remove from pending
            localState.pendingCommands.delete(message.sequence);
        } else {
            // Prediction was wrong - revert and reapply
            console.warn('Server rejected command:', serverPayload.error);

            const fleet = localState.getFleet(pendingCmd.fleet_id);

            // Revert to pre-prediction state
            Object.assign(fleet, serverPayload.actual_state);

            // Replay all subsequent pending commands
            const laterCommands = Array.from(localState.pendingCommands.values())
                .filter(cmd => cmd.sequence > message.sequence);

            for (const cmd of laterCommands) {
                replayCommand(cmd);
            }

            localState.pendingCommands.delete(message.sequence);
        }
    }
});
```

#### 3. Entity Interpolation (Handle Late Updates)

```javascript
// Smoothly interpolate between known states
class EntityStateInterpolator {
    constructor(entity) {
        this.entity = entity;
        this.stateBuffer = []; // Circular buffer of past states
        this.renderDelay = 100; // Render 100ms in the past
    }

    addState(state) {
        this.stateBuffer.push({
            timestamp: Date.now(),
            position: state.position,
            rotation: state.rotation,
            // ... other interpolable properties
        });

        // Keep buffer size manageable
        if (this.stateBuffer.length > 20) {
            this.stateBuffer.shift();
        }
    }

    getInterpolatedState(renderTime) {
        const targetTime = renderTime - this.renderDelay;

        // Find two states to interpolate between
        let before = null, after = null;

        for (let i = 0; i < this.stateBuffer.length - 1; i++) {
            if (this.stateBuffer[i].timestamp <= targetTime &&
                this.stateBuffer[i + 1].timestamp >= targetTime) {
                before = this.stateBuffer[i];
                after = this.stateBuffer[i + 1];
                break;
            }
        }

        if (!before || !after) {
            // Fallback to latest state
            return this.stateBuffer[this.stateBuffer.length - 1];
        }

        // Linear interpolation
        const timeDiff = after.timestamp - before.timestamp;
        const t = (targetTime - before.timestamp) / timeDiff;

        return {
            position: {
                x: before.position.x + (after.position.x - before.position.x) * t,
                y: before.position.y + (after.position.y - before.position.y) * t
            },
            rotation: before.rotation + (after.rotation - before.rotation) * t
        };
    }
}
```

### Handling Packet Loss

```javascript
// Reliable messaging layer over WebSocket
class ReliableWebSocket {
    constructor(url) {
        this.socket = new WebSocket(url);
        this.pendingAcks = new Map(); // sequence -> {message, timestamp, retries}
        this.ackTimeout = 3000; // 3 seconds
        this.maxRetries = 3;

        this.setupHandlers();
        this.startAckChecker();
    }

    send(message) {
        const sequence = message.sequence;

        // Store for potential retry
        this.pendingAcks.set(sequence, {
            message,
            timestamp: Date.now(),
            retries: 0
        });

        // Send message
        this.socket.send(msgpack.encode(message));
    }

    setupHandlers() {
        this.socket.on('message', (data) => {
            const message = msgpack.decode(data);

            if (message.type === MESSAGE_TYPES.ACK) {
                // Remove from pending
                this.pendingAcks.delete(message.sequence);
            }

            // ... handle other message types
        });
    }

    startAckChecker() {
        setInterval(() => {
            const now = Date.now();

            for (const [sequence, pending] of this.pendingAcks.entries()) {
                if (now - pending.timestamp > this.ackTimeout) {
                    if (pending.retries < this.maxRetries) {
                        // Retry
                        console.warn(`Retrying message ${sequence} (attempt ${pending.retries + 1})`);
                        this.socket.send(msgpack.encode(pending.message));
                        pending.retries++;
                        pending.timestamp = now;
                    } else {
                        // Give up
                        console.error(`Message ${sequence} failed after ${this.maxRetries} retries`);
                        this.pendingAcks.delete(sequence);

                        // Notify application layer
                        this.onMessageFailed(pending.message);
                    }
                }
            }
        }, 1000);
    }
}
```

### Network Optimization: Delta Compression

```javascript
// Compress state updates using delta encoding
class StateDeltaEncoder {
    constructor() {
        this.lastSentState = {};
    }

    encode(currentState) {
        const delta = {
            changed: {},
            removed: []
        };

        // Find changed properties
        for (const [key, value] of Object.entries(currentState)) {
            if (JSON.stringify(value) !== JSON.stringify(this.lastSentState[key])) {
                delta.changed[key] = value;
            }
        }

        // Find removed properties
        for (const key of Object.keys(this.lastSentState)) {
            if (!(key in currentState)) {
                delta.removed.push(key);
            }
        }

        // Update last sent state
        this.lastSentState = JSON.parse(JSON.stringify(currentState));

        return delta;
    }
}

// Example usage
const encoder = new StateDeltaEncoder();

// First update: send full state
const state1 = {
    energy: 1000,
    minerals: 500,
    rare_elements: 50
};
console.log(encoder.encode(state1));
// Output: { changed: { energy: 1000, minerals: 500, rare_elements: 50 }, removed: [] }

// Second update: only energy changed
const state2 = {
    energy: 1050, // +50
    minerals: 500,
    rare_elements: 50
};
console.log(encoder.encode(state2));
// Output: { changed: { energy: 1050 }, removed: [] }
// Saves ~60% bandwidth!
```

---

## 7. Security & Anti-Cheat

### Multi-Layer Security Approach

#### Layer 1: Input Validation

```javascript
// Validate all client inputs server-side
const commandSchema = {
    CMD_MOVE_FLEET: {
        fleet_id: { type: 'integer', min: 1 },
        destination: {
            type: 'object',
            properties: {
                x: { type: 'number', min: 0, max: 10000 },
                y: { type: 'number', min: 0, max: 10000 }
            }
        }
    },
    CMD_ATTACK: {
        fleet_id: { type: 'integer', min: 1 },
        target_id: { type: 'integer', min: 1 },
        target_type: { type: 'string', enum: ['colony', 'fleet'] }
    }
};

function validateCommand(commandType, payload) {
    const schema = commandSchema[commandType];
    // Use Joi or similar validation library
    const { error } = joi.validate(payload, schema);
    if (error) {
        throw new ValidationError(error.message);
    }
}
```

#### Layer 2: Rate Limiting

```javascript
// Prevent command flooding
class RateLimiter {
    constructor(redis) {
        this.redis = redis;
        this.limits = {
            CMD_MOVE_FLEET: { window: 60, max: 20 },  // 20 per minute
            CMD_ATTACK: { window: 60, max: 10 },      // 10 per minute
            CMD_TRADE: { window: 60, max: 30 }        // 30 per minute
        };
    }

    async checkLimit(playerId, commandType) {
        const limit = this.limits[commandType];
        if (!limit) return true; // No limit set

        const key = `ratelimit:${playerId}:${commandType}`;
        const count = await this.redis.incr(key);

        if (count === 1) {
            await this.redis.expire(key, limit.window);
        }

        if (count > limit.max) {
            throw new RateLimitError(`Too many ${commandType} commands`);
        }

        return true;
    }
}
```

#### Layer 3: State Verification

```javascript
// Verify game state consistency
async function verifyFleetMovement(fleetId, destination, playerId) {
    // 1. Fleet exists and belongs to player
    const fleet = await db.fleets.findOne({ id: fleetId, player_id: playerId });
    if (!fleet) throw new AuthorizationError('Fleet not found or unauthorized');

    // 2. Fleet is not busy
    if (fleet.status !== 'idle') {
        throw new GameLogicError('Fleet is busy');
    }

    // 3. Distance is possible given fleet speed
    const distance = Math.sqrt(
        Math.pow(destination.x - fleet.current_x, 2) +
        Math.pow(destination.y - fleet.current_y, 2)
    );
    const maxDistance = fleet.total_speed * 3600; // Max 1 hour travel
    if (distance > maxDistance) {
        throw new GameLogicError('Destination too far');
    }

    // 4. Destination is valid (not in protected zone, etc.)
    const sector = getSector(destination.x, destination.y);
    if (sector.type === 'protected') {
        throw new GameLogicError('Cannot move to protected sector');
    }

    return true;
}
```

#### Layer 4: Anomaly Detection

```javascript
// Machine learning-based cheat detection
class CheatDetector {
    constructor() {
        this.playerProfiles = new Map(); // playerId -> profile
    }

    async analyzePlayerBehavior(playerId, action) {
        let profile = this.playerProfiles.get(playerId);
        if (!profile) {
            profile = {
                actions: [],
                flags: 0,
                lastFlagTime: null
            };
            this.playerProfiles.set(playerId, profile);
        }

        profile.actions.push({
            type: action.type,
            timestamp: Date.now(),
            data: action.data
        });

        // Keep only recent actions (last hour)
        const oneHourAgo = Date.now() - 3600000;
        profile.actions = profile.actions.filter(a => a.timestamp > oneHourAgo);

        // Detect anomalies
        const anomalies = [
            this.detectImpossibleSpeed(profile),
            this.detectResourceAnomaly(profile),
            this.detectBotBehavior(profile)
        ];

        if (anomalies.some(a => a.detected)) {
            profile.flags++;
            profile.lastFlagTime = Date.now();

            // Alert moderators
            await this.alertModerators(playerId, anomalies);

            // Auto-ban if too many flags
            if (profile.flags > 5) {
                await this.banPlayer(playerId, 'Automated cheat detection');
            }
        }
    }

    detectImpossibleSpeed(profile) {
        // Check for instant travel
        const moveActions = profile.actions.filter(a => a.type === 'move_fleet');

        for (let i = 1; i < moveActions.length; i++) {
            const prev = moveActions[i - 1];
            const curr = moveActions[i];

            const timeDiff = (curr.timestamp - prev.timestamp) / 1000; // seconds
            const distance = Math.sqrt(
                Math.pow(curr.data.x - prev.data.x, 2) +
                Math.pow(curr.data.y - prev.data.y, 2)
            );

            const minTime = distance / 1000; // Assume max speed of 1000 units/sec

            if (timeDiff < minTime * 0.9) {
                return {
                    detected: true,
                    type: 'impossible_speed',
                    confidence: 0.95
                };
            }
        }

        return { detected: false };
    }

    detectResourceAnomaly(profile) {
        // Check for sudden resource spikes
        // (Compare against expected production rates)
        // Implementation omitted for brevity
        return { detected: false };
    }

    detectBotBehavior(profile) {
        // Check for inhuman precision and timing
        const actionIntervals = [];
        for (let i = 1; i < profile.actions.length; i++) {
            actionIntervals.push(profile.actions[i].timestamp - profile.actions[i - 1].timestamp);
        }

        if (actionIntervals.length < 10) return { detected: false };

        // Calculate standard deviation
        const mean = actionIntervals.reduce((a, b) => a + b) / actionIntervals.length;
        const variance = actionIntervals.reduce((sum, val) => sum + Math.pow(val - mean, 2), 0) / actionIntervals.length;
        const stdDev = Math.sqrt(variance);

        // Humans have irregular timing; bots are too consistent
        if (stdDev < 50 && mean < 1000) { // Less than 50ms variance
            return {
                detected: true,
                type: 'bot_behavior',
                confidence: 0.7
            };
        }

        return { detected: false };
    }
}
```

#### Layer 5: Encryption & Authentication

```javascript
// JWT-based authentication
const jwt = require('jsonwebtoken');

function generateToken(userId) {
    return jwt.sign(
        { user_id: userId, issued_at: Date.now() },
        process.env.JWT_SECRET,
        { expiresIn: '24h' }
    );
}

function verifyToken(token) {
    try {
        const decoded = jwt.verify(token, process.env.JWT_SECRET);
        return decoded.user_id;
    } catch (err) {
        throw new AuthenticationError('Invalid token');
    }
}

// WebSocket authentication
io.use((socket, next) => {
    const token = socket.handshake.auth.token;
    try {
        const userId = verifyToken(token);
        socket.userId = userId;
        next();
    } catch (err) {
        next(new Error('Authentication failed'));
    }
});
```

---

## 8. Scalability Considerations

### Horizontal Scaling Strategy

#### Database Scaling

**PostgreSQL (Read Replicas)**
```
┌─────────────┐
│   Master    │────────┐
│  (Writes)   │        │
└─────────────┘        │ Replication
                       │
       ┌───────────────┴───────────────┐
       │                               │
┌──────▼──────┐               ┌────────▼──────┐
│  Replica 1  │               │   Replica 2   │
│   (Reads)   │               │    (Reads)    │
└─────────────┘               └───────────────┘
```

**Sharding Strategy (if needed)**
```javascript
// Shard by player ID (consistent hashing)
function getShardForPlayer(playerId) {
    const shardCount = 4;
    return playerId % shardCount;
}

// Or shard by galaxy sector for better locality
function getShardForSector(sectorId) {
    const shardMap = {
        'core': 0,
        'mid_rim': 1,
        'outer_rim': 2,
        'deep_space': 3
    };
    const region = getSectorRegion(sectorId);
    return shardMap[region];
}
```

#### Service Scaling

**Kubernetes Deployment**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: game-world-service
spec:
  replicas: 5  # Start with 5 instances
  selector:
    matchLabels:
      app: game-world
  template:
    metadata:
      labels:
        app: game-world
    spec:
      containers:
      - name: game-world
        image: game-world-service:latest
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
        env:
        - name: REDIS_URL
          value: "redis://redis-cluster:6379"
        - name: DB_URL
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: connection-string
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: game-world-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: game-world-service
  minReplicas: 5
  maxReplicas: 50
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

#### Load Balancing

```
                    ┌────────────────┐
                    │   CloudFlare   │
                    │   (CDN + DDoS) │
                    └────────┬───────┘
                             │
                    ┌────────▼────────┐
                    │  Load Balancer  │
                    │   (NGINX/HAProxy)│
                    └────────┬────────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
     ┌────▼────┐        ┌────▼────┐       ┌────▼────┐
     │  API    │        │  API    │       │  API    │
     │Gateway 1│        │Gateway 2│       │Gateway 3│
     └─────────┘        └─────────┘       └─────────┘
```

#### WebSocket Connection Distribution

```javascript
// Sticky sessions with Redis pub/sub for cross-server communication
class DistributedWebSocketManager {
    constructor(serverId) {
        this.serverId = serverId;
        this.localConnections = new Map(); // socketId -> socket
        this.redis = new Redis();

        // Subscribe to broadcasts
        this.redis.subscribe(`ws:broadcast:all`);
        this.redis.subscribe(`ws:broadcast:${serverId}`);

        this.redis.on('message', (channel, message) => {
            this.handleBroadcast(JSON.parse(message));
        });
    }

    addConnection(socket, userId) {
        this.localConnections.set(socket.id, socket);

        // Store server location for this user
        this.redis.hset('ws:user_servers', userId, this.serverId);
    }

    async sendToUser(userId, message) {
        // Check if user is on this server
        const userServerId = await this.redis.hget('ws:user_servers', userId);

        if (userServerId === this.serverId) {
            // Local delivery
            const socket = Array.from(this.localConnections.values())
                .find(s => s.userId === userId);
            if (socket) socket.emit('message', message);
        } else {
            // Remote delivery via Redis pub/sub
            this.redis.publish(`ws:broadcast:${userServerId}`, JSON.stringify({
                type: 'send_to_user',
                user_id: userId,
                message
            }));
        }
    }

    broadcastToAll(message) {
        this.redis.publish('ws:broadcast:all', JSON.stringify({
            type: 'broadcast',
            message
        }));
    }

    handleBroadcast(data) {
        if (data.type === 'send_to_user') {
            // Deliver to local user
            const socket = Array.from(this.localConnections.values())
                .find(s => s.userId === data.user_id);
            if (socket) socket.emit('message', data.message);
        } else if (data.type === 'broadcast') {
            // Broadcast to all local connections
            this.localConnections.forEach(socket => {
                socket.emit('message', data.message);
            });
        }
    }
}
```

#### Caching Strategy

```javascript
// Multi-tier caching
class CacheManager {
    constructor() {
        this.l1Cache = new Map(); // In-memory (per instance)
        this.l2Cache = new Redis(); // Shared Redis
        this.l3Cache = 'database'; // PostgreSQL
    }

    async get(key) {
        // L1: In-memory cache (fastest)
        if (this.l1Cache.has(key)) {
            return this.l1Cache.get(key);
        }

        // L2: Redis (fast)
        const redisValue = await this.l2Cache.get(key);
        if (redisValue) {
            const value = JSON.parse(redisValue);
            this.l1Cache.set(key, value); // Promote to L1
            return value;
        }

        // L3: Database (slow)
        const dbValue = await this.fetchFromDatabase(key);
        if (dbValue) {
            this.l2Cache.setex(key, 300, JSON.stringify(dbValue)); // 5min TTL
            this.l1Cache.set(key, dbValue);
            return dbValue;
        }

        return null;
    }

    async set(key, value, ttl = 300) {
        this.l1Cache.set(key, value);
        await this.l2Cache.setex(key, ttl, JSON.stringify(value));
    }

    async invalidate(key) {
        this.l1Cache.delete(key);
        await this.l2Cache.del(key);
    }
}
```

---

## 9. Performance Optimization Strategies

### Server-Side Optimizations

#### 1. Batch Processing

```javascript
// Process game ticks in batches
class BatchProcessor {
    constructor(batchSize = 1000) {
        this.batchSize = batchSize;
        this.queue = [];
    }

    async processTick() {
        // Get all entities that need processing
        const entities = await db.query(`
            SELECT id, type, player_id, last_tick_processed
            FROM game_entities
            WHERE last_tick_processed < NOW() - INTERVAL '5 seconds'
            LIMIT 100000
        `);

        // Process in batches
        for (let i = 0; i < entities.length; i += this.batchSize) {
            const batch = entities.slice(i, i + this.batchSize);

            // Parallel processing within batch
            await Promise.all(batch.map(entity => this.processEntity(entity)));
        }
    }

    async processEntity(entity) {
        // Entity-specific logic (resource production, construction, etc.)
        const updates = this.calculateUpdates(entity);

        // Queue database update
        this.queue.push({
            table: entity.type,
            id: entity.id,
            updates
        });

        // Flush queue when full
        if (this.queue.length >= this.batchSize) {
            await this.flushQueue();
        }
    }

    async flushQueue() {
        if (this.queue.length === 0) return;

        // Bulk update using batch SQL
        // (Much faster than individual updates)
        await db.batchUpdate(this.queue);
        this.queue = [];
    }
}
```

#### 2. Connection Pooling

```javascript
// PostgreSQL connection pool
const { Pool } = require('pg');

const pool = new Pool({
    host: process.env.DB_HOST,
    port: 5432,
    database: 'galactic_dominion',
    user: process.env.DB_USER,
    password: process.env.DB_PASS,

    // Pool configuration
    max: 20,              // Maximum pool size
    idleTimeoutMillis: 30000,
    connectionTimeoutMillis: 2000,
});

// Use pool for queries
async function queryDatabase(sql, params) {
    const client = await pool.connect();
    try {
        const result = await client.query(sql, params);
        return result.rows;
    } finally {
        client.release(); // Return to pool
    }
}
```

#### 3. Query Optimization

```sql
-- Create indexes for common queries
CREATE INDEX idx_fleets_position ON fleets USING GIST (
    point(current_x, current_y)
);

CREATE INDEX idx_colonies_sector ON colonies(sector_id);

CREATE INDEX idx_market_orders_active ON market_orders(resource_type, order_type)
WHERE status = 'active';

-- Use covering indexes
CREATE INDEX idx_players_resources ON players(id)
INCLUDE (energy, minerals, rare_elements, population);

-- Partitioning for large tables
CREATE TABLE market_transactions_y2025m11 PARTITION OF market_transactions
FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
```

### Client-Side Optimizations

#### 1. Asset Loading

```javascript
// Progressive asset loading
class AssetManager {
    async loadAssets() {
        // Critical assets (load immediately)
        await this.loadCriticalAssets([
            'ui/hud.png',
            'models/colony_icon.glb',
            'models/fleet_icon.glb'
        ]);

        // Show loading screen with "Play" button
        showLoadingScreen();

        // Background loading (non-blocking)
        this.loadSecondaryAssets([
            'models/colony_detailed.glb',
            'models/ships/*',
            'textures/space_skybox/*'
        ]).then(() => {
            console.log('All assets loaded');
        });
    }

    async loadCriticalAssets(paths) {
        return Promise.all(paths.map(path => this.loadAsset(path)));
    }

    loadSecondaryAssets(paths) {
        // Load in background, non-blocking
        return new Promise(resolve => {
            requestIdleCallback(async () => {
                await Promise.all(paths.map(path => this.loadAsset(path)));
                resolve();
            });
        });
    }
}
```

#### 2. Render Optimization

```javascript
// Babylon.js optimization settings
scene.autoClear = false; // Manual clearing
scene.autoClearDepthAndStencil = false;

// Use hardware scaling for lower-end devices
if (deviceTier === 'low') {
    engine.setHardwareScalingLevel(1.5); // Render at 66% resolution
}

// Freeze unchanging objects
staticMeshes.forEach(mesh => mesh.freezeWorldMatrix());

// Use occlusion queries
scene.enableOcclusionQuery = true;

// Limit shadow generators
shadowGenerator.mapSize = 1024; // Instead of 2048
shadowGenerator.useBlurExponentialShadowMap = true;
shadowGenerator.usePoissonSampling = false; // Faster

// Particle system optimization
particleSystem.manualEmitCount = 100; // Limit particles
particleSystem.targetStopDuration = 2; // Short-lived effects
```

#### 3. Memory Management

```javascript
// Dispose unused assets
class EntityManager {
    constructor(scene) {
        this.scene = scene;
        this.entities = new Map();
        this.disposalQueue = [];
    }

    addEntity(id, mesh) {
        this.entities.set(id, {
            mesh,
            lastSeen: Date.now()
        });
    }

    removeEntity(id) {
        this.disposalQueue.push(id);
    }

    cleanup() {
        // Run cleanup every 10 seconds
        setInterval(() => {
            const now = Date.now();

            // Remove entities not seen in 30 seconds
            for (const [id, entity] of this.entities.entries()) {
                if (now - entity.lastSeen > 30000) {
                    this.disposalQueue.push(id);
                }
            }

            // Dispose queued entities
            for (const id of this.disposalQueue) {
                const entity = this.entities.get(id);
                if (entity) {
                    entity.mesh.dispose(); // Free GPU memory
                    this.entities.delete(id);
                }
            }

            this.disposalQueue = [];

            // Force garbage collection hint
            if (window.gc) window.gc();
        }, 10000);
    }
}
```

---

## 10. Development Roadmap

### Phase 1: Prototype (Months 1-3)
**Goal**: Prove core concepts

**Deliverables**:
- [ ] Basic galaxy map rendering (1000 entities)
- [ ] Player authentication and account creation
- [ ] Simple colony management (resource production)
- [ ] Basic WebSocket communication
- [ ] Database schema v1
- [ ] Single-server deployment

**Success Criteria**:
- 50 concurrent players
- 30 FPS on mid-tier devices
- <500ms average latency

### Phase 2: Core Features (Months 4-6)
**Goal**: Implement main gameplay loops

**Deliverables**:
- [ ] Combat system (real-time battles)
- [ ] Market economy (player trading)
- [ ] Alliance system (basic)
- [ ] Research trees
- [ ] Fleet management
- [ ] Anti-cheat v1

**Success Criteria**:
- 500 concurrent players
- Combat feels responsive (<100ms)
- Market has liquidity (100+ orders)

### Phase 3: Scaling (Months 7-9)
**Goal**: Support massive player counts

**Deliverables**:
- [ ] Microservices architecture
- [ ] Database sharding
- [ ] Load balancing
- [ ] Advanced map rendering (10k+ entities)
- [ ] WebAssembly optimization
- [ ] CDN integration

**Success Criteria**:
- 5000+ concurrent players
- 60 FPS with 10k entities visible
- <200ms p99 latency

### Phase 4: Polish & Launch (Months 10-12)
**Goal**: Production-ready release

**Deliverables**:
- [ ] Advanced alliance governance
- [ ] Tutorial system
- [ ] Admin tools
- [ ] Analytics dashboard
- [ ] Comprehensive testing
- [ ] Launch marketing campaign

**Success Criteria**:
- 10,000+ registered players at launch
- <0.1% critical bug rate
- 99.9% uptime

### Phase 5: Post-Launch (Months 13+)
**Goal**: Live operations and expansion

**Deliverables**:
- [ ] Seasonal events
- [ ] New ship types / buildings
- [ ] Territory control mechanics
- [ ] Mobile app (React Native)
- [ ] Esports features (tournaments)

---

## 11. Technology Stack Summary

### Frontend
| Component | Technology | Reasoning |
|-----------|-----------|-----------|
| 3D Engine | Babylon.js 7.x | WebGPU, physics, LOD support |
| UI Framework | React 18+ | Ecosystem, developer familiarity |
| State Management | Zustand + Immer | Simplicity, performance |
| Networking | Socket.IO | WebSocket with fallbacks |
| Binary Protocol | MessagePack | 60% smaller than JSON |
| WASM | Rust (wasm-bindgen) | Max performance for critical paths |

### Backend
| Component | Technology | Reasoning |
|-----------|-----------|-----------|
| API Server | Node.js (NestJS) | TypeScript, structure, ecosystem |
| Game Logic | Rust | Performance for tick processing |
| WebSocket Gateway | Node.js (Socket.IO) | Integration with API server |
| Message Bus | Redis Pub/Sub | Low latency, simple |
| Task Queue | Bull (Redis-backed) | Reliable background jobs |

### Data Layer
| Component | Technology | Reasoning |
|-----------|-----------|-----------|
| Primary DB | PostgreSQL 16+ | ACID, relational data, JSON support |
| Cache | Redis 7+ | In-memory speed |
| Time-Series | TimescaleDB | Efficient historical queries |
| Document Store | MongoDB 7+ | Flexible schema for logs/chat |
| Search | Elasticsearch 8+ | Full-text, geospatial search |

### Infrastructure
| Component | Technology | Reasoning |
|-----------|-----------|-----------|
| Container | Docker | Standardized deployment |
| Orchestration | Kubernetes | Auto-scaling, resilience |
| Load Balancer | NGINX | Performance, WebSocket support |
| CDN | CloudFlare | DDoS protection, global reach |
| Monitoring | Grafana + Prometheus | Metrics, alerting |
| Logging | ELK Stack | Centralized log analysis |

---

## 12. Conclusion

This blueprint provides a comprehensive technical foundation for building a scalable, performant browser-based MMORTS game. The architecture emphasizes:

1. **Performance**: WebAssembly, instanced rendering, LOD, delta compression
2. **Scalability**: Microservices, horizontal scaling, sharding, caching
3. **Real-time**: WebSocket protocol, client prediction, server reconciliation
4. **Security**: Input validation, rate limiting, anti-cheat, encryption
5. **Complexity**: Intricate combat, dynamic economy, advanced governance

**Key Innovations**:
- Hybrid rendering (3D + 2D) for massive entity counts
- Physics-lite combat for realism without performance cost
- Player-driven economy with regional scarcity
- Democratic alliance governance with weighted voting
- Multi-tier caching and synchronization

**Next Steps**:
1. Set up development environment
2. Implement prototype (Phase 1)
3. Conduct performance testing with simulated load
4. Iterate based on metrics and player feedback
5. Scale infrastructure progressively

**Estimated Development Time**: 12-18 months with a team of 5-8 developers

**Estimated Infrastructure Cost** (at 10k concurrent players):
- Servers: $2,000-$5,000/month
- Database: $500-$1,500/month
- CDN: $200-$500/month
- Total: ~$3,000-$7,000/month

This game is technically ambitious but achievable with modern web technologies. The key to success is iterative development, continuous performance testing, and player-centric design.

---

**Document Version**: 1.0
**Last Updated**: 2025-11-03
**Author**: Technical Architecture Team
