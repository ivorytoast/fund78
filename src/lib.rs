use serde::{Deserialize, Serialize};
use serde_json;
use std::collections::VecDeque;
use std::fs::File;
use std::fs::OpenOptions;
use std::io::Result;
use std::io::Write;
use std::sync::{Arc, mpsc};
use std::time::Duration;
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

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RequestStats {
    pub count: u64,
    #[serde(serialize_with = "serialize_duration")]
    pub total_duration: Duration,
    #[serde(serialize_with = "serialize_option_duration")]
    pub min_duration: Option<Duration>,
    #[serde(serialize_with = "serialize_option_duration")]
    pub max_duration: Option<Duration>,
    #[serde(serialize_with = "serialize_option_duration")]
    pub average_duration: Option<Duration>,
}

fn serialize_duration<S>(duration: &Duration, serializer: S) -> std::result::Result<S::Ok, S::Error>
where
    S: serde::Serializer,
{
    serializer.serialize_u128(duration.as_micros())
}

fn serialize_option_duration<S>(
    duration: &Option<Duration>,
    serializer: S,
) -> std::result::Result<S::Ok, S::Error>
where
    S: serde::Serializer,
{
    match duration {
        Some(d) => serializer.serialize_some(&d.as_micros()),
        None => serializer.serialize_none(),
    }
}

impl RequestStats {
    fn new() -> Self {
        RequestStats {
            count: 0,
            total_duration: Duration::ZERO,
            min_duration: None,
            max_duration: None,
            average_duration: None,
        }
    }

    fn update(&mut self, duration: Duration) {
        self.count += 1;
        self.total_duration += duration;

        self.min_duration = Some(match self.min_duration {
            Some(min) => min.min(duration),
            None => duration,
        });

        self.max_duration = Some(match self.max_duration {
            Some(max) => max.max(duration),
            None => duration,
        });

        // Update the average
        self.average_duration = Some(self.total_duration / self.count as u32);
    }

    fn average(&self) -> Option<Duration> {
        self.average_duration
    }

    fn print_stats(&self) {
        println!("\n=== Request Statistics ===");
        println!("Total requests: {}", self.count);
        if let Some(avg) = self.average_duration {
            println!("Average time: {:?}", avg);
        }
        if let Some(min) = self.min_duration {
            println!("Fastest request: {:?}", min);
        }
        if let Some(max) = self.max_duration {
            println!("Slowest request: {:?}", max);
        }
        println!("========================\n");
    }

    fn to_json(&self) -> String {
        serde_json::json!({
            "type": "stats",
            "data": self
        })
        .to_string()
    }
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
    stats: RequestStats,
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
            stats: RequestStats::new(),
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
            mut stats,
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
                        process_events(
                            &mut in_file,
                            &mut out_file,
                            &workers,
                            pending_events.clone(),
                            &mut stats,
                        );

                        // Broadcast events to all outputs
                        for event in &pending_events {
                            let json = serde_json::to_string(&event).unwrap();
                            let _ = broadcast_sender.send(json);
                        }

                        // Broadcast stats update to all outputs
                        let stats_json = stats.to_json();
                        let _ = broadcast_sender.send(stats_json);

                        pending_events.clear();
                    }
                    Err(_) => {
                        println!("All inputs closed, shutting down engine");
                        stats.print_stats();
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
        process_events(
            &mut self.in_file,
            &mut self.out_file,
            &self.workers,
            events,
            &mut self.stats,
        );
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
    stats: &mut RequestStats,
) {
    while let Some(task) = events.pop_back() {
        let start = std::time::Instant::now();

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

        let duration = start.elapsed();
        stats.update(duration);

        println!(
            "Request '{}' took {:?} | Avg: {:?} | Min: {:?} | Max: {:?} | Count: {}",
            task.task,
            duration,
            stats.average().unwrap_or(Duration::ZERO),
            stats.min_duration.unwrap_or(Duration::ZERO),
            stats.max_duration.unwrap_or(Duration::ZERO),
            stats.count
        );
    }
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
