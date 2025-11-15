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

## Quick Start Example

The Fund78 API is designed to be simple - you only need to create workers, inputs, outputs, and let the engine handle everything:

```rust
use fund78::{BitcoinPriceWorker, Engine, Input, Output, PolygonInput, WebSocketOutput, Worker};

fn main() {
    // 1. Create workers
    let workers: Vec<Box<dyn Worker>> = vec![
        Box::new(BitcoinPriceWorker::new())
    ];

    // 2. Create inputs
    let inputs: Vec<Box<dyn Input>> = vec![
        Box::new(PolygonInput::new(api_key))
    ];

    // 3. Create outputs
    let outputs: Vec<Box<dyn Output>> = vec![
        Box::new(WebSocketOutput::new("127.0.0.1:8080".to_string()))
    ];

    // 4. Start the engine - it handles all threading automatically!
    let engine = Engine::new(workers, inputs, outputs).unwrap();
    engine.run();  // Spawns threads, manages lifecycle, keeps app alive
}
```

**That's it!** No manual thread spawning, no handle management, no joining threads. The engine does it all.

## Testing the WebSocket Connection

You can test the WebSocket server using your browser's developer console:

```javascript
const ws = new WebSocket('ws://localhost:8080');
ws.onmessage = (e) => console.log('Bitcoin update:', JSON.parse(e.data));
ws.onopen = () => console.log('Connected to Bitcoin feed');
```

You should see real-time Bitcoin trade data streaming in.

## Architecture

### Core Design Philosophy

Fund78 enforces a **deterministic, unidirectional data flow** through abstraction layers:

- **Inputs** can ONLY send events to the engine (never receive)
- **Outputs** can ONLY receive events from the engine (never send)
- **Engine** is the single-threaded orchestrator that ALL data flows through
- **No direct communication** between inputs and outputs

This architecture guarantees predictable behavior and makes testing/debugging straightforward.

### Abstraction Layers

#### 1. Input Abstraction (`src/input/input.rs`)

```rust
pub trait Input: Send {
    fn start(self: Box<Self>, handle: InputHandle);
}
```

**InputHandle** provides:
- `send_event(event: Event)` - Send events to engine
- Complete abstraction from underlying channels (users never see `mpsc::Sender`)

**Creating a new input:**
```rust
pub struct MyInput { /* config */ }

impl Input for MyInput {
    fn start(self: Box<Self>, handle: InputHandle) {
        // Your logic here
        handle.send_event(Event { task: "...", payload: 0 });
    }
}
```

#### 2. Output Abstraction (`src/output/output.rs`)

```rust
pub trait Output: Send {
    fn start(self: Box<Self>, handle: OutputHandle);
}
```

**OutputHandle** provides:
- `subscribe()` - Get a receiver for multiple consumers
- `recv()` / `blocking_recv()` - Receive single events
- Complete abstraction from underlying channels (users never see `broadcast::Sender`)

**Creating a new output:**
```rust
pub struct MyOutput { /* config */ }

impl Output for MyOutput {
    fn start(self: Box<Self>, handle: OutputHandle) {
        let mut rx = handle.subscribe();
        while let Ok(event) = rx.blocking_recv() {
            // Process event
        }
    }
}
```

#### 3. Worker Abstraction (`src/worker/worker.rs`)

```rust
pub trait Worker: Send + Sync {
    fn handles_task(&self) -> &str;
    fn process(&self, event: Event) -> Event;
}
```

**Worker trait clearly defines:**
- Workers handle events that match their specific topic/task
- They execute a function that takes an Event and returns another Event
- Complete abstraction for event processing logic

**Creating a new worker:**
```rust
pub struct BitcoinPriceWorker;

impl BitcoinPriceWorker {
    pub fn new() -> Self {
        BitcoinPriceWorker
    }
}

impl Worker for BitcoinPriceWorker {
    fn handles_task(&self) -> &str {
        "bitcoin_price"
    }

    fn process(&self, event: Event) -> Event {
        let timestamp = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs();

        Event {
            task: "bitcoin_price_accepted".to_string(),
            payload: format!("bitcoin price {} accepted at time: {}", event.payload, timestamp)
                .parse()
                .unwrap_or(0),
        }
    }
}
```

