use std::path;

pub struct Dockerfile {
    pub path: String,
    pub content: String,
}

impl Dockerfile {
    pub fn new(path: path::PathBuf) -> Self {
        Self {
            path: String::from(path.to_str().unwrap()),
            content: Self::read(path),
        }
    }

    pub fn validate(&self) -> bool {
        true
    }

    fn read(path: path::PathBuf) -> String {
        String::from("content")
    }
}
