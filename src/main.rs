mod dockerfile;

use dockerfile::Dockerfile;
use glob::glob;
use std::process;

#[derive(Debug, PartialEq)]
enum Status {
    Skip,
    Pass,
    Fail,
}

fn main() {
    let dockerfiles = get_dockerfiles();

    let status = match dockerfiles.len() {
        0 => Status::Skip,
        _ => validate_dockerfiles(&dockerfiles),
    };

    match status {
        Status::Skip => skip(),
        Status::Pass => pass(),
        Status::Fail => fail(&dockerfiles),
    }
}

fn get_dockerfiles() -> Vec<Dockerfile> {
    return glob("**/Dockerfile")
        .unwrap()
        .map(|path| Dockerfile::new(path.unwrap()))
        .collect();
}

fn validate_dockerfiles(dockerfiles: &[Dockerfile]) -> Status {
    let results: Vec<bool> = dockerfiles
        .into_iter()
        .map(|dockerfile| dockerfile.validate())
        .collect();

    if results.into_iter().all(|result| result == true) {
        Status::Pass
    } else {
        Status::Fail
    }
}

fn skip() {
    println!("[PASS] No Dockerfiles found.");
    process::exit(0);
}

fn pass() {
    println!("[PASS] All Dockerfiles are using versioned images.");
    process::exit(0);
}

fn fail(dockerfiles: &[Dockerfile]) {
    println!("[FAIL] Found Dockerfiles not using versioned images.");
    for dockerfile in dockerfiles {
        println!("{}: {}", dockerfile.path, dockerfile.content);
    }
    process::exit(1);
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_get_dockerfiles() {
        let dockerfiles = get_dockerfiles();

        assert_eq!(dockerfiles.len(), 2);
    }

    #[test]
    fn test_validate_dockerfiles() {
        let dockerfiles: Vec<Dockerfile> = vec![];
        let result = validate_dockerfiles(&dockerfiles);

        assert_eq!(result, Status::Pass);
    }
}
