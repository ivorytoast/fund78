use std::fs::{File, OpenOptions};
use std::io::{ErrorKind, Result};

pub fn create_or_open_file(file_name: &str) -> Result<File> {
    match File::create_new(file_name) {
        Ok(f) => Ok(f),
        Err(e) => {
            if e.kind() != ErrorKind::AlreadyExists {
                return Err(e);
            }

            OpenOptions::new().write(true).append(true).open(file_name)
        }
    }
}

pub fn create_or_overwrite_file(file_name: &str) -> Result<File> {
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

        let file = create_or_open_file(&path).expect("Failed to create file");
        assert!(file.metadata().is_ok());

        fs::remove_file(&path).unwrap();
    }
}
