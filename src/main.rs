use axum::{
  routing::{get, post},
  Router,
};
mod shutdown_signal;
use shutdown_signal::shutdown_signal;

mod speech;
use speech::speech;

#[tokio::main]
async fn main() {
  tracing_subscriber::fmt::init();

  let app = Router::new()
    .route("/", get(root))
    .route("/v1/audio/speech", post(speech));

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
