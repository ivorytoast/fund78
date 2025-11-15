# Lesson: Understanding Channels (mpsc & broadcast)

**Date:** 2025-11-15
**Topics:** Channels, `mpsc`, `broadcast`, Thread Communication, InputHandle, OutputHandle

---

## 📚 Part 1: What Are Channels?

### The Problem: Threads Need to Communicate

Imagine you have multiple threads running:
- **PolygonInput thread** - Receiving Bitcoin prices
- **Engine thread** - Processing events
- **WebSocketOutput thread** - Sending to clients

**Question:** How do they safely share data without data races or corruption?

**Answer:** Channels! 🎉

### What is a Channel?

A channel is like a **pipe** between threads:
- One end **sends** data
- The other end **receives** data
- Rust guarantees it's **thread-safe**

```
Thread 1                    Thread 2
┌──────────┐               ┌──────────┐
│ Sender   │──[Channel]───>│ Receiver │
└──────────┘               └──────────┘
```

### Rust's Golden Rule

**"Don't communicate by sharing memory; share memory by communicating."**

Instead of:
```rust
// ❌ Multiple threads accessing shared data (needs locks, mutexes, scary!)
let shared_data = Arc<Mutex<Vec<Event>>>;
```

Do this:
```rust
// ✅ Send data through a channel (safe, simple, fast!)
sender.send(event).unwrap();
```

---

## 📚 Part 2: `mpsc` Channels (Multi-Producer, Single-Consumer)

### What is mpsc?

`mpsc` stands for **M**ulti-**P**roducer, **S**ingle-**C**onsumer.

```
Multiple Senders              One Receiver
┌──────────┐                 ┌──────────┐
│ Input 1  │──┐              │          │
└──────────┘  │              │  Engine  │
              ├─[Channel]───>│          │
┌──────────┐  │              │          │
│ Input 2  │──┘              └──────────┘
└──────────┘
```

**Perfect for:** Many inputs sending to one engine!

### How Your Code Uses mpsc

Let's look at your `lib.rs`:

```rust
// Creating the channel
let (input_sender, input_receiver) = mpsc::channel();
//   ^^^^^^^^^^^^  ^^^^^^^^^^^^^^
//   |             |
//   |             Engine keeps this (receives events)
//   |
//   Cloned and given to each input (sends events)
```

### Your InputHandle (The Abstraction)

In `src/input/input.rs`:

```rust
pub struct InputHandle {
    sender: mpsc::Sender<Event>,  // Hidden from users!
}

impl InputHandle {
    pub fn send_event(&self, event: Event) -> Result<(), SendError<Event>> {
        self.sender.send(event)  // Simple API!
    }
}
```

**Why hide the channel?**
- Users don't need to know about `mpsc`
- Simpler API: just call `send_event()`
- Can change implementation later without breaking code
- Enforces architecture: inputs can ONLY send

### Memory Model: mpsc

```
┌─────────────────────────────────────────┐
│  Input Thread 1                         │
│  ┌────────────────┐                     │
│  │ InputHandle    │                     │
│  │ ├─ sender ─────┼─┐                   │
│  └────────────────┘ │                   │
└─────────────────────┼───────────────────┘
                      │
┌─────────────────────┼───────────────────┐
│  Input Thread 2     │                   │
│  ┌────────────────┐ │                   │
│  │ InputHandle    │ │                   │
│  │ ├─ sender ─────┼─┤                   │
│  └────────────────┘ │                   │
└─────────────────────┼───────────────────┘
                      │
                      ▼
              ┌──────────────┐
              │    Queue     │ (In heap memory)
              │  [Event]     │
              │  [Event]     │
              │  [Event]     │
              └──────────────┘
                      │
                      ▼
┌─────────────────────────────────────────┐
│  Engine Thread                          │
│  ┌────────────────┐                     │
│  │  Receiver      │                     │
│  │                │<─── recv() blocks   │
│  └────────────────┘      until data!    │
└─────────────────────────────────────────┘
```

### Key Properties of mpsc

