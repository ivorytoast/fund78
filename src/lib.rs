use serde::{Deserialize, Serialize};
use serde_json;
use std::collections::VecDeque;
use std::fs::File;
use std::fs::OpenOptions;
use std::io::Result;
use std::io::Write;
use std::sync::{mpsc, Arc};
use tokio::sync::broadcast;

pub mod input;
pub mod output;
pub mod worker;

pub use input::{Input, InputHandle, PolygonInput};
pub use output::{Output, OutputHandle, WebSocketOutput};
pub use worker::{BitcoinPriceWorker, FunctionWorker, Worker};

#[derive(Deserialize, Serialize, Debug, Clone)]
pub struct Event {
    pub task: String,
    pub payload: i32,
}

pub struct Engine {
    in_file: File,
    out_file: File,
    workers: Vec<Box<dyn Worker>>,
    inputs: Vec<Box<dyn Input>>,
    outputs: Vec<Box<dyn Output>>,
    input_sender: mpsc::Sender<Event>,
    input_receiver: mpsc::Receiver<Event>,
    broadcast_sender: Arc<broadcast::Sender<String>>,
}

impl Engine {
    pub fn new(
        workers: Vec<Box<dyn Worker>>,
        inputs: Vec<Box<dyn Input>>,
        outputs: Vec<Box<dyn Output>>,
    ) -> Result<Self> {
        let in_file = create_or_append_file("in.log")?;
        let out_file = create_or_append_file("out.log")?;

        let (input_sender, input_receiver) = mpsc::channel();
        let (broadcast_sender, _) = broadcast::channel(100);

        Ok(Engine {
            in_file,
            out_file,
            workers,
            inputs,
            outputs,
            input_sender,
            input_receiver,
            broadcast_sender: Arc::new(broadcast_sender),
        })
    }

    /// Create a handle for an input to send events to the engine
    /// This is the ONLY way to get a handle to send events
    pub fn create_input_handle(&self) -> InputHandle {
        InputHandle::new(self.input_sender.clone())
    }

    /// Create a handle for an output to receive events from the engine
    /// This is the ONLY way to get a handle to receive events
    pub fn create_output_handle(&self) -> OutputHandle {
        OutputHandle::new(self.broadcast_sender.clone())
    }

    /// Run the engine's main processing loop
    /// This consumes the engine, spawns all input/output threads, and manages their lifecycle
    /// Blocks until all threads complete
    pub fn run(self) {
        use std::thread;

        println!("Starting engine...");

        // Extract the components we need before moving self
        let Engine {
            mut in_file,
            mut out_file,
            workers,
            inputs,
            outputs,
            input_sender,
            input_receiver,
            broadcast_sender,
        } = self;

        // Spawn a thread for each input
        let input_handles: Vec<_> = inputs
            .into_iter()
            .enumerate()
            .map(|(i, input)| {
                let handle = InputHandle::new(input_sender.clone());
                thread::spawn(move || {
                    println!("Input {} started", i);
                    input.start(handle);
                    println!("Input {} finished", i);
                })
            })
            .collect();

        // Spawn a thread for each output
        let output_handles: Vec<_> = outputs
            .into_iter()
            .enumerate()
            .map(|(i, output)| {
                let handle = OutputHandle::new(broadcast_sender.clone());
                thread::spawn(move || {
                    println!("Output {} started", i);
                    output.start(handle);
                    println!("Output {} finished", i);
                })
            })
            .collect();

        // Run the engine processing loop in a separate thread
        let engine_handle = thread::spawn(move || {
            println!("Engine processing loop started");
            let mut pending_events = VecDeque::new();

            loop {
                match input_receiver.recv() {
                    Ok(event) => {
                        pending_events.push_back(event.clone());

                        // Drain any additional events that are immediately available
                        while let Ok(event) = input_receiver.try_recv() {
                            pending_events.push_back(event);
                        }

                        // Process all pending events
                        process_events(&mut in_file, &mut out_file, &workers, pending_events.clone());

                        // Broadcast events to all outputs
                        for event in &pending_events {
                            let json = serde_json::to_string(&event).unwrap();
                            let _ = broadcast_sender.send(json);
                        }

                        pending_events.clear();
                    }
                    Err(_) => {
                        println!("All inputs closed, shutting down engine");
                        break;
                    }
                }
            }
        });

        println!("All systems running...");

        // Wait for all threads to complete
        for (i, handle) in input_handles.into_iter().enumerate() {
            if let Err(e) = handle.join() {
                eprintln!("Input {} panicked: {:?}", i, e);
            }
        }

        // Wait for engine to finish processing
        if let Err(e) = engine_handle.join() {
            eprintln!("Engine panicked: {:?}", e);
        }

        // Wait for all outputs to complete
        for (i, handle) in output_handles.into_iter().enumerate() {
            if let Err(e) = handle.join() {
                eprintln!("Output {} panicked: {:?}", i, e);
            }
        }

        println!("Engine shutdown complete");
    }

