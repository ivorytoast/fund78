// Exercise 3: Exploring broadcast Basics
// Copy this file to: src/bin/broadcast_test.rs
// Run with: cargo run --bin broadcast_test

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
