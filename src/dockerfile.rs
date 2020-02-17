use lazy_static::lazy_static;
use regex::{Regex, RegexSetBuilder};
use std::{collections::HashSet, error::Error, fs, path};

lazy_static! {
    static ref FROM_DIRECTIVE: regex::Regex =
        Regex::new(r"(?i)^FROM\s(?P<image>[^\s]+)?(?:\sAS\s)?(?P<stage>[\w+]+)?").unwrap();
    static ref VERSION_REGEX: &'static str = r"v?([0-9]+(?:(?:\.[a-z0-9]+)|(?:-(?:kb)?[0-9]+))*)";
    static ref TAG_REGEX: regex::RegexSet = RegexSetBuilder::new(&[
        format!(r"{}(-[a-z0-9.\-]+)?", *VERSION_REGEX),
        format!(r"([a-z0-9.\-]+-)?{}", *VERSION_REGEX),
        format!(r"([a-z\-]+-)?{}(-[a-z\-]+)?", *VERSION_REGEX),
    ])
    .case_insensitive(true)
    .build()
    .unwrap();
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
            .filter_map(|line| line.name("stage"))
            .map(|stage| stage.as_str())
            .collect();

        let images: Vec<_> = from
            .iter()
            .filter_map(|line| line.name("image"))
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
        let invalid_images: HashSet<_> = self.validate_image_tags();

        ValidationResult {
            invalid_images: invalid_images.into_iter().collect(),
        }
    }

    fn validate_image_tags(&self) -> HashSet<Image> {
        self.images()
            .into_iter()
            .filter(|image| match &image.tag {
                Some(tag) => !TAG_REGEX.is_match(tag.as_str()),
                None => true,
            })
            .collect()
    }
}

#[derive(Hash, Eq, PartialEq)]
pub struct Image {
    pub repository: String,
    pub tag: Option<String>,
}

pub struct ValidationResult {
    pub invalid_images: Vec<Image>,
}
