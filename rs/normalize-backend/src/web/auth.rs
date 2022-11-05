use std::{rc::Rc, ops::Deref};

use actix_web::{dev::{Transform, Service, ServiceRequest, ServiceResponse}, Error, HttpMessage, http::header, body::MessageBody};
use futures_util::future::{self, FutureExt as _, LocalBoxFuture, TryFutureExt as _};

use crate::{accounts::actions::verify_session_token, error::ApiError};

pub struct Authenticate {
    secret: String,
}

impl Authenticate {
    pub fn new<S: Into<String>>(secret: S) -> Authenticate {
        Authenticate { secret: secret.into() }
    }
}

impl<S, B> Transform<S, ServiceRequest> for Authenticate
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = Error> + 'static,
    S::Future: 'static,
    B: MessageBody + 'static,
    B::Error: Into<Error>,
{
    type Response = ServiceResponse<B>;
    type Error = Error;
    type Transform = AuthenticateMiddleware<S>;
    type InitError = ();
    type Future = future::Ready<Result<Self::Transform, Self::InitError>>;

    fn new_transform(&self, service: S) -> Self::Future {
        future::ok(AuthenticateMiddleware {
            inner: Rc::new(Inner {
                service,
                secret: self.secret.clone(),
            }),
        })
    }
}

pub struct AuthenticateMiddleware<S> {
    inner: Rc<Inner<S>>
}

struct Inner<S> {
    service: S,
    secret: String,
}

// This impl allows the forward_ready macro to work correctly.
impl<S> Deref for Inner<S> {
    type Target = S;

    fn deref(&self) -> &Self::Target {
        &self.service
    }
}


impl<S, B> Service<ServiceRequest> for AuthenticateMiddleware<S>
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = Error> + 'static,
    S::Future: 'static,
    B: MessageBody + 'static,
    B::Error: Into<Error>,
{
    type Response = ServiceResponse<B>;
    type Error = S::Error;
    type Future = LocalBoxFuture<'static, Result<Self::Response, Self::Error>>;

    actix_service::forward_ready!(inner);

    fn call(&self, req: ServiceRequest) -> Self::Future {
        let inner = Rc::clone(&self.inner);

        async move {
            let service = &inner.service;
            let secret = inner.secret.as_bytes();

            let authorization_header = req.headers()
                .get(header::AUTHORIZATION)
                .and_then(|header| header.to_str().ok()) ;

            if let Some(header) = authorization_header {
                let mut parts = header.splitn(2, " ");
                match parts.next() {
                    Some(scheme) if scheme == "Bearer" => (),
                    _ => return Err(ApiError::new(400, "Invalid authorization header").into()),
                };

                let token = parts
                    .next()
                    .ok_or_else(|| ApiError::new(400, "Invalid authorization header"))?;

                let claims = verify_session_token(secret, token)?;
                println!("Claims: {:?}", claims);
            }

            let res = service.call(req).await?;

            Ok(res)
        }
        .boxed_local()
    }
}
