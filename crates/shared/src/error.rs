use axum::{
  http::StatusCode,
  response::{Response, IntoResponse},
  Json
};
use serde::Serialize;

pub struct AppError {
  /// An error message.
  pub error: anyhow::Error,
  pub status: StatusCode,
}

impl AppError {
  pub fn new(error: anyhow::Error, status: Option<StatusCode>) -> Self {
    Self {
      error,
      status: status.unwrap_or_else(|| StatusCode::INTERNAL_SERVER_ERROR),
    }
  }
}

#[derive(Serialize)]
pub struct AppErrorResponse {
  /// An error message.
  pub error: String,
  pub status: u16,
}

impl AppErrorResponse {
  pub fn new(err: AppError) -> Self {
    Self {
      error: err.error.to_string(),
      status: err.status.as_u16(),
    }
  }
}

impl IntoResponse for AppError {
  fn into_response(self) -> Response {
      (self.status, Json(AppErrorResponse::new(self))).into_response()
  }
}

impl<T> From<T> for AppError
where
  T: Into<anyhow::Error>,
{
  fn from(t: T) -> Self {
    Self::new(t.into(), None)
  }
}

// impl Display for AppError {
//     fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
//         writeln!(f, "{:?}", self.error)?;
//         self.context.fmt(f)?;
//         Ok(())
//     }
// }
