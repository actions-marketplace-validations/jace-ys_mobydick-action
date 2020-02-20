use lazy_static::lazy_static;
use regex::{Regex, RegexSetBuilder};
use std::{collections::HashSet, error::Error, fs, path::PathBuf};

lazy_static! {
    static ref FROM_DIRECTIVE: regex::Regex =
        Regex::new(r"(?i)^FROM\s(?P<image>[^\s]+)?(?:\sAS\s)?(?P<stage>[\w+]+)?").unwrap();
    static ref VERSION_REGEX: &'static str = r"v?([0-9]+(?:(?:\.[a-z0-9]+)|(?:-(?:kb)?[0-9]+))*)";
    static ref TAG_REGEX: regex::RegexSet = RegexSetBuilder::new(&[
        format!(r"^{}(?P<suffix>-[a-z0-9.\-]+)?$", *VERSION_REGEX),
        format!(r"^(?P<prefix>[a-z0-9.\-]+-)?{}$", *VERSION_REGEX),
        format!(
            r"^(?P<prefix>[a-z\-]+-)?{}(?P<suffix>-[a-z\-]+)?$",
            *VERSION_REGEX
        ),
    ])
    .case_insensitive(true)
    .build()
    .unwrap();
}

pub struct Dockerfile {
    pub path: PathBuf,
    pub content: String,
}

impl Dockerfile {
    pub fn new(path: PathBuf) -> Result<Self, Box<dyn Error>> {
        let dockerfile = Self {
            path: path.to_owned(),
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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_dockerfile_images_none() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(r"FROM"),
        };

        assert_eq!(dockerfile.images().len(), 0);
    }

    #[test]
    fn test_dockerfile_images_one() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(r"FROM rust:latest"),
        };
        let images = dockerfile.images();

        assert_eq!(images.len(), 1);
        assert!(images.contains(&Image {
            repository: String::from("rust"),
            tag: Some(String::from("latest"))
        }));
    }

    #[test]
    fn test_dockerfile_images_sha() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(r"FROM rust:8d81c7bb21fa44bf6dffa1c5c3eff9be08dcd81a"),
        };
        let images = dockerfile.images();

        assert_eq!(images.len(), 1);
        assert!(images.contains(&Image {
            repository: String::from("rust"),
            tag: Some(String::from("8d81c7bb21fa44bf6dffa1c5c3eff9be08dcd81a"))
        }));
    }

    #[test]
    fn test_dockerfile_images_multiple() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(
                r"FROM rust:latest
                FROM alpine",
            ),
        };
        let images = dockerfile.images();

        assert_eq!(images.len(), 2);
        assert!(images.contains(&Image {
            repository: String::from("rust"),
            tag: Some(String::from("latest"))
        }));
        assert!(images.contains(&Image {
            repository: String::from("alpine"),
            tag: None
        }));
    }

    #[test]
    fn test_dockerfile_images_multistage_copy() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(
                r"FROM rust:latest AS builder
                FROM alpine
                COPY --from=builder",
            ),
        };
        let images = dockerfile.images();

        assert_eq!(images.len(), 2);
        assert!(images.contains(&Image {
            repository: String::from("rust"),
            tag: Some(String::from("latest"))
        }));
        assert!(images.contains(&Image {
            repository: String::from("alpine"),
            tag: None
        }));
    }

    #[test]
    fn test_dockerfile_images_multistage_alias() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(
                r"FROM rust:latest AS first
                FROM first AS second
                FROM alpine",
            ),
        };
        let images = dockerfile.images();

        assert_eq!(images.len(), 2);
        assert!(images.contains(&Image {
            repository: String::from("rust"),
            tag: Some(String::from("latest"))
        }));
        assert!(images.contains(&Image {
            repository: String::from("alpine"),
            tag: None
        }));
    }

    #[test]
    fn test_dockerfile_image_tags_none() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(r"FROM rust"),
        };
        let invalid_images = dockerfile.validate().invalid_images;

        assert_eq!(invalid_images.len(), 1)
    }

    #[test]
    fn test_dockerfile_image_tags_latest() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(r"FROM rust:latest"),
        };
        let invalid_images = dockerfile.validate().invalid_images;

        assert_eq!(invalid_images.len(), 1)
    }

    #[test]
    fn test_dockerfile_image_tags_semver() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(
                r"FROM rust:1.41
                FROM alpine:3.11",
            ),
        };
        let invalid_images = dockerfile.validate().invalid_images;

        assert_eq!(invalid_images.len(), 0)
    }

    #[test]
    fn test_dockerfile_image_tags_sha() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(r"FROM rust:8d81c7bb21fa44bf6dffa1c5c3eff9be08dcd81a"),
        };
        let invalid_images = dockerfile.validate().invalid_images;

        assert_eq!(invalid_images.len(), 1)
    }

    #[test]
    fn test_dockerfile_image_tags_date() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(r"FROM ubuntu:20200112"),
        };
        let invalid_images = dockerfile.validate().invalid_images;

        assert_eq!(invalid_images.len(), 0)
    }

    #[test]
    fn test_dockerfile_image_tags_datenumber() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(r"FROM eu.gcr.io/test/repository:2020011201"),
        };
        let invalid_images = dockerfile.validate().invalid_images;

        assert_eq!(invalid_images.len(), 0)
    }

    #[test]
    fn test_dockerfile_image_tags_semver_and_latest() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(
                r"FROM rust:1.41
                FROM alpine:latest",
            ),
        };
        let invalid_images = dockerfile.validate().invalid_images;

        assert_eq!(invalid_images.len(), 1)
    }

    #[test]
    fn test_dockerfile_image_tags_sha_and_date() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(
                r"FROM rust:8d81c7bb21fa44bf6dffa1c5c3eff9be08dcd81a
                FROM ubuntu:20200112",
            ),
        };
        let invalid_images = dockerfile.validate().invalid_images;

        assert_eq!(invalid_images.len(), 1);
    }

    #[test]
    fn test_dockerfile_image_tags_semver_and_datenumber() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(
                r"FROM rust:1.41
                FROM eu.gcr.io/test/repository:2020011201",
            ),
        };
        let invalid_images = dockerfile.validate().invalid_images;

        assert_eq!(invalid_images.len(), 0)
    }

    #[test]
    fn test_dockerfile_image_tags_prefix() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(r"FROM ubuntu:bionic-20200112"),
        };
        let invalid_images = dockerfile.validate().invalid_images;

        assert_eq!(invalid_images.len(), 0)
    }

    #[test]
    fn test_dockerfile_image_tags_suffix() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(r"FROM ubuntu:20200112-bionic"),
        };
        let invalid_images = dockerfile.validate().invalid_images;

        assert_eq!(invalid_images.len(), 0)
    }

    #[test]
    fn test_dockerfile_image_tags_multistage_copy() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(
                r"FROM rust:latest AS builder
                FROM alpine
                COPY --from=builder",
            ),
        };
        let invalid_images = dockerfile.validate().invalid_images;

        assert_eq!(invalid_images.len(), 2);
    }

    #[test]
    fn test_dockerfile_image_tags_multistage_alias() {
        let dockerfile = Dockerfile {
            path: PathBuf::new(),
            content: String::from(
                r"FROM rust:1.41 AS first
                FROM first AS second
                FROM alpine:3.11",
            ),
        };
        let invalid_images = dockerfile.validate().invalid_images;

        assert_eq!(invalid_images.len(), 0);
    }
}