**FunctionWorker for quick implementations:**
```rust
// For simple use cases, use FunctionWorker instead of creating a struct
let worker = FunctionWorker::new("my_task".to_string(), |event| {
    Event {
        task: "my_task_processed".to_string(),
        payload: event.payload * 2,
    }
});
```

#### 4. Engine as Orchestrator (`src/lib.rs`)

The **Engine** manages all inputs, outputs, and workers with automatic thread management:

```rust
// 1. Create workers
let workers: Vec<Box<dyn Worker>> = vec![
    Box::new(BitcoinPriceWorker::new())
];

// 2. Create inputs
let inputs: Vec<Box<dyn Input>> = vec![
    Box::new(PolygonInput::new(api_key))
];

// 3. Create outputs
let outputs: Vec<Box<dyn Output>> = vec![
    Box::new(WebSocketOutput::new("127.0.0.1:8080".to_string()))
];

// 4. Create and run engine - it handles all threading automatically
let engine = Engine::new(workers, inputs, outputs)?;
engine.run();  // Blocks until all threads complete
```

**Key guarantees:**
- Engine spawns and manages all threads automatically
- Handles are created internally - impossible to bypass the engine
- All events flow through deterministic processing
- Simple API - no manual thread management required

### Components

- **PolygonInput** (`src/input/polygon_input.rs`): WebSocket client for Polygon.io Bitcoin trades
- **Engine** (`src/lib.rs`): Single-threaded event processor with workers
- **WebSocketOutput** (`src/output/websocket_output.rs`): WebSocket server broadcasting to clients
- **Workers** (`src/worker/`): Pluggable event handlers that implement the Worker trait
  - **BitcoinPriceWorker** (`src/worker/bitcoin_price_worker.rs`): Processes bitcoin price events

### High-Level Data Flow

```
PolygonInput → InputHandle → Engine → OutputHandle → WebSocketOutput
   (Input)                    (Central Hub)              (Output)
                                  ↓
                            in.log & out.log
```

### Detailed Internal Flow

#### Thread Architecture

```
┌─────────────────────────────────────────────────┐
│              Main Thread                        │
│  1. Create Workers, Inputs, Outputs             │
│  2. Create Engine with all components           │
│  3. Call engine.run()                           │
│                                                  │
│  ┌────────────────────────────────────────────┐ │
│  │  Engine.run() - Spawns all threads:        │ │
│  │  - Input threads (with InputHandles)       │ │
│  │  - Output threads (with OutputHandles)     │ │
│  │  - Engine processing thread                │ │
│  │  - Joins all threads (keeps app alive)     │ │
│  └────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────┘
          │
          ├────────────────────┬────────────────────┐
          ▼                    ▼                    ▼
┌──────────────────┐  ┌────────────────┐  ┌──────────────────┐
│  Input Thread    │  │ Engine Thread  │  │  Output Thread   │
│  (PolygonInput)  │  │                │  │ (WebSocketOutput)│
│                  │  │  Central Hub   │  │                  │
│  Has:            │  │                │  │  Has:            │
│  - InputHandle   │  │  Owns:         │  │  - OutputHandle  │
│                  │  │  - Channels    │  │                  │
│  Can ONLY:       │  │  - Workers     │  │  Can ONLY:       │
│  ✓ Send events   │  │  - Log files   │  │  ✓ Recv events   │
│  ✗ Receive       │  │                │  │  ✗ Send events   │
│                  │  │  Processes:    │  │                  │
│  Connects to     │  │  - Events      │  │  Listens on      │
│  Polygon.io      │  │  - via Workers │  │  localhost:8080  │
│                  │  │  - Logging     │  │                  │
│  Sends Bitcoin   │  │  - Broadcasting│  │  Broadcasts to   │
│  trades          │  │                │  │  clients         │
└──────────────────┘  └────────────────┘  └──────────────────┘
          │                   ▲ │                   ▲
          │   InputHandle     │ │   OutputHandle    │
          │   .send_event()   │ │   .subscribe()    │
          └───────────────────┘ └───────────────────┘

     (Hidden: mpsc::Sender)   (Hidden: broadcast::Sender)
```

