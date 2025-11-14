use std::collections::VecDeque;
use std::io::Write;

use fund78::create_or_overwrite_file;

fn main() {
    let mut file = match create_or_overwrite_file("out.log") {
        Ok(f) => f,
        Err(e) => {
            eprintln!("Failed to create or open file: {}", e);
            return;
        }
    };

    let mut tasks = VecDeque::new();
    tasks.push_back(10);
    tasks.push_back(20);
    tasks.push_back(30);

    while let Some(task) = tasks.pop_back() {
        if let Err(e) = writeln!(file, "{}", task) {
            eprintln!("Failed to write to file: {}", e);
            return;
        }
    }

    println!("Tasks written successfully!")
}
