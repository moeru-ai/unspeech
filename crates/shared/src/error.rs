use axum::{
  http::StatusCode,
  response::{Response, IntoResponse},
  Json
};
use serde::Serialize;

pub enum AppError {
  AnyhowError(anyhow::Error),
  ReqwestError(reqwest::Error),
}

impl IntoResponse for AppError {
  fn into_response(self) -> Response {
      // How we want errors responses to be serialized
      #[derive(Serialize)]
      struct ErrorResponse {
          message: String,
      }

      let (status, message) = match self {
        AppError::AnyhowError(err) => (
          StatusCode::INTERNAL_SERVER_ERROR,
          err.to_string(),
        ),
        AppError::ReqwestError(err) => (
          StatusCode::INTERNAL_SERVER_ERROR,
          format!("Failed to fetch: {}", err),
        ),
      };

      (status, Json(ErrorResponse { message })).into_response()
  }
}

impl AppError {
  pub fn anyhow(err: String) -> Self {
    Self::AnyhowError(anyhow::anyhow!(err))
  }
}

impl From<reqwest::Error> for AppError {
  fn from(error: reqwest::Error) -> Self {
      Self::ReqwestError(error)
  }
}

// impl AppError {
//     #[must_use]
//     pub fn new(error: String, error_details: Option<Value>, status: Option<StatusCode>) -> Self {
//         Self {
//             error,
//             error_details,
//             error_id: Uuid::now_v7(),
//             status: status.unwrap_or(StatusCode::INTERNAL_SERVER_ERROR),
//             context: SpanTrace::capture(),
//         }
//     }

//     #[must_use]
//     pub fn not_found(kind: &str, name: &str) -> Self {
//         Self {
//             error: format!("Unable to find {kind} named {name}"),
//             error_details: None,
//             error_id: Uuid::now_v7(),
//             status: StatusCode::NOT_FOUND,
//             context: SpanTrace::capture(),
//         }
//     }

//     #[must_use]
//     pub fn anyhow(error: &anyhow::Error) -> Self {
//         Self::new(error.to_string(), None, None)
//     }
// }

// impl IntoResponse for AppError {
//     fn into_response(self) -> Response {
//         (self.status, Json(self)).into_response()
//     }
// }

// impl Display for AppError {
//     fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
//         writeln!(f, "{:?}", self.error)?;
//         self.context.fmt(f)?;
//         Ok(())
//     }
// }

// impl<T> From<T> for AppError
// where
//     T: Into<anyhow::Error>,
// {
//     fn from(t: T) -> Self {
//         Self::anyhow(&t.into())
//     }
// }