    pub fn process(&mut self, events: VecDeque<Event>) {
        process_events(&mut self.in_file, &mut self.out_file, &self.workers, events);
    }
}

pub fn engine_process(mut events: VecDeque<Event>, workers: Vec<Box<dyn Worker>>) {
    let in_file_result = create_or_append_file("in.log");
    let in_file = match in_file_result {
        Ok(f) => f,
        Err(e) => {
            eprintln!("Failed to create or open IN file: {}", e);
            return;
        }
    };
    let out_file_result = create_or_append_file("out.log");
    let out_file = match out_file_result {
        Ok(f) => f,
        Err(e) => {
            eprintln!("Failed to create or open OUT file: {}", e);
            return;
        }
    };

    while let Some(task) = events.pop_back() {
        if let Err(e) = serde_json::to_writer(&in_file, &task) {
            eprintln!("Failed to write to IN file: {}", e);
            return;
        }

        if let Err(e) = writeln!(&in_file) {
            eprintln!("Failed to write IN newline: {}", e);
            return;
        }

        for worker in &workers {
            if worker.handles_task() == task.task {
                let event = worker.process(task.clone());
                if let Err(e) = serde_json::to_writer(&out_file, &event) {
                    eprintln!("Failed to write to file: {}", e);
                    return;
                };

                if let Err(e) = writeln!(&out_file) {
                    eprintln!("Failed to write newline: {}", e);
                    return;
                }
            }
        }
    }

    println!("Events written successfully!")
}

fn create_or_append_file(file_name: &str) -> Result<File> {
    match OpenOptions::new().create(true).append(true).open(file_name) {
        Ok(f) => Ok(f),
        Err(e) => Err(e),
    }
}

/// Helper function to process events (used by both Engine::process and Engine::run)
fn process_events(
    in_file: &mut File,
    out_file: &mut File,
    workers: &[Box<dyn Worker>],
    mut events: VecDeque<Event>,
) {
    while let Some(task) = events.pop_back() {
        if let Err(e) = serde_json::to_writer(&mut *in_file, &task) {
            eprintln!("Failed to write to IN file: {}", e);
            return;
        }
        if let Err(e) = writeln!(&mut *in_file) {
            eprintln!("Failed to write IN newline: {}", e);
            return;
        }

        for worker in workers {
            if worker.handles_task() == task.task {
                let event = worker.process(task.clone());
                if let Err(e) = serde_json::to_writer(&mut *out_file, &event) {
                    eprintln!("Failed to write to file: {}", e);
                    return;
                };
                if let Err(e) = writeln!(&mut *out_file) {
                    eprintln!("Failed to write newline: {}", e);
                    return;
                }
            }
        }
    }
    println!("Events written successfully!")
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;
    use tempfile::NamedTempFile;

    #[test]
    fn test_create_or_open_file_new() {
        let temp = NamedTempFile::new().unwrap();
        let path = temp.path().to_str().unwrap().to_string();
        drop(temp);
        let file = create_or_append_file(&path).expect("Failed to create file");
        assert!(file.metadata().is_ok());

        fs::remove_file(&path).unwrap();
    }
}
