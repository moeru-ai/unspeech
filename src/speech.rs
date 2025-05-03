use axum::{Json, body::Bytes, debug_handler, extract::State, http::HeaderMap};
use axum_extra::{
  TypedHeader,
  headers::{Authorization, authorization::Bearer},
};
use reqwest::Client;
use unspeech_shared::{
  AppError,
  speech::{SpeechOptions, process_speech_options},
};

#[debug_handler]
pub async fn speech(
  State(client): State<Client>,
  TypedHeader(bearer): TypedHeader<Authorization<Bearer>>,
  Json(body): Json<SpeechOptions>,
) -> Result<(HeaderMap, Bytes), AppError> {
  let options = process_speech_options(body)?;
  let token = bearer.token();

  match options.provider.as_str() {
    #[cfg(feature = "openai")]
    "openai" => unspeech_provider_openai::speech::handle(options, client, token).await,
    _ => Err(AppError::new(
      anyhow::anyhow!("Unsupported provider: {}", options.provider),
      None,
    )),
  }
}
