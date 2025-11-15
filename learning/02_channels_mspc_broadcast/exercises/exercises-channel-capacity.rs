// Exercise 5: Channel Capacity and Blocking
// Copy this file to: src/bin/channel_capacity.rs
// Run with: cargo run --bin channel_capacity

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
