// Exercise 5: Dynamic Dispatch (dyn)
// Copy this file to: src/bin/dyn_vs_static.rs
// Run with: cargo run --bin dyn_vs_static

use fund78::Worker;

// Static dispatch (compile-time, fast)
fn process_static<T: Worker>(worker: T) {
    println!("[Static] Task: {}", worker.handles_task());
    // Compiler knows exact type - can inline!
}

// Dynamic dispatch (runtime, flexible)
fn process_dynamic(worker: Box<dyn Worker>) {
    println!("[Dynamic] Task: {}", worker.handles_task());
    // Runtime lookup via vtable
}

fn main() {
    use fund78::BitcoinPriceWorker;

    println!("=== Static vs Dynamic Dispatch ===\n");

    let worker1 = BitcoinPriceWorker::new();
    let worker2 = BitcoinPriceWorker::new();

    // Static dispatch (like Java generics: <T extends Worker>)
    process_static(worker1);

    // Dynamic dispatch (like Java: Worker w = new BitcoinWorker())
    process_dynamic(Box::new(worker2));

    println!("\nKey difference:");
    println!("  Static:  Compiler knows exact type → fast, can inline");
    println!("  Dynamic: Runtime lookup via vtable → flexible, slight overhead");

    println!("\nYour Vec<Box<dyn Worker>> uses dynamic dispatch");
    println!("because you need different types in same Vec!");
}
