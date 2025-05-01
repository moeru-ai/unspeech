use axum::{
  Json,
  extract::State,
  http::StatusCode,
  debug_handler,
};
use axum_extra::{
  headers::{authorization::Bearer, Authorization},
  TypedHeader
};
use reqwest::Client;
use unspeech_shared::speech::{
  SpeechOptions,
  ProcessedSpeechOptions,
  process_speech_options
};

#[debug_handler]
pub async fn speech(
  State(_client): State<Client>,
  TypedHeader(bearer): TypedHeader<Authorization<Bearer>>,
  Json(options): Json<SpeechOptions>,
) -> (StatusCode, Json<ProcessedSpeechOptions>) {
  tracing::info!("Bearer {}", bearer.token());

  (StatusCode::OK, Json(process_speech_options(options)))
}
