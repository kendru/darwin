use std::collections::HashMap;

use crate::schema::{user, user_token};
use chrono::NaiveDateTime;
use serde::{Serialize, Deserialize};
use uuid::Uuid;

#[derive(Queryable, QueryableByName, Serialize, Debug)]
#[serde(rename_all = "camelCase")]
#[table_name="user"]
pub struct User {
    pub id: Uuid,
    pub email_address: String,

    #[serde(skip_serializing)]
    pub password_digest: Option<Vec<u8>>,

    pub name: Option<String>,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
    pub email_verified_at: Option<NaiveDateTime>,
}

// New user.

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct NewUserInput {
    pub email_address: String,
    pub name: Option<String>,
    pub password: Option<String>,
}

#[derive(Insertable)]
#[table_name="user"]
pub struct NewUser<'a> {
    pub email_address: &'a str,
    pub name: Option<&'a str>,
    pub password_digest: Option<&'a [u8]>,
}

// User tokens.

#[derive(Queryable, Debug)]
pub struct UserToken {
    pub user_id: Uuid,
    pub token: Vec<u8>,
    pub created_at: NaiveDateTime,
    pub expires_at: NaiveDateTime,
    pub redeemed_at: Option<NaiveDateTime>,
    pub user_token_type_id: i16,
}

#[derive(Insertable)]
#[table_name="user_token"]
pub struct NewUserToken<'a> {
    pub user_id: &'a Uuid,
    pub token: &'a str,
    pub user_token_type_id: i16,
}

// Verify email.

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SendEmailVerificationInput {
    pub email_address: String,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct VerifyEmailInput {
    pub token: String,
}

// Login token.

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SendTokenInput {
    pub email_address: String,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct VerifyTokenInput {
    pub token: String,
}

// Session.

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct SessionTokenResponse {
    pub token: String,
    pub token_type: TokenType,
    pub claims: Claims,
}

#[derive(Debug, Serialize, Deserialize)]
pub enum TokenType {
    Jwt,
    Opaque,
}

#[derive(Debug, Serialize, Deserialize, PartialEq, Clone)]
pub struct Claims {
    #[serde(rename = "sub")]
    pub subject: Uuid,

    #[serde(rename = "iss")]
    pub issuer: Option<String>,

    #[serde(rename = "aud")]
    pub audience: Option<String>,

    #[serde(rename = "exp")]
    pub expiration: u64,

    #[serde(rename = "nbf")]
    pub not_before: Option<u64>,

    #[serde(rename = "iat")]
    pub issued_at: Option<u64>,

    #[serde(rename = "jti")]
    pub id: Option<String>,

    #[serde(flatten)]
    pub custom_claims: Option<HashMap<String, String>>,
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    #[test]
    fn test_claims_parse() {
        let input = json!({
            "sub": "b5f0e099-a588-4b29-89a9-7b47fe68e7ec",
            "iss": "https://auth.example.com",
            "aud": "https://myapp.example.com",
            "exp": 1641703000,
            "jti": "some-id",
            "urn:myapp:user:name": "Test User",
            "urn:myapp:user:role": "Editor",
        }).to_string();
        println!("{}", &input);
        let claims = serde_json::from_str::<Claims>(&input).unwrap();

        assert_eq!(claims, Claims {
            subject: Uuid::parse_str("b5f0e099-a588-4b29-89a9-7b47fe68e7ec").unwrap(),
            issuer: Some("https://auth.example.com".to_string()),
            audience: Some("https://myapp.example.com".to_string()),
            expiration: 1641703000,
            id: Some("some-id".to_string()),
            issued_at: None,
            not_before: None,
            custom_claims: Some(HashMap::from([
                ("urn:myapp:user:name".to_string(), "Test User".to_string()),
                ("urn:myapp:user:role".to_string(), "Editor".to_string()),
            ])),
        })
    }
}
