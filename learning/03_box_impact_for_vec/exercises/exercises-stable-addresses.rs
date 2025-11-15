// Exercise 3: Stable Memory Addresses
// Copy this file to: src/bin/stable_addresses.rs
// Run with: cargo run --bin stable_addresses

fn main() {
    println!("=== Memory Address Stability ===\n");

    struct Task {
        id: u32,
    }

    // Test 1: Vec<T> - addresses can change
    println!("Test 1: Vec<Task>");
    let mut tasks1: Vec<Task> = Vec::with_capacity(2);
    tasks1.push(Task { id: 1 });

    let addr_before = &tasks1[0] as *const Task;
    println!("  Before resize: {:p}", addr_before);

    // Force reallocation
    for i in 2..10 {
        tasks1.push(Task { id: i });
    }

    let addr_after = &tasks1[0] as *const Task;
    println!("  After resize:  {:p}", addr_after);
    println!("  Same address? {}\n", addr_before == addr_after);

    // Test 2: Vec<Box<T>> - addresses stay stable
    println!("Test 2: Vec<Box<Task>>");
    let mut tasks2: Vec<Box<Task>> = Vec::with_capacity(2);
    tasks2.push(Box::new(Task { id: 1 }));

    let addr_before = &**tasks2[0] as *const Task;
    println!("  Before resize: {:p}", addr_before);

    // Force reallocation
    for i in 2..10 {
        tasks2.push(Box::new(Task { id: i }));
    }

    let addr_after = &**tasks2[0] as *const Task;
    println!("  After resize:  {:p}", addr_after);
    println!("  Same address? {}\n", addr_before == addr_after);
}
