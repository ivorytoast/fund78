// Exercise 2: Check Your Worker Size
// Copy this file to: src/bin/check_worker_size.rs
// Run with: cargo run --bin check_worker_size

use fund78::BitcoinPriceWorker;

fn main() {
    let size = std::mem::size_of::<BitcoinPriceWorker>();

    println!("BitcoinPriceWorker size: {} bytes\n", size);

    if size < 100 {
        println!("→ Small type! Vec<BitcoinPriceWorker> is fine");
    } else if size < 1024 {
        println!("→ Medium type. Vec<BitcoinPriceWorker> is still OK");
    } else {
        println!("→ Large type! Vec<Box<BitcoinPriceWorker>> might help");
    }

    println!("\n=== Why you use Box<dyn Worker> ===");
    println!("NOT because workers are large,");
    println!("but because you want DIFFERENT worker types in same Vec!");
    println!("\nVec<Box<dyn Worker>> enables:");
    println!("  - BitcoinPriceWorker");
    println!("  - EthereumPriceWorker");
    println!("  - StockPriceWorker");
    println!("  All in the SAME Vec!");
}
