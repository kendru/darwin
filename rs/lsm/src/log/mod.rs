pub mod segment;

pub struct LogEntry {
    pub key: Vec<u8>,
    pub value: Vec<u8>,
}

pub fn test_log() {
    println!("Test from log");
}