1. **Multiple senders** - Clone the sender, give to many threads
2. **One receiver** - Cannot be cloned
3. **FIFO** - First In, First Out (ordered)
4. **Blocking** - `recv()` sleeps until data arrives (zero CPU waste!)
5. **Bounded or Unbounded** - Queue can grow or have a limit

---

## 📚 Part 3: `broadcast` Channels (Multi-Producer, Multi-Consumer)

### What is broadcast?

A broadcast channel lets you send to **multiple receivers** at once!

```
One Sender                 Multiple Receivers
┌──────────┐              ┌──────────┐
│          │──[Channel]──>│ Output 1 │
│  Engine  │              └──────────┘
│          │              ┌──────────┐
│          │──[Channel]──>│ Output 2 │
└──────────┘              └──────────┘
                          ┌──────────┐
                          │ Output 3 │
                          └──────────┘
```

**Perfect for:** Engine broadcasting to many outputs!

### How Your Code Uses broadcast

In `lib.rs`:

```rust
// Creating the broadcast channel
let (broadcast_sender, _) = broadcast::channel(100);
//   ^^^^^^^^^^^^^^^^  ^
//   |                 |
//   |                 Dropped (we create receivers later)
//   |
//   Engine keeps this (sends to all outputs)
```

Notice `Arc<broadcast::Sender<String>>`? That's so we can clone and share it!

### Your OutputHandle (The Abstraction)

In `src/output/output.rs`:

```rust
pub struct OutputHandle {
    sender: Arc<broadcast::Sender<String>>,  // Hidden from users!
}

impl OutputHandle {
    pub fn subscribe(&self) -> broadcast::Receiver<String> {
        self.sender.subscribe()  // Create a new receiver!
    }

    pub async fn recv(&self) -> Result<String, RecvError> {
        let mut rx = self.subscribe();
        rx.recv().await  // Simple API!
    }
}
```

**Why hide the channel?**
- Users don't need to know about `broadcast`
- Simple API: just call `subscribe()` or `recv()`
- Each output can have multiple receivers (for multiple WebSocket clients!)

### Memory Model: broadcast

```
┌─────────────────────────────────────────┐
│  Engine Thread                          │
│  ┌────────────────────────────────┐     │
│  │ Arc<broadcast::Sender>         │     │
│  │  ├─ pointer ───────────────┐   │     │
│  └────────────────────────────│───┘     │
└───────────────────────────────│─────────┘
                                │
                                ▼
                        ┌──────────────┐
                        │ Broadcast    │ (Shared, on heap)
                        │ Channel      │
                        │ - Circular   │
                        │   Buffer     │
                        └──────────────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
                ▼               ▼               ▼
┌───────────────────┐ ┌───────────────┐ ┌───────────────┐
│ Output Thread 1   │ │ Output Th. 2  │ │ Output Th. 3  │
│ ┌───────────────┐ │ │ ┌───────────┐ │ │ ┌───────────┐ │
│ │ Receiver 1    │ │ │ │Receiver 2 │ │ │ │Receiver 3 │ │
│ │ (subscribed)  │ │ │ │(subscribed│ │ │ │(subscribed│ │
│ └───────────────┘ │ │ └───────────┘ │ │ └───────────┘ │
└───────────────────┘ └───────────────┘ └───────────────┘
```

**Key difference from mpsc:** Each receiver gets a **copy** of every message!

### Key Properties of broadcast

1. **One or more senders** - Can be cloned with `Arc`
2. **Multiple receivers** - Call `subscribe()` to create new ones
3. **Each receiver gets all messages** - True broadcasting!
4. **Lagging receivers** - If one receiver is slow, it might miss messages
5. **Bounded** - Fixed size circular buffer (you set the capacity)

---

## 📚 Part 4: Your Complete Channel Architecture

Let's trace a Bitcoin price through both channels:

