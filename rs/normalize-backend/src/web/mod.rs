use crate::state::{DbPool, AppInfo};
use crate::{accounts::routes as accounts_routes, mail::DynMailer};
use crate::config;
use actix_web::web::Data;
use actix_web::{dev::Server, middleware, web, App, HttpServer};

mod auth;

pub fn start_http_server(
    cfg: &config::Config,
    info: AppInfo,
    pool: DbPool,
    mailer: DynMailer,
) -> std::io::Result<Server> {
    let mailer_data = web::Data::new(mailer);
    let server = HttpServer::new(move || {
        App::new()
            .wrap(auth::Authenticate::new(&info.signing_secret))
            .wrap(middleware::NormalizePath::new(
                middleware::TrailingSlash::Always,
            ))
            .app_data(Data::new(info.clone()))
            .app_data(Data::new(pool.clone()))
            .app_data(mailer_data.clone())
            .service(web::scope("/accounts").configure(accounts_routes::users_route_config))
            .service(web::scope("/login").configure(accounts_routes::login_route_config))
    })
    .bind(&cfg.http_addr)?
    .run();

    Ok(server)
}
