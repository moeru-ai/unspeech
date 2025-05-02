use axum::{
  body::Bytes,
  debug_handler,
  extract::State,
  Json
};
use axum_extra::{
  headers::{authorization::Bearer, Authorization},
  TypedHeader
};
use reqwest::Client;
use unspeech_shared::{
  AppError,
  speech::{
    SpeechOptions,
    process_speech_options
  }
};

#[debug_handler]
pub async fn speech(
  State(client): State<Client>,
  TypedHeader(bearer): TypedHeader<Authorization<Bearer>>,
  Json(body): Json<SpeechOptions>,
) -> Result<Bytes, AppError> {
  tracing::info!("Bearer {}", bearer.token());

  let options = process_speech_options(body);

  match options.provider.as_str() {
    #[cfg(feature = "openai")]
    "openai" => unspeech_provider_openai::speech::handle(options, client).await,
    _ => Err(AppError::anyhow(format!("Unsupported provider: {}", options.provider))),
  }
}
