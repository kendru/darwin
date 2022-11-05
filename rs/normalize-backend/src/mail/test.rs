use async_trait::async_trait;

use super::mailer::Mailer;
use crate::error::ApiError;

#[derive(Debug, Clone, Copy)]
pub struct TestMailer;

#[async_trait]
impl Mailer for TestMailer {
    async fn send_one(&self, content: &str, recipient: &str) -> Result<(), ApiError> {
        println!("--- Test Mailer ---\nTo: {}\n\n{}\n-------------------", recipient, content);
        Ok(())
    }
}
