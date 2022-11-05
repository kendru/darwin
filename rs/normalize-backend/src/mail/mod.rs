use self::{mailer::Mailer, test::TestMailer};

pub mod mailer;
pub mod test;

pub type DynMailer = Box<dyn Mailer + Send + Sync>;

pub fn new_mailer() -> DynMailer {
    Box::new(TestMailer)
}
