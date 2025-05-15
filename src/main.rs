use std::time::Duration;

use axum::{
  Router,
  routing::{get, post},
};
use reqwest::Client;

mod shutdown_signal;
use shutdown_signal::shutdown_signal;

mod speech;
use speech::speech;
use unspeech_shared::AppError;

#[tokio::main]
async fn main() -> Result<(), AppError> {
  tracing_subscriber::fmt::init();

  let client = Client::builder().timeout(Duration::from_secs(60)).build()?;

  let app = Router::new()
    .route("/", get(root))
    .route("/v1/audio/speech", post(speech))
    .with_state(client);

  let listener = tokio::net::TcpListener::bind("127.0.0.1:3000").await?;

  tracing::debug!("listening on {}", listener.local_addr()?);

  Ok(
    axum::serve(listener, app)
      .with_graceful_shutdown(shutdown_signal())
      .await?,
  )
}

async fn root() -> &'static str {
  "Hello, World!"
}
