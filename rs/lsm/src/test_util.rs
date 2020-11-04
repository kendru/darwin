use rand;
use std::path::{Path, PathBuf};
use std::fs;

pub struct TmpDir {
    path: PathBuf,
}

impl TmpDir {
    pub(crate) fn new() -> TmpDir {
        let mut path = PathBuf::new();
        path.push("target");
        path.push("testdata");
        path.push(format!("test-{:20}", rand::random::<u64>()));
        fs::create_dir_all(&path).expect("could not create test data directory");

        TmpDir { path }
    }
}

impl Drop for TmpDir {
    fn drop(&mut self) {
        fs::remove_dir_all(&self.path).expect("could not remove test data directory");
    }
}

impl AsRef<Path> for TmpDir {
    fn as_ref(&self) -> &Path {
        self.path.as_ref()
    }
}
