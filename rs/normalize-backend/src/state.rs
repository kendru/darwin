use diesel::prelude::*;
use diesel::r2d2::{Pool, ConnectionManager, PooledConnection};

pub type DbPool = Pool<ConnectionManager<PgConnection>>;
pub type DbConn = PooledConnection<ConnectionManager<PgConnection>>;

#[derive(Debug, Clone)]
pub struct AppInfo {
    pub name: String,
    pub version: String,
    pub url: String,
    pub signing_secret: String,
}