**Key Points:**
- Engine.run() automatically spawns all threads
- No manual thread management required
- Engine blocks until all threads complete
- Handles are created and passed internally

**Key Architectural Enforcements:**
- Input and Output threads NEVER communicate directly
- All communication goes through handles (abstractions)
- Engine owns and controls all underlying channels
- Impossible to bypass the engine's processing logic

#### Bitcoin Price Flow (Step-by-Step)

```
1. Polygon.io sends Bitcoin trade data
   ↓
   [PolygonInput - Input Thread]

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

4. Send via InputHandle abstraction
   ↓
   handle.send_event(event)  ← User API
        ↓
   (Hidden: mpsc::Sender.send())

5. Engine receives event
   ↓
   [Engine Thread - engine.run()]
   Receives from mpsc::Receiver

6. Write to in.log
   ↓
   {"task":"bitcoin_price","payload":96261}

7. Match event to Worker
   ↓
   Worker: handles_task() == "bitcoin_price"
   (BitcoinPriceWorker)

8. Execute Worker's process method
   ↓
   worker.process(event)
   Receives full Event struct with task and payload

9. Worker returns new Event
   ↓
   Event {
     task: "bitcoin_price_accepted",
     payload: 0  // timestamp formatting
   }

10. Write to out.log
    ↓
    {"task":"bitcoin_price_accepted","payload":0}

11. Broadcast via internal channel
    ↓
    broadcast::Sender.send(json_string)

12. WebSocketOutput receives via OutputHandle
    ↓
    [WebSocketOutput - Output Thread]
    rx = handle.subscribe()  ← User API
    rx.recv() → Returns event
         ↓
    (Hidden: broadcast::Receiver.recv())

13. Forward to all connected clients
    ↓
    All WebSocket clients receive:
    {"task":"bitcoin_price","payload":96261}
```

#### Complete Data Flow Diagram

```
┌──────────────┐
│ Polygon.io   │
└──────┬───────┘
       │ Bitcoin Trade JSON
       ▼
┌─────────────────────────────┐
│  PolygonInput (Input impl)  │
│  - Deserialize              │
│  - Create Event             │
└──────┬──────────────────────┘
       │
       │ handle.send_event(event)
       │ ↓ (User API)
       │ ↓
       │ mpsc::Sender (Hidden)
       │
       ▼
┌─────────────────────────────┐
│  Engine (Single-threaded)   │
│  - Receive via mpsc         │
│  - Write in.log          │ ──────► in.log
│  - Match Worker             │
│  - Process Event            │
│  - Write out.log         │ ──────► out.log
│  - Broadcast via broadcast  │
└──────┬──────────────────────┘
       │
       │ broadcast::Sender (Hidden)
       │ ↓
       │ ↓ (User API)
       │ handle.subscribe() / recv()
       │
       ▼
┌─────────────────────────────┐
│  WebSocketOutput (Output)   │
│  - Accept Clients           │
│  - Forward Messages         │
└──────┬──────────────────────┘
       │
       ├──────► Client 1 (Browser)
       ├──────► Client 2 (Browser)
       └──────► Client N (Browser)
```

**Abstraction Boundaries:**
- Top layer: PolygonInput/WebSocketOutput (user implementations)
- Middle layer: InputHandle/OutputHandle (user API)
- Bottom layer: mpsc/broadcast channels (hidden from users)

### Channel Types Used (Hidden Implementation)

**Note:** Users never interact with these channels directly - they use InputHandle/OutputHandle abstractions.

1. **`mpsc::channel`** (Input → Engine)
   - **Hidden in**: `InputHandle`
   - **Exposed as**: `handle.send_event(event)`
   - Multi-producer, single-consumer
   - Used for: Sending events from any input to the engine
   - Blocking: Engine thread sleeps until events arrive (zero CPU when idle)
   - **Users don't see**: `mpsc::Sender` or `mpsc::Receiver`

