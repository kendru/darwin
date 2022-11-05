use config;
use serde::Deserialize;

#[derive(Deserialize, Debug)]
pub struct Config {
    pub http_addr: String,
    pub http_scheme: String,
    pub postgres_addr: String,
    pub postgres_db: String,
    pub postgres_user: String,
    pub postgres_password: String,
    pub signing_secret: String,
}

impl Config {
    pub fn db_url(&self) -> String {
        format!(
            "postgres://{}:{}@{}/{}",
            self.postgres_user,
            self.postgres_password,
            self.postgres_addr,
            self.postgres_db,
        )
    }
}


pub fn load_config<S: AsRef<str>>(path: Option<S>) -> Result<Config, config::ConfigError> {
    let mut config = config::Config::default();

    if let Some(config_path) = path {
        let path_name = config_path.as_ref();
        log::info!("Reading config from {}.", path_name);
        config.merge(config::File::with_name(path_name))?;
    }

    config.merge(config::Environment::new())?;

    config.try_into::<Config>()
}
