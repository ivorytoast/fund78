# Fund78 - Real-Time Bitcoin Price Processing Engine

A Rust application that streams real-time Bitcoin prices from Polygon.io, processes them through a custom event-driven engine, and broadcasts updates to WebSocket clients.

## Features

- **Real-time Bitcoin Price Streaming**: Connects to Polygon.io's WebSocket API to receive live Bitcoin trade data
- **Event-Driven Processing Engine**: Custom engine that processes events through configurable workers
- **WebSocket Server**: Broadcasts Bitcoin price updates to connected clients on `ws://localhost:8080`
- **Multi-threaded Architecture**: Runs Polygon connection, WebSocket server, and event processing on separate threads
- **Persistent Logging**: Logs all incoming events and processed outputs to `in.log` and `out.log`

## Prerequisites

- Rust (latest stable version)
- A Polygon.io API key (get one at https://polygon.io)

## Setup

### 1. Clone the Repository

```bash
git clone <your-repo-url>
cd fund78
```

### 2. Environment Variables

**IMPORTANT**: You must create a `.env` file in the project root directory before running the application.

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Open `.env` and add your Polygon.io API key:
   ```
   POLYGON_API_KEY=your_actual_api_key_here
   ```

**Note**: The `.env` file is ignored by git to keep your API key secure. Never commit your actual API key to version control.

### 3. Install Dependencies

```bash
cargo build
```

## Running the Application

```bash
cargo run
```

The application will:
1. Connect to Polygon.io's crypto WebSocket feed
2. Start a WebSocket server on `ws://localhost:8080`
3. Process Bitcoin price events through the engine
4. Log all activity to `in.log` and `out.log`

## Testing the WebSocket Connection

You can test the WebSocket server using your browser's developer console:

```javascript
const ws = new WebSocket('ws://localhost:8080');
ws.onmessage = (e) => console.log('Bitcoin update:', JSON.parse(e.data));
ws.onopen = () => console.log('Connected to Bitcoin feed');
```

You should see real-time Bitcoin trade data streaming in.

## Architecture

### Components

- **Polygon.io Connection**: Asynchronous WebSocket client that authenticates and subscribes to Bitcoin trades
- **Event Engine**: Processes events through registered workers with persistent file handles for efficient logging
- **WebSocket Server**: Accepts client connections and broadcasts price updates using a broadcast channel
- **Workers**: Pluggable event handlers that process specific task types

### High-Level Data Flow

```
Polygon.io → Bitcoin Trade Event → Engine Processing → WebSocket Broadcast
                                  ↓
                            in.log & out.log
```

### Detailed Internal Flow

#### Thread Architecture

```
┌─────────────────────┐
│   Main Thread       │
│  - Initial Setup    │
│  - Thread Spawning  │
└─────────────────────┘
          │
          ├──────────────────────────────────────────────┐
          │                                              │
          ▼                                              ▼
┌─────────────────────┐                    ┌─────────────────────┐
│  Polygon Thread     │                    │  WebSocket Thread   │
│  (Tokio Runtime)    │                    │  (Tokio Runtime)    │
│                     │                    │                     │
│  Connects to        │                    │  Listens on         │
│  Polygon.io WSS     │                    │  localhost:8080     │
└─────────────────────┘                    └─────────────────────┘
          │                                              │
          │ Sends Event                                  │ Broadcasts
          │ via mpsc::channel                            │ via broadcast::channel
          ▼                                              ▼
┌─────────────────────┐                    ┌─────────────────────┐
│  Engine Thread      │                    │  Connected Clients  │
│                     │                    │  (Browsers, Apps)   │
│  Processes Events   │                    └─────────────────────┘
│  via Workers        │
│                     │
│  Writes to Logs     │
└─────────────────────┘
```

#### Bitcoin Price Flow (Step-by-Step)

```
1. Polygon.io sends Bitcoin trade data
   ↓
   [Polygon Thread - get_market_data()]
   
2. Deserialize JSON into CryptoTrade struct
   ↓
   {
     "ev": "XT",
     "pair": "BTC-USD", 
     "p": 96261.47,      ← Bitcoin price
     "t": 1763179530104,
     "s": 0.00010288,
     ...
   }
   
3. Create Event struct
   ↓
   Event {
     task: "bitcoin_price",
     payload: 96261 (price as i32)
   }
   
4. Send to Engine via mpsc::channel
   ↓
   event_tx.send(event) 
   
   [Channel automatically wakes Engine Thread]
   
5. Engine Thread receives event
   ↓
   event_rx.recv() → Returns Event
   
6. Write to in.log
   ↓
   {"task":"bitcoin_price","payload":96261}
   
7. Match event to Worker
   ↓
   Worker: handles_task == "bitcoin_price"
   
8. Execute Worker's job function
   ↓
   handle_bitcoin_price(96261)
   
9. Worker returns new Event
   ↓
   Event {
     task: "bitcoin_price_accepted",
     payload: 0  // timestamp formatting
   }
   
10. Write to out.log
    ↓
    {"task":"bitcoin_price_accepted","payload":0}
```

#### Concurrent Broadcast Flow

While the engine processes events, the Polygon thread also broadcasts raw data:

```
Polygon Thread
   │
   ├─── event_tx.send(event) ──→ [Engine Thread]
   │
   └─── broadcast_tx.send(json) ──→ [Broadcast Channel]
                                          │
                                          ├──→ Client 1
                                          ├──→ Client 2  
                                          └──→ Client N
```

### Channel Types Used

1. **`mpsc::channel`** (Polygon → Engine)
   - Multi-producer, single-consumer
   - Used for: Sending Bitcoin price events to engine
   - Blocking: Engine thread sleeps until events arrive (zero CPU when idle)

2. **`broadcast::channel`** (Polygon → WebSocket Clients)
   - Multi-producer, multi-consumer
   - Used for: Broadcasting raw JSON to all connected WebSocket clients
   - Non-blocking: Clients receive updates asynchronously

### Why Channels?

- **Zero CPU Usage When Idle**: The engine thread blocks on `recv()` instead of polling
- **Thread Safety**: Safe communication between threads without locks
- **Immediate Processing**: Events are processed instantly when they arrive
- **Batch Optimization**: Multiple events arriving simultaneously are processed together

## Project Structure

```
fund78/
├── src/
│   ├── main.rs           # Application entry point, threading, and WebSocket logic
│   ├── lib.rs            # Engine and Event definitions
│   └── sample.rs         # Sample events and workers
├── .env.example          # Example environment variables
├── .gitignore           # Git ignore rules (includes .env)
├── Cargo.toml           # Project dependencies
└── README.md            # This file
```

## Dependencies

- `tokio` - Async runtime
- `tokio-tungstenite` - WebSocket client/server
- `futures-util` - Async utilities
- `serde` & `serde_json` - JSON serialization
- `dotenv` - Environment variable management

## Development

### Adding New Workers

To add a new event worker:

1. Define a handler function:
   ```rust
   fn handle_my_event(payload: i32) -> Event {
       Event {
           task: "my_output".to_string(),
           payload: payload * 2,
       }
   }
   ```

2. Register it in `main.rs`:
   ```rust
   workers.push(Worker {
       handles_task: "my_task".to_string(),
       job: handle_my_event,
   });
   ```

### Log Files

- `in.log` - Records all incoming events to the engine
- `out.log` - Records all processed output events

Both files append data (never overwrite) and use JSON format with one event per line.

## Troubleshooting

### "POLYGON_API_KEY must be set in .env file"

Make sure you've created a `.env` file in the project root with your API key.

### "Failed to connect to Polygon"

- Verify your API key is valid
- Check your internet connection
- Ensure the `native-tls` feature is enabled in `Cargo.toml`

### WebSocket clients can't connect

- Ensure port 8080 is not in use by another application
- Check firewall settings
