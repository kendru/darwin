use crate::log;

pub mod config;

pub struct Agent {
    log: log::Log<std::fs::File>,
}

impl Agent {

    pub fn new(cfg: config::Config) -> Agent {
        let log = log::Log::open(format!("{}/{}", cfg.log_dir, "log")).expect("Error opening log");
        
        Agent{
            log,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_create_agent() {
        let agent = Agent::new(config::Config{
            log_dir: "./data/log".into(),
            sstable_dir: "".into(),
        });
    }
}