```
1. PolygonInput receives Bitcoin price
   ↓
   [InputHandle.send_event()]
   ↓
   mpsc::Sender → Queue → mpsc::Receiver
   ↓
2. Engine receives via mpsc
   ↓
   [Processes with Worker]
   ↓
   [broadcast::Sender.send()]
   ↓
   Broadcast Channel (clones to all receivers)
   ↓
   ┌─────────┬─────────┬─────────┐
   ↓         ↓         ↓         ↓
3. Output 1  Output 2  Output 3  ...
   WebSocket Console   Database  (etc)
```

### Code Flow in Your System

**Step 1: Input sends event** (`polygon_input.rs`):
```rust
let event = Event {
    task: "bitcoin_price".to_string(),
    payload: trade.p as i32,
};
handle.send_event(event).unwrap();  // Goes through mpsc!
```

**Step 2: Engine receives** (`lib.rs`):
```rust
loop {
    match self.input_receiver.recv() {  // Blocks here (mpsc)
        Ok(event) => {
            // Process event...
            let json = serde_json::to_string(&event).unwrap();
            let _ = self.broadcast_sender.send(json);  // Broadcast!
        }
        Err(_) => break,  // All inputs closed
    }
}
```

**Step 3: Output receives** (`websocket_output.rs`):
```rust
let mut rx = handle.subscribe();  // Creates broadcast receiver
while let Ok(processed_event) = rx.recv().await {  // Gets broadcasts!
    // Send to WebSocket client
}
```

---

## 📚 Part 5: mpsc vs broadcast - When to Use Each?

| Feature | mpsc | broadcast |
|---------|------|-----------|
| **Senders** | Multiple (clone sender) | Multiple (use Arc) |
| **Receivers** | Single (cannot clone) | Multiple (subscribe) |
| **Message delivery** | One receiver gets it | All receivers get it |
| **Use case** | Many → One | One → Many |
| **Your usage** | Inputs → Engine | Engine → Outputs |
| **Ordering** | FIFO guaranteed | Per-receiver FIFO |
| **Performance** | Very fast | Fast, but copies data |

---

## 🧪 Exercise 1: Exploring mpsc Basics

Create `src/bin/mpsc_test.rs`:

```rust
use std::sync::mpsc;
use std::thread;
use std::time::Duration;

fn main() {
    println!("=== mpsc Channel Basics ===\n");

    // Create a channel
    let (sender, receiver) = mpsc::channel();

    // Spawn a thread that sends messages
    thread::spawn(move || {
        let messages = vec![
            "Hello",
            "from",
            "another",
            "thread!",
        ];

        for msg in messages {
            println!("Sending: {}", msg);
            sender.send(msg).unwrap();
            thread::sleep(Duration::from_millis(500));
        }
        println!("Sender done!");
    });

    // Receive messages in main thread
    println!("Waiting for messages...\n");
    for received in receiver {
        println!("Received: {}", received);
    }

    println!("\n=== Observations ===");
    println!("1. Did recv() block until messages arrived?");
    println!("2. What happened when the sender was dropped?");
    println!("3. Is the order preserved?");
}
```

Run: `cargo run --bin mpsc_test`

**Questions:**
1. What happens to the main thread while waiting for messages?
2. How does the loop know when to stop?
3. Try removing the `sleep` - does it still work?

---

## 🧪 Exercise 2: Multiple Senders (mpsc)

Create `src/bin/mpsc_multi.rs`:

```rust
use std::sync::mpsc;
use std::thread;
use std::time::Duration;

fn main() {
    println!("=== Multiple Senders with mpsc ===\n");

    let (sender, receiver) = mpsc::channel();

    // Spawn 3 sender threads
    for i in 1..=3 {
        let sender_clone = sender.clone();
        thread::spawn(move || {
            for j in 1..=3 {
                let msg = format!("Thread {} - Message {}", i, j);
                println!("Sending: {}", msg);
                sender_clone.send(msg).unwrap();
                thread::sleep(Duration::from_millis(100 * i as u64));
            }
        });
    }

    // Drop original sender (important!)
    drop(sender);

    // Receive all messages
    println!("Receiving...\n");
    for received in receiver {
        println!("Got: {}", received);
    }

    println!("\n=== Observations ===");
    println!("1. Why did we need to drop the original sender?");
    println!("2. Are messages interleaved from different threads?");
    println!("3. What guarantees the receiver loop ends?");
}
```

