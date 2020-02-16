use lazy_static::lazy_static;
use regex::Regex;
use std::{collections::HashSet, error::Error, fs, path};

lazy_static! {
    static ref FROM_DIRECTIVE: regex::Regex =
        Regex::new(r"(?i)^FROM\s([^\s]+)?(?:\sAS\s)?([\w+]+)?").unwrap();
}

pub struct Dockerfile {
    pub path: String,
    pub content: String,
}

impl Dockerfile {
    pub fn new(path: path::PathBuf) -> Result<Self, Box<dyn Error>> {
        let dockerfile = Self {
            path: String::from(path.to_str().unwrap()),
            content: fs::read_to_string(path)?,
        };
        Ok(dockerfile)
    }

    fn images(&self) -> Vec<Image> {
        let from: Vec<_> = self
            .content
            .lines()
            .filter_map(|line| FROM_DIRECTIVE.captures(line.trim()))
            .collect();

        let stages: HashSet<_> = from
            .iter()
            .filter_map(|line| line.get(2))
            .map(|stage| stage.as_str())
            .collect();

        let images = from
            .iter()
            .filter_map(|line| line.get(1))
            .filter(|image| stages.get(image.as_str()).is_none())
            .map(|image| self.parse_image_name(image.as_str()))
            .collect();

        images
    }

    fn parse_image_name(&self, image_name: &str) -> Image {
        let image: Vec<_> = image_name.splitn(2, ':').collect();
        Image {
            repository: String::from(image[0]),
            tag: match image.len() {
                1 => None,
                _ => Some(String::from(image[1])),
            },
        }
    }

    pub fn validate(&self) -> ValidationResult {
        ValidationResult {
            invalid_images: self.images(),
        }
    }
}

pub struct Image {
    pub repository: String,
    pub tag: Option<String>,
}

pub struct ValidationResult {
    pub invalid_images: Vec<Image>,
}
