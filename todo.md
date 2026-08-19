
## V1 - WebSocket Gateway

**Goal:** Establish reliable real-time communication between clients and  
the server.

Basic Goal: Players can connect to the server, choose a unique username, send chat messages, and every message is broadcast to all connected players. Players can join and leave cleanly without affecting the server.

### Responsibilities

- Accept WebSocket connections.

- Assign usernames or validate chosen names.

- Handle connection lifecycle (connect/disconnect).

- Implement heartbeat (Ping/Pong).

- Maintain a registry of active connections.

- Serialize and deserialize protocol messages.

- Gracefully clean up dead connections.

**Completion Criteria** - Multiple clients can connect simultaneously. -  
Clients can exchange messages through the server. - Disconnects are  
detected and cleaned up correctly.

## V2 - Room Manager

**Goal:** Allow players to organize into isolated game rooms.

### Responsibilities

- Create rooms.

- Join existing rooms.

- Leave rooms.

- Broadcast player join/leave events.

- Track room membership.

- Delete empty rooms.

- Prevent duplicate usernames within a room.

**Completion Criteria** - Multiple rooms can exist independently. -  
Events stay isolated to the correct room.

## V3 - Game Engine

**Goal:** Control the overall game flow.

### Responsibilities

- Manage game phases.

- Rotate the drawer.

- Select or assign words.

- Start and stop round timers.

- Transition between rounds.

- End the game when appropriate.

**Completion Criteria** - The game progresses automatically through  
complete rounds without manual intervention.

## V4 - Guessing & Scoring

**Goal:** Implement the core gameplay rules.

### Responsibilities

- Validate guesses.

- Detect correct guesses.

- Award points to guessers.

- Award points to the drawer.

- Broadcast guess results.

- End rounds early when everyone has guessed correctly.

- Maintain a live scoreboard.

**Completion Criteria** - Guessing and scoring behave consistently  
throughout multiple rounds.

## V5 - Browser Client (Minimal)

**Goal:** Build the simplest possible interface to interact with the  
backend.

### Responsibilities

- Connect to the WebSocket server.

- Display chat.

- Display player list.

- Display scoreboard.

- Show game status and timers.

- Join/create rooms.

**Intentionally Excluded** - Drawing. - Fancy UI. - Animations. -  
Responsive design.

**Completion Criteria** - The entire game can be played except for the  
drawing portion.

## V6 - Drawing Protocol

**Goal:** Add real-time collaborative drawing.

### Responsibilities

- Define drawing message types.

- Send drawing events to the server.

- Broadcast drawing events.

- Synchronize canvases.

- Handle brush color and size.

- Clear the canvas between rounds.

- Send the current canvas state to late joiners if needed.

**Completion Criteria** - All players see the same drawing in real time  
with minimal latency.

## V7 - Gameplay Enhancements

**Goal:** Add features that improve gameplay and match Skribbl-like  
behavior.

### Responsibilities

- Three-word selection for the drawer.

- Automatic word selection on timeout.

- Time-based scoring bonuses.

- Vote kick.

- Drawing likes/dislikes.

- Spectator mode.

- Better reconnect behavior.

- Additional brush tools or quality-of-life improvements.

**Completion Criteria** - The experience feels polished and feature  
complete.

## V8 - Production Readiness

**Goal:** Make the server reliable under real-world conditions.

### Responsibilities

- Structured logging.

- Metrics.

- Graceful shutdown.

- Rate limiting.

- Input validation.

- Heartbeat timeouts.

- Panic recovery.

- Race detector cleanup.

- Load testing.

- Memory and CPU profiling.

- Configuration management.

- Robust error handling.

A lot of the stuff here like logging and validation you should be doing as you build not just after everything is done.

**Completion Criteria** - The server remains stable, observable, and  
maintainable under sustained use.