Run: `cargo run --bin mpsc_multi`

**Challenge:** What happens if you don't `drop(sender)`?

---

## 🧪 Exercise 3: Exploring broadcast Basics

Create `src/bin/broadcast_test.rs`:

```rust
use tokio::sync::broadcast;
use tokio::time::{sleep, Duration};

#[tokio::main]
async fn main() {
    println!("=== broadcast Channel Basics ===\n");

    // Create a broadcast channel with capacity of 10
    let (sender, _) = broadcast::channel(10);

    // Create 3 receivers
    let mut rx1 = sender.subscribe();
    let mut rx2 = sender.subscribe();
    let mut rx3 = sender.subscribe();

    // Spawn receiver tasks
    tokio::spawn(async move {
        while let Ok(msg) = rx1.recv().await {
            println!("Receiver 1 got: {}", msg);
        }
    });

    tokio::spawn(async move {
        while let Ok(msg) = rx2.recv().await {
            println!("Receiver 2 got: {}", msg);
        }
    });

    tokio::spawn(async move {
        while let Ok(msg) = rx3.recv().await {
            println!("Receiver 3 got: {}", msg);
        }
    });

    // Send messages
    for i in 1..=5 {
        let msg = format!("Message {}", i);
        println!("Broadcasting: {}", msg);
        sender.send(msg).unwrap();
        sleep(Duration::from_millis(100)).await;
    }

    sleep(Duration::from_secs(1)).await;

    println!("\n=== Observations ===");
    println!("1. Did all receivers get all messages?");
    println!("2. Can you add a 4th receiver mid-stream?");
    println!("3. What happens to messages sent before subscribe()?");
}
```

Run: `cargo run --bin broadcast_test`

**Questions:**
1. Do all receivers get every message?
2. What's the difference from mpsc?
3. Try subscribing a receiver AFTER sending some messages - what happens?

---

## 🧪 Exercise 4: Understanding Your InputHandle

Let's trace how your code uses mpsc:

Create `src/bin/input_handle_trace.rs`:

```rust
use fund78::{Event, Input, InputHandle};
use std::thread;
use std::time::Duration;

struct SimpleInput {
    name: String,
    count: i32,
}

impl SimpleInput {
    fn new(name: String, count: i32) -> Self {
        SimpleInput { name, count }
    }
}

impl Input for SimpleInput {
    fn start(self: Box<Self>, handle: InputHandle) {
        println!("[{}] Input started!", self.name);

        for i in 1..=self.count {
            let event = Event {
                task: format!("{}_task", self.name),
                payload: i,
            };

            println!("[{}] Sending event #{}", self.name, i);
            handle.send_event(event).unwrap();
            thread::sleep(Duration::from_millis(500));
        }

        println!("[{}] Input finished!", self.name);
    }
}

fn main() {
    println!("=== Understanding InputHandle (mpsc) ===\n");

    // This simulates what Engine does
    use std::sync::mpsc;

    let (sender, receiver) = mpsc::channel();

    // Create InputHandles (like Engine does)
    let handle1 = InputHandle::new(sender.clone());
    let handle2 = InputHandle::new(sender.clone());

    // Simulate spawning inputs (like Engine.run() does)
    let input1: Box<dyn Input> = Box::new(SimpleInput::new("Input1".to_string(), 3));
    let input2: Box<dyn Input> = Box::new(SimpleInput::new("Input2".to_string(), 3));

    thread::spawn(move || input1.start(handle1));
    thread::spawn(move || input2.start(handle2));

    drop(sender);  // Important!

    // Simulate Engine receiving (like Engine.run() does)
    println!("\n[Engine] Receiving events...\n");
    for event in receiver {
        println!("[Engine] Got: task={}, payload={}", event.task, event.payload);
    }

    println!("\n=== Observations ===");
    println!("1. Are events from both inputs interleaved?");
    println!("2. This is exactly what your Engine does!");
    println!("3. The InputHandle hides the mpsc complexity!");
}
```

