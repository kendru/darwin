use argon2;
use diesel::prelude::*;
use diesel::sql_types::{Interval, Text};
use diesel::{result::Error as DieselError, PgConnection};
use jsonwebtoken::{Header, EncodingKey, DecodingKey, Validation};
use rand::{
    distributions::{Alphanumeric, Uniform},
    thread_rng, Rng,
};
use uuid::Uuid;

use crate::error::{ApiError, ApiResult};

use super::models::{NewUser, NewUserInput, SendTokenInput, User, VerifyTokenInput, Claims};

pub fn list_users(conn: &PgConnection) -> Result<Vec<User>, ApiError> {
    use crate::schema::user::dsl::*;
    let users = user.limit(1000).load::<User>(conn)?;
    Ok(users)
}

pub fn get_user(conn: &PgConnection, find_id: &str) -> Result<Option<User>, ApiError> {
    let user_id = Uuid::parse_str(find_id).map_err(|_e| DieselError::NotFound)?; // TODO: Use a custom error type.
    let user = {
        use crate::schema::user::dsl::*;
        user.filter(id.eq(&user_id))
            .first::<User>(conn)
            .optional()?
    };

    Ok(user)
}

pub fn get_user_by_email_address(
    conn: &PgConnection,
    find_email_address: &str,
) -> Result<Option<User>, ApiError> {
    let user = {
        use crate::schema::user::dsl::*;
        user.filter(email_address.eq(find_email_address))
            .first::<User>(conn)
            .optional()
    }?;
    Ok(user)
}

pub fn create_user(conn: &PgConnection, input: NewUserInput) -> Result<User, ApiError> {
    let password_digest = input.password.map(|password| {
        let salt: Vec<u8> = thread_rng()
            .sample_iter(Uniform::new(0, 255))
            .take(8)
            .collect();
        let config = argon2::Config::default();
        argon2::hash_raw(password.as_bytes(), &salt, &config)
            // TODO: Use custom error type and return error on hash failure.
            .expect("Hash should not fail.")
    });

    let new_user = NewUser {
        email_address: &input.email_address,
        name: input.name.as_ref().map(|n| n.as_str()),
        password_digest: password_digest.as_ref().map(|pw| pw.as_ref()),
    };

    let user = diesel::insert_into(crate::schema::user::table)
        .values(&new_user)
        .get_result(conn)?;

    Ok(user)
}

pub fn new_login_token(
    conn: &PgConnection,
    input: SendTokenInput,
) -> Result<String, ApiError> {
    let token = gen_random_token(12);
    create_user_token(conn, "login", &input.email_address, &token, 10)?;
    Ok(token)
}

pub fn verify_login_token(
    conn: &PgConnection,
    input: VerifyTokenInput,
) -> Result<User, ApiError> {
    redeem_user_token(conn, "login", &input.token)
}

pub fn new_email_token(
    conn: &PgConnection,
    input: SendTokenInput,
) -> Result<String, ApiError> {
    let token = gen_random_token(12);
    create_user_token(conn, "email", &input.email_address, &token, 60 * 24 * 2)?;
    Ok(token)
}

pub fn verify_email_token(
    conn: &PgConnection,
    input: VerifyTokenInput,
) -> Result<User, ApiError> {
    conn.build_transaction().run::<_, ApiError, _>(|| {
        use crate::schema::user::dsl::*;
        use diesel::dsl::*;

        let u = redeem_user_token(conn, "email", &input.token)?;
        diesel::update(user.filter(id.eq(&u.id)))
            .set(email_verified_at.eq(now))
            .execute(conn)?;

        Ok(u)
    })
}

pub fn create_session_token(secret: &[u8], claims: &Claims) -> ApiResult<String> {
    let header = Header::default();
    let key = EncodingKey::from_secret(secret);
    // TODO: Implement From<jsonwebtoken::error::Error> for ApiError.
    jsonwebtoken::encode(&header, claims, &key)
        .map_err(|err| ApiError::from_error(Box::new(err), 500))
}

pub fn verify_session_token(secret: &[u8], token: &str) -> ApiResult<Claims> {
    let key = DecodingKey::from_secret(secret);
    println!("Token: {}", token);
    jsonwebtoken::decode::<Claims>(token, &key, &Validation::default())
        .map(|data| data.claims)
        .map_err(|err| ApiError::from_error(Box::new(err), 500))
}

fn create_user_token(
    conn: &PgConnection,
    token_type: &str,
    email_address: &str,
    token: &str,
    expires_minutes: i32,
) -> Result<(), ApiError> {
    use diesel::dsl::*;
    sql_query(include_str!("sql/insert_user_token.sql"))
        .bind::<Text, _>(email_address)
        .bind::<Text, _>(token)
        .bind::<Text, _>(token_type)
        .bind::<Interval, _>(expires_minutes.minutes())
        .execute(conn)?;
    Ok(())
}

fn redeem_user_token(conn: &PgConnection, token_type: &str, token: &str) -> Result<User, ApiError> {
    use diesel::dsl::*;
    let user = sql_query(include_str!("sql/redeem_user_token.sql"))
        .bind::<Text, _>(token_type)
        .bind::<Text, _>(token)
        .load::<User>(conn)
        .and_then(first_or_not_found)?;
    Ok(user)
}

fn gen_random_token(chars: usize) -> String {
    thread_rng()
        .sample_iter(Alphanumeric)
        .take(chars)
        .map(char::from)
        .collect()
}

fn first_or_not_found<T>(rows: Vec<T>) -> Result<T, DieselError> {
    rows.into_iter().next().ok_or(DieselError::NotFound)
}
