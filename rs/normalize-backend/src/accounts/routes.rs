use actix_web::{web, Error as ResponseError, HttpResponse};
use chrono::Utc;
use uuid::Uuid;

use super::models::{NewUserInput, VerifyTokenInput, SessionTokenResponse, TokenType, Claims};
use super::{actions, models::SendTokenInput};
use crate::error::ApiError;
use crate::mail::DynMailer;
use crate::state::{AppInfo, DbConn, DbPool};

fn must_get_conn(pool: web::Data<DbPool>) -> DbConn {
    pool.get().expect("could not get db connection from pool")
}

async fn list_users(pool: web::Data<DbPool>) -> Result<HttpResponse, ResponseError> {
    let conn = must_get_conn(pool);
    let users = web::block(move || actions::list_users(&conn)).await??;

    Ok(HttpResponse::Ok().json(users))
}

async fn get_user(
    pool: web::Data<DbPool>,
    id: web::Path<String>,
) -> Result<HttpResponse, ResponseError> {
    let conn = must_get_conn(pool);
    let user = web::block(move || actions::get_user(&conn, id.as_str()))
        .await??
        .ok_or(ApiError::new(404, "Not found"))?;

    Ok(HttpResponse::Ok().json(user))
}

async fn create_user(
    pool: web::Data<DbPool>,
    input: web::Json<NewUserInput>,
) -> Result<HttpResponse, ResponseError> {
    let conn = must_get_conn(pool);
    let user = web::block(move || actions::create_user(&conn, input.into_inner())).await??;

    Ok(HttpResponse::Created().json(user))
}

async fn send_login_token(
    info: web::Data<AppInfo>,
    pool: web::Data<DbPool>,
    mailer: web::Data<DynMailer>,
    input: web::Json<SendTokenInput>,
) -> Result<HttpResponse, ResponseError> {
    let conn = must_get_conn(pool);
    let email_address = input.email_address.clone();
    // TODO: Do not send if the user's email address has not been verified.
    let token = web::block(move || actions::new_login_token(&conn, input.into_inner())).await??;
    let message = format!(
        r#"
<h1>Your login token for {}</h1>
<a href="{}/login/token/verify?token={}">Login</a>
"#,
        &info.name, &info.url, &token
    );
    mailer.send_one(&message, &email_address).await?;

    Ok(HttpResponse::Ok().body("ok"))
}

async fn verify_login_token(
    info: web::Data<AppInfo>,
    pool: web::Data<DbPool>,
    input: web::Query<VerifyTokenInput>,
) -> Result<HttpResponse, ResponseError> {
    let conn = must_get_conn(pool);
    let user = web::block(move || actions::verify_login_token(&conn, input.into_inner())).await??;
    let expiration = Utc::now()
        .checked_add_signed(chrono::Duration::minutes(10))
        .expect("valid timestamp")
        .timestamp();
    let claims = Claims {
        subject: user.id,
        issuer: Some(info.url.clone()),
        audience: None,
        expiration: expiration as u64,
        not_before: Some(0),
        issued_at: Some(0),
        id: Some(Uuid::new_v4().to_string()),
        custom_claims: None,
    };
    let claims_copy = claims.clone();
    let token = web::block(move || actions::create_session_token(info.signing_secret.as_bytes(), &claims_copy)).await??;

    Ok(HttpResponse::Ok().json(SessionTokenResponse {
        token,
        claims,
        token_type: TokenType::Jwt,
    }))
}

async fn send_email_token(
    info: web::Data<AppInfo>,
    pool: web::Data<DbPool>,
    mailer: web::Data<DynMailer>,
    input: web::Json<SendTokenInput>,
) -> Result<HttpResponse, ResponseError> {
    let conn = must_get_conn(pool);
    let email_address = input.email_address.clone();
    // TODO: Do not send if the user's email address has not been verified.
    let token = web::block(move || actions::new_email_token(&conn, input.into_inner())).await??;
    let message = format!(
        r#"
<h1>Verify your email address for {}</h1>
<p><a href="{}/email/token/verify?token={}">Verify</a></p>
<p>If you did not sign up for {}, you can safely ignore this message.</p>
"#,
        &info.name, &info.url, &token, &info.name
    );
    mailer.send_one(&message, &email_address).await?;

    Ok(HttpResponse::Ok().body("ok"))
}

async fn verify_email_token(
    pool: web::Data<DbPool>,
    input: web::Query<VerifyTokenInput>,
) -> Result<HttpResponse, ResponseError> {
    let conn = must_get_conn(pool);
    let user = web::block(move || actions::verify_email_token(&conn, input.into_inner())).await??;

    Ok(HttpResponse::Ok().json(user))
}

pub fn users_route_config(cfg: &mut web::ServiceConfig) {
    cfg.service(
        web::scope("/users")
            // Users routes
            .route("/", web::get().to(list_users))
            .route("/", web::post().to(create_user))

            // Email verification routes
            .service(web::scope("/email-token")
                .route("/", web::post().to(send_email_token))
                .service(web::scope("/verify").route("/", web::get().to(verify_email_token))))

            // User routes
            .service(web::scope("/{id}")
                .route("/", web::get().to(get_user))
            //     .route("/", web::put().to(update_user))
            //     .route("/", web::delete().to(delete_user))),
        ),
    );
}

pub fn login_route_config(cfg: &mut web::ServiceConfig) {
    cfg.service(
        // One-time token login.
        web::scope("/token")
            .route("/", web::post().to(send_login_token))
            .service(web::scope("/verify").route("/", web::get().to(verify_login_token))),

        // TODO: Password login.
        // TODO: TOTP login.
    );
}
