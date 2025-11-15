// Exercise 4: Understanding Your InputHandle
// Copy this file to: src/bin/input_handle_trace.rs
// Run with: cargo run --bin input_handle_trace

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
