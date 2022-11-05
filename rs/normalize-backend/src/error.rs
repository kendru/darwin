use std::{error, fmt, result};

use actix_web::{http::StatusCode, HttpResponse, ResponseError};
use diesel::result::{DatabaseErrorKind, Error as DieselError};
use log::error;
use serde::Deserialize;
use serde_json::json;

pub type ApiResult<T> = result::Result<T, ApiError>;

#[derive(Debug, Deserialize)]
pub struct ApiError {
    message: String,
    status_code: u16,
    #[serde(skip_deserializing)]
    source: Option<Box<dyn error::Error + Send + 'static>>,
}

impl ApiError {
    pub fn new<S: Into<String>>(status_code: u16, message: S) -> ApiError {
        ApiError {
            message: message.into(),
            status_code,
            source: None,
        }
    }

    pub fn from_error(error: Box<dyn error::Error + Send + 'static>, status_code: u16) -> ApiError {
        ApiError {
            message: format!("{}", error),
            status_code,
            source: Some(error),
        }
    }

    pub fn not_found() -> ApiError {
        ApiError::new(404, "Not Found")
    }

    pub fn internal_server_error() -> ApiError {
        ApiError::new(500, "Internal Server Error")
    }
}

impl fmt::Display for ApiError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        f.write_str(self.message.as_str())
    }
}

impl error::Error for ApiError {
    fn source(&self) -> Option<&(dyn error::Error + 'static)> {
        // There has to be a better way to drop the Send bound from this trait object.
        self.source.as_ref().map(|e| {
            let without_send = *e.downcast_ref::<&(dyn error::Error)>().unwrap();
            without_send
        })
    }
}

impl From<DieselError> for ApiError {
    fn from(error: DieselError) -> Self {
        match error {
            DieselError::DatabaseError(DatabaseErrorKind::UnableToSendCommand, err) => {
                ApiError::new(500, err.message().to_string())
            }
            DieselError::DatabaseError(_, err) => {
                ApiError::new(409, err.message().to_string())
            },
            DieselError::NotFound => ApiError::new(404, "Not found"),
            error => ApiError::new(500, format!("Diesel error: {}", error)),
        }
    }
}

impl ResponseError for ApiError {
    fn status_code(&self) -> StatusCode {
        match StatusCode::from_u16(self.status_code) {
            Ok(status_code) => status_code,
            Err(_) => StatusCode::INTERNAL_SERVER_ERROR,
        }
    }

    fn error_response(&self) -> HttpResponse {
        let status_code = self.status_code();
        let message = if status_code.as_u16() < 500 {
            self.message.clone()
        } else {
            error!("{}", self.message);
            "Internal Server Error".to_string()
        };

        HttpResponse::build(status_code).json(json!({ "message": message }))
    }
}
