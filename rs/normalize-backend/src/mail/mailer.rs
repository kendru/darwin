use async_trait::async_trait;

use crate::error::ApiError;

#[async_trait]
pub trait Mailer {
    async fn send_one(&self, content: &str, recipient: &str) -> Result<(), ApiError>;
}


