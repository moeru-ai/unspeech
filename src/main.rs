use axum::{
  http::StatusCode,
  response::IntoResponse,
  routing::{get, post},
  Json, Router,
};
use serde::{Deserialize, Serialize};

mod shutdown_signal;
use shutdown_signal::shutdown_signal;

#[tokio::main]
async fn main() {
  tracing_subscriber::fmt::init();

  let app = Router::new()
    .route("/", get(root));

  let listener = tokio::net::TcpListener::bind("127.0.0.1:3000")
    .await
    .unwrap();

  axum::serve(listener, app)
    .with_graceful_shutdown(shutdown_signal())
    .await
    .unwrap();
}

async fn root() -> &'static str {
  "Hello, World!"
}
