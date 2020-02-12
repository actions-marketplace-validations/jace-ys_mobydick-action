use std::{error::Error, fs, path};

pub struct Dockerfile {
    pub path: String,
    pub content: String,
}

pub struct Image {
    pub name: String,
    pub tag: String,
}

pub struct ValidationResult {
    pub invalid_images: Vec<Image>,
}

impl Dockerfile {
    pub fn new(path: path::PathBuf) -> Result<Self, Box<dyn Error>> {
        let dockerfile = Self {
            path: String::from(path.to_str().unwrap()),
            content: fs::read_to_string(path)?,
        };
        Ok(dockerfile)
    }

    pub fn validate(&self) -> ValidationResult {
        ValidationResult {
            invalid_images: vec![],
        }
    }
}