Run: `cargo run --bin input_handle_trace`

**This is your actual architecture!** See how it works?

---

## 🧪 Exercise 5: Channel Capacity and Blocking

Create `src/bin/channel_capacity.rs`:

```rust
use std::sync::mpsc;
use std::thread;
use std::time::Duration;

fn main() {
    println!("=== Channel Capacity ===\n");

    // Create a bounded channel (capacity of 3)
    let (sender, receiver) = mpsc::sync_channel(3);

    thread::spawn(move || {
        for i in 1..=10 {
            println!("Trying to send {}...", i);
            sender.send(i).unwrap();
            println!("  Sent {}!", i);
        }
    });

    thread::sleep(Duration::from_secs(2));
    println!("\nNow receiving...\n");

    for received in receiver {
        println!("Received: {}", received);
        thread::sleep(Duration::from_millis(500));
    }

    println!("\n=== Observations ===");
    println!("1. Did sending block after 3 messages?");
    println!("2. What happened when we started receiving?");
    println!("3. Try changing capacity to 0 - what happens?");
}
```

Run: `cargo run --bin channel_capacity`

**Note:** Your `mpsc::channel()` is UNBOUNDED - it never blocks the sender!

---

## 🎯 Challenge: Implement a Mini Engine

Create your own mini engine that uses both channel types:

Create `src/bin/mini_engine.rs`:

```rust
use std::sync::{mpsc, Arc};
use std::thread;
use std::time::Duration;
use tokio::sync::broadcast;

fn main() {
    println!("=== Mini Engine Challenge ===\n");

    // TODO: Create an mpsc channel for inputs → engine
    // TODO: Create a broadcast channel for engine → outputs
    // TODO: Spawn 2 input threads that send numbers
    // TODO: Spawn an engine thread that:
    //       - Receives from mpsc
    //       - Multiplies by 2
    //       - Broadcasts result
    // TODO: Spawn 2 output threads that print received values

    println!("Your turn to implement this!");
    println!("Hint: Look at input_handle_trace.rs and broadcast_test.rs");
}
```

**Solution:** Try implementing this yourself first!

---

## 🤔 Questions to Think About

1. Why does `mpsc::Receiver` not implement `Clone`?
2. Why does `broadcast::Sender` need `Arc`?
3. What happens if a broadcast receiver is slow?
4. How does `recv()` block without spinning (using CPU)?
5. Could you use broadcast for Input → Engine? Why or why not?
6. Could you use mpsc for Engine → Output? Why or why not?

---

## 📖 Key Concepts Summary

### mpsc (Multi-Producer, Single-Consumer)
- ✅ Many senders → One receiver
- ✅ Clone the sender, share with threads
- ✅ Receiver cannot be cloned
- ✅ One receiver gets each message
- ✅ Perfect for: Inputs → Engine

### broadcast (Multi-Producer, Multi-Consumer)
- ✅ Senders need Arc to share
- ✅ Call `subscribe()` for new receivers
- ✅ All receivers get all messages
- ✅ Has a capacity limit (circular buffer)
- ✅ Perfect for: Engine → Outputs

### Your Abstractions
- **InputHandle** wraps `mpsc::Sender` - inputs only send
- **OutputHandle** wraps `broadcast::Sender` - outputs only receive
- Users never see the underlying channels!
- Simple, safe API that enforces architecture

---

## 📚 Further Reading

- [The Rust Book - Message Passing](https://doc.rust-lang.org/book/ch16-02-message-passing.html)
- [std::sync::mpsc documentation](https://doc.rust-lang.org/std/sync/mpsc/)
- [tokio::sync::broadcast documentation](https://docs.rs/tokio/latest/tokio/sync/broadcast/)
- [Tokio Tutorial - Channels](https://tokio.rs/tokio/tutorial/channels)

---

## 📝 Notes & Observations

Use this space to write down your insights as you work through the exercises:

```
Your notes here...




```
