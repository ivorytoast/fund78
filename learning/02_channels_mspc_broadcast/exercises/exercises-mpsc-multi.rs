// Exercise 2: Multiple Senders (mpsc)
// Copy this file to: src/bin/mpsc_multi.rs
// Run with: cargo run --bin mpsc_multi

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