2. **`broadcast::channel`** (Engine → Outputs)
   - **Hidden in**: `OutputHandle`
   - **Exposed as**: `handle.subscribe()` and `handle.recv()`
   - Multi-producer, multi-consumer
   - Used for: Broadcasting processed events to all outputs
   - Non-blocking: Outputs receive updates asynchronously
   - **Users don't see**: `broadcast::Sender` or `broadcast::Receiver`

**Why hide the channels?**
- Simplifies the API for creating new inputs/outputs
- Enforces architectural constraints at compile-time
- Allows changing implementation without breaking user code
- Makes testing easier (mock handles instead of channels)

### Why This Architecture?

**Abstraction Benefits:**
- **Enforced Unidirectional Flow**: Type system prevents inputs from receiving or outputs from sending
- **Simple API**: Creating new inputs/outputs requires no knowledge of channels or threading
- **Compile-Time Safety**: Impossible to bypass the engine or create direct input→output paths
- **Testability**: Mock handles instead of mocking complex channel infrastructure
- **Flexibility**: Can change channel implementation without breaking user code

**Performance Benefits:**
- **Zero CPU When Idle**: Engine blocks on `recv()` instead of polling
- **Immediate Processing**: Events processed instantly when they arrive
- **Scalable Broadcasting**: Multiple outputs/clients without performance impact
- **Thread Safety**: Safe communication without locks or mutexes

**Architectural Benefits:**
- **Deterministic Processing**: Single-threaded engine ensures predictable event ordering
- **Central Logging**: All business logic and logging in one place (engine)
- **Separation of Concerns**: Clear boundaries between input/processing/output
- **Extensibility**: Add new inputs/outputs by implementing simple traits

## Project Structure

```
fund78/
├── src/
│   ├── main.rs                          # Application entry point
│   ├── lib.rs                           # Engine, Event definitions
│   ├── input/                           # Input abstraction and implementations
│   │   ├── mod.rs                       # Input module exports
│   │   ├── input.rs                     # Input trait & InputHandle abstraction
│   │   └── polygon_input.rs             # Polygon.io input implementation
│   ├── output/                          # Output abstraction and implementations
│   │   ├── mod.rs                       # Output module exports
│   │   ├── output.rs                    # Output trait & OutputHandle abstraction
│   │   └── websocket_output.rs          # WebSocket server output implementation
│   └── worker/                          # Worker abstraction and implementations
│       ├── mod.rs                       # Worker module exports
│       ├── worker.rs                    # Worker trait & FunctionWorker
│       └── bitcoin_price_worker.rs      # Bitcoin price worker implementation
├── .env.example                         # Example environment variables
├── .gitignore                           # Git ignore rules (includes .env)
├── Cargo.toml                           # Project dependencies
└── README.md                            # This file
```

