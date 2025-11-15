// Exercise 3: Exploring Your Workers Vec
// Copy this file to: src/bin/worker_size.rs
// Run with: cargo run --bin worker_size

use fund78::{BitcoinPriceWorker, Worker};

fn main() {
    println!("=== Worker Sizes ===");

    // Create workers vec like in main.rs
    let workers: Vec<Box<dyn Worker>> = vec![
        Box::new(BitcoinPriceWorker::new()),
    ];

    println!("Vec<Box<dyn Worker>> size: {} bytes",
        std::mem::size_of_val(&workers));

    println!("Box<dyn Worker> size: {} bytes",
        std::mem::size_of::<Box<dyn Worker>>());

    println!("BitcoinPriceWorker size: {} bytes",
        std::mem::size_of::<BitcoinPriceWorker>());

    println!("\nVec details:");
    println!("  Length: {}", workers.len());
    println!("  Capacity: {}", workers.capacity());

    println!("\n=== Your Observations ===");
    println!("1. How big is the Vec itself?");
    println!("2. How big is each Box<dyn Worker>?");
    println!("3. How big is a BitcoinPriceWorker?");
    println!("4. Try drawing the memory layout!");
}
