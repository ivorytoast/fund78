// Exercise 2: Understanding Box Size
// Copy this file to: src/bin/box_test.rs
// Run with: cargo run --bin box_test

fn main() {
    println!("=== Box Sizes ===");

    // Different sized types
    let small: Box<u8> = Box::new(42);           // 1 byte of data
    let medium: Box<[u8; 100]> = Box::new([0; 100]); // 100 bytes of data
    let large: Box<[u8; 10000]> = Box::new([0; 10000]); // 10KB of data

    println!("Box<u8> size: {} bytes", std::mem::size_of_val(&small));
    println!("Box<[u8; 100]> size: {} bytes", std::mem::size_of_val(&medium));
    println!("Box<[u8; 10000]> size: {} bytes", std::mem::size_of_val(&large));

    println!("\nThe actual data sizes:");
    println!("u8 size: {} bytes", std::mem::size_of::<u8>());
    println!("[u8; 100] size: {} bytes", std::mem::size_of::<[u8; 100]>());
    println!("[u8; 10000] size: {} bytes", std::mem::size_of::<[u8; 10000]>());

    println!("\n=== Your Observations ===");
    println!("1. Are all the Box sizes the same?");
    println!("2. What does this tell you about Box?");
    println!("3. Why is this important for Vec<Box<dyn Trait>>?");
}
