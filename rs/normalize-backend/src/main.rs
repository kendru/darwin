#[macro_use]
extern crate diesel;
#[macro_use]
extern crate diesel_migrations;

// pub mod async_helpers;
pub mod accounts;
pub mod config;
pub mod error;
pub mod mail;
pub mod schema;
pub mod state;
pub mod web;

use clap::Parser;
use diesel::prelude::*;
use diesel::r2d2::{ConnectionManager, Pool};
use state::AppInfo;

use crate::mail::new_mailer;

embed_migrations!();
fn run_migration(conn: &PgConnection) {
    embedded_migrations::run_with_output(conn, &mut std::io::stderr())
        .expect("could not run migrations");
}

#[derive(Parser, Debug)]
#[clap(name = "normalize")]
struct CmdArgs {
    #[clap(short, long)]
    config_file: Option<String>,

    #[clap(short, long)]
    migrate: bool,
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();

    let args = CmdArgs::parse();
    let cfg = config::load_config(args.config_file.as_ref())
        .map_err(|err| {
            log::error!("Could not load config: {}", err);
            err
        })
        .expect("Could not load config");

    log::info!("Setting up db connection pool");
    let db_conn_manager = ConnectionManager::<PgConnection>::new(cfg.db_url());
    let pool = Pool::builder()
        .connection_timeout(std::time::Duration::from_secs(5))
        .build(db_conn_manager)
        .expect("Could not create DB connection pool.");

    if args.migrate {
        log::info!("Running database migrations.");
        let conn = pool.get().expect("Cannot get db connection");
        run_migration(&conn);
        return Ok(());
    }

    // TODO: Pass config into factory functions.
    let mailer = new_mailer();
    let info = AppInfo {
        name: "normalize.news".to_string(),
        version: env!("CARGO_PKG_VERSION").to_string(),
        url: format!("{}://{}", &cfg.http_scheme, &cfg.http_addr),
        signing_secret: cfg.signing_secret.clone(),
    };

    log::info!("Listening on {}", &cfg.http_addr);
    web::start_http_server(&cfg, info, pool, mailer)?.await
}
