// Exercise 1: Comparing Vec Patterns
// Copy this file to: src/bin/box_comparison.rs
// Run with: cargo run --bin box_comparison

fn main() {
    println!("=== Vec<T> vs Vec<Box<T>> ===\n");

    // Small type (like your BitcoinPriceWorker)
    #[derive(Clone)]
    struct SmallWorker {
        id: u32,
    }

    let small_direct: Vec<SmallWorker> = vec![
        SmallWorker { id: 1 },
        SmallWorker { id: 2 },
    ];

    let small_boxed: Vec<Box<SmallWorker>> = vec![
        Box::new(SmallWorker { id: 1 }),
        Box::new(SmallWorker { id: 2 }),
    ];

    println!("Small type (4 bytes each):");
    println!("  Vec<SmallWorker>:      {} bytes",
        std::mem::size_of_val(&small_direct));
    println!("  Vec<Box<SmallWorker>>: {} bytes",
        std::mem::size_of_val(&small_boxed));
    println!("  → Box adds overhead for small types!\n");

    // HUGE type
    #[derive(Clone)]
    struct HugeWorker {
        data: [u8; 10000],  // 10KB of data!
    }

    let huge_direct: Vec<HugeWorker> = vec![
        HugeWorker { data: [0; 10000] },
        HugeWorker { data: [0; 10000] },
    ];

    let huge_boxed: Vec<Box<HugeWorker>> = vec![
        Box::new(HugeWorker { data: [0; 10000] }),
        Box::new(HugeWorker { data: [0; 10000] }),
    ];

    println!("Huge type (10KB each):");
    println!("  Vec<HugeWorker>:      {} bytes",
        std::mem::size_of_val(&huge_direct));
    println!("  Vec<Box<HugeWorker>>: {} bytes",
        std::mem::size_of_val(&huge_boxed));
    println!("  → Same Vec size! But Box makes growing faster\n");

    println!("=== When to use Box<T>: ===");
    println!("1. Type is VERY large (>1KB)");
    println!("2. Need stable memory addresses");
    println!("3. Recursive types (type contains itself)");
    println!("4. Avoiding expensive moves");
}
