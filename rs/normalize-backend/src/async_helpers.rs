use actix_web::{web, error::BlockingError};

use crate::error::{ApiResult, ApiError};

// This proxies to actix_web::web::block with the appropriate result coercions applied.
pub async fn block<F, I, E>(f: F) -> ApiResult<I>
where
    F: FnOnce() -> Result<I, E> + Send + 'static,
    I: Send + 'static,
    E: Send + std::fmt::Debug + Into<ApiError> + 'static,
{
    web::block(f)
        .await
        .map_err(|e| e.try_into())
}
