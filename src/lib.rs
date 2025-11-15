pub mod sample;

use serde::{Deserialize, Serialize};
use serde_json;
use std::collections::VecDeque;
use std::fs::File;
use std::io::Result;
use std::io::Write;

#[derive(Deserialize, Serialize, Debug)]
pub struct Event {
    pub task: String,
    pub payload: i32,
}

pub struct Worker {
    pub handles_task: String,
    pub job: fn(i32) -> Event,
}

pub fn engine_process(mut events: VecDeque<Event>, workers: Vec<Worker>) {
    let in_file_result = create_or_overwrite_file("in.log");

    let in_file = match in_file_result {
        Ok(f) => f,
        Err(e) => {
            eprintln!("Failed to create or open IN file: {}", e);
            return;
        }
    };

    let out_file_result = create_or_overwrite_file("out.log");

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
            if worker.handles_task == task.task {
                let event = (worker.job)(task.payload);
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

fn create_or_overwrite_file(file_name: &str) -> Result<File> {
    match File::create(file_name) {
        Ok(f) => Ok(f),
        Err(e) => {
            return Err(e);
        }
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

        let file = create_or_overwrite_file(&path).expect("Failed to create file");
        assert!(file.metadata().is_ok());

        fs::remove_file(&path).unwrap();
    }
}
