// Exercise 1: Exploring mpsc Basics
// Copy this file to: src/bin/mpsc_test.rs
// Run with: cargo run --bin mpsc_test

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
