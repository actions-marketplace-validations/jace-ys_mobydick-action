mod dockerfile;

use dockerfile::{Dockerfile, ValidationResult};
use glob::glob;
use std::{error::Error, process};

fn main() {
    let dockerfiles = get_dockerfiles().expect("[ERROR] Failed to get Dockerfiles.");
    if dockerfiles.is_empty() {
        skip();
    };

    let invalid = validate_dockerfiles(&dockerfiles);
    if invalid.is_empty() {
        pass();
    }

    fail(&invalid);
}

fn get_dockerfiles() -> Result<Vec<Dockerfile>, Box<dyn Error>> {
    glob("**/*Dockerfile*")?
        .map(|path| Dockerfile::new(path?))
        .collect()
}

fn validate_dockerfiles(dockerfiles: &[Dockerfile]) -> Vec<(&Dockerfile, ValidationResult)> {
    dockerfiles
        .iter()
        .map(|dockerfile| (dockerfile, dockerfile.validate()))
        .filter(|(_, result)| !result.invalid_images.is_empty())
        .collect()
}

fn skip() {
    println!("[PASS] No Dockerfiles found.");
    process::exit(0);
}

fn pass() {
    println!("[PASS] All Dockerfiles are using versioned images.");
    process::exit(0);
}

fn fail(results: &[(&Dockerfile, ValidationResult)]) {
    println!("[FAIL] Found Dockerfiles not using versioned images.");
    results.iter().for_each(|(dockerfile, result)| {
        println!("{}:", dockerfile.path.display());
        result.invalid_images.iter().for_each(|image| {
            let image_tag = match &image.tag {
                Some(tag) => format!(":{}", tag),
                None => String::new(),
            };
            println!("- {}{}", image.repository, image_tag);
        });
        println!();
    });
    process::exit(1);
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::{fs, path::PathBuf};

    #[test]
    fn test_get_dockerfiles() {
        let dockerfiles = get_dockerfiles().unwrap();

        assert_eq!(dockerfiles.len(), 1);
        assert_eq!(dockerfiles[0].path, PathBuf::from("Dockerfile"));
        assert_eq!(
            dockerfiles[0].content,
            fs::read_to_string(PathBuf::from("Dockerfile")).unwrap()
        )
    }
}
