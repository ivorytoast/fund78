// Exercise 1: Exploring Vec Basics
// Copy this file to: src/bin/vec_test.rs
// Run with: cargo run --bin vec_test

fn main() {
    println!("=== Vec Basics ===");

    let mut numbers: Vec<i32> = vec![1, 2, 3];

    println!("Initial: {:?}", numbers);
    println!("Length: {}", numbers.len());
    println!("Capacity: {}", numbers.capacity());
    println!("Size in memory: {} bytes", std::mem::size_of_val(&numbers));

    // Add elements and watch capacity grow
    for i in 4..=10 {
        numbers.push(i);
        println!("After push({}): len={}, cap={}", i, numbers.len(), numbers.capacity());
    }

    println!("\n=== Your Observations ===");
    println!("1. What is the initial capacity?");
    println!("2. When does the capacity increase?");
    println!("3. How much does it increase by?");
    println!("4. What is the size of the Vec struct itself?");
}