**Folder Structure:**
- **input/**: All input-related code (traits and implementations)
- **output/**: All output-related code (traits and implementations)
- **worker/**: All worker-related code (traits and implementations)

**File Responsibilities:**
- **main.rs**: Minimal setup - creates workers/inputs/outputs, starts engine
- **lib.rs**: Core engine logic, event processing, and automatic thread management
- **input/input.rs**: Input abstraction that hides channel complexity
- **input/polygon_input.rs**: Concrete input implementation for Polygon.io
- **output/output.rs**: Output abstraction that hides channel complexity
- **output/websocket_output.rs**: Concrete output implementation for WebSocket server
- **worker/worker.rs**: Worker trait and FunctionWorker implementation
- **worker/bitcoin_price_worker.rs**: Concrete worker for processing bitcoin prices

## Dependencies

- `tokio` - Async runtime
- `tokio-tungstenite` - WebSocket client/server
- `futures-util` - Async utilities
- `serde` & `serde_json` - JSON serialization
- `dotenv` - Environment variable management

## Development

### Adding New Inputs

To add a new input source (e.g., REST API, Kafka, file watcher):

1. **Create your input struct and implement the Input trait:**

```rust
// src/input/my_input.rs
use crate::{Input, InputHandle, Event};

pub struct MyInput {
    config: String,
}

impl MyInput {
    pub fn new(config: String) -> Self {
        MyInput { config }
    }
}

impl Input for MyInput {
    fn start(self: Box<Self>, handle: InputHandle) {
        // Your input logic here
        loop {
            // Get data from your source
            let data = fetch_data(&self.config);

            // Send to engine
            let event = Event {
                task: "my_task".to_string(),
                payload: data,
            };
            handle.send_event(event).unwrap();
        }
    }
}
```

2. **Register it in main.rs:**

```rust
let inputs: Vec<Box<dyn Input>> = vec![
    Box::new(PolygonInput::new(api_key)),
    Box::new(MyInput::new("my_config".to_string())),
];

let engine = Engine::new(workers, inputs, outputs)?;
engine.run();  // Automatically spawns thread for your input
```

**Key points:**
- No need to spawn threads manually - Engine does it for you
- No need to know about channels or thread management
- Just call `handle.send_event()` to send events
- Engine handles all the threading complexity

### Adding New Outputs

To add a new output destination (e.g., database, file, HTTP endpoint):

1. **Create your output struct and implement the Output trait:**

```rust
// src/output/my_output.rs
use crate::{Output, OutputHandle};

pub struct MyOutput {
    config: String,
}

impl MyOutput {
    pub fn new(config: String) -> Self {
        MyOutput { config }
    }
}

impl Output for MyOutput {
    fn start(self: Box<Self>, handle: OutputHandle) {
        let mut rx = handle.subscribe();

        // Receive and process events
        while let Ok(event) = rx.blocking_recv() {
            // Your output logic here
            write_to_destination(&self.config, event);
        }
    }
}
```

2. **Register it in main.rs:**

```rust
let outputs: Vec<Box<dyn Output>> = vec![
    Box::new(WebSocketOutput::new("127.0.0.1:8080".to_string())),
    Box::new(MyOutput::new("my_config".to_string())),
];

let engine = Engine::new(workers, inputs, outputs)?;
engine.run();  // Automatically spawns thread for your output
```

**Key points:**
- No need to spawn threads manually - Engine does it for you
- No need to know about channels or thread management
- Just call `handle.subscribe()` or `handle.recv()` to receive events
- Engine handles all the threading complexity

### Adding New Workers

To add a new event worker for processing, you have two options:

#### Option 1: Create a Dedicated Worker Struct (Recommended)

1. Create a new file `src/worker/my_worker.rs`:
   ```rust
   use crate::{Event, Worker};

   pub struct MyWorker {
       // Add any configuration fields here
   }

   impl MyWorker {
       pub fn new() -> Self {
           MyWorker {}
       }
   }

   impl Worker for MyWorker {
       fn handles_task(&self) -> &str {
           "my_task"
       }

       fn process(&self, event: Event) -> Event {
           // Your processing logic here
           Event {
               task: "my_task_processed".to_string(),
               payload: event.payload * 2,
           }
       }
   }
   ```

2. Export it in `src/worker/mod.rs`:
   ```rust
   pub mod my_worker;
   pub use my_worker::MyWorker;
   ```

3. Export it in `src/lib.rs`:
   ```rust
   pub use worker::{BitcoinPriceWorker, FunctionWorker, MyWorker, Worker};
   ```

4. Register it in `main.rs`:
   ```rust
   let workers: Vec<Box<dyn Worker>> = vec![
       Box::new(BitcoinPriceWorker::new()),
       Box::new(MyWorker::new()),
   ];

   let engine = Engine::new(workers, inputs, outputs)?;
   engine.run();
   ```

#### Option 2: Use FunctionWorker (For Simple Cases)

For simple one-off workers, use `FunctionWorker`:

```rust
use fund78::{FunctionWorker, Worker};

let workers: Vec<Box<dyn Worker>> = vec![
    Box::new(BitcoinPriceWorker::new()),
    Box::new(FunctionWorker::new("my_task".to_string(), |event| {
        Event {
            task: "my_task_processed".to_string(),
            payload: event.payload * 2,
        }
    })),
];

let engine = Engine::new(workers, inputs, outputs)?;
engine.run();
```

**When to use each approach:**
- **Dedicated Worker Struct**: Complex logic, needs configuration, reusable, testable
- **FunctionWorker**: Simple transformations, one-off use cases, prototyping

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
