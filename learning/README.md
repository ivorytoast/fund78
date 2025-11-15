# Learning Resources for Fund78

This folder contains lessons and exercises to help you understand the Rust concepts used in Fund78.

## 📚 Lessons

Each lesson is a standalone markdown file with:
- Conceptual explanations
- Code examples
- Memory diagrams
- Hands-on exercises
- Questions to think about

### Available Lessons

1. **[lesson-trait-box-vec.md](./lesson-trait-box-vec.md)** - Understanding Traits, Box, and Vec
   - What are traits and why use them?
   - How `Vec<T>` works and memory layout
   - What `Box<T>` does and when to use it
   - Why `Vec<Box<dyn Trait>>` is necessary
   - Dynamic dispatch with `dyn`

2. **[lesson-channels-mpsc-broadcast.md](./lesson-channels-mpsc-broadcast.md)** - Understanding Channels
   - What are channels and why use them for thread communication?
   - `mpsc` channels (Multi-Producer, Single-Consumer)
   - `broadcast` channels (Multi-Producer, Multi-Consumer)
   - How `InputHandle` wraps `mpsc::Sender`
   - How `OutputHandle` wraps `broadcast::Sender`
   - Complete data flow through your system

3. **[lesson-understanding-box-impact-for-vec.md](./lesson-understanding-box-impact-for-vec.md)** - Understanding Box Impact for Vec
   - `Vec<T>` vs `Vec<Box<T>>` vs `Vec<Box<dyn Trait>>`
   - When to use each pattern and why
   - Java interfaces vs Rust traits (`dyn` keyword)
   - Memory layouts and performance trade-offs
   - Recursive types and stable addresses
   - Static vs dynamic dispatch

## 🧪 Running Exercises

Most lessons include code exercises. To run them:

1. Create the exercise file in `src/bin/` (examples provided in lessons)
2. Run with: `cargo run --bin <filename_without_rs>`
3. For compile errors: `cargo build --bin <filename_without_rs>`

Example:
```bash
cargo run --bin vec_test
cargo run --bin box_test
cargo run --bin worker_size
```

## 📝 How to Use This

1. Read a lesson file
2. Try the exercises hands-on
3. Write your observations in the "Notes" section at the bottom
4. Experiment with variations
5. Come back later to review!

## 🎯 Learning Path

Suggested order:
1. **Traits, Box, and Vec** (start here!) - Understand the type system
2. **Channels (mpsc & broadcast)** - Learn how threads communicate
3. **Understanding Box Impact for Vec** - When to use Box and why
4. [More lessons to come...]

---

Happy learning! 🦀
