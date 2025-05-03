use axum::{
  body::Bytes,
  http::{HeaderMap, header},
};
use reqwest::Client;
use serde::Serialize;
use unspeech_shared::{AppError, speech::ProcessedSpeechOptions};

#[derive(Serialize)]
// https://platform.openai.com/docs/api-reference/audio/createSpeech
pub struct OpenAISpeechOptions {
  // The text to generate audio for. The maximum length is 4096 characters.
  pub input: String,
  // One of the available TTS models.
  pub model: String,
  // The voice to use when generating the audio.
  pub voice: String,
  // Control the voice of your generated audio with additional instructions.
  #[serde(skip_serializing_if = "Option::is_none")]
  pub instructions: Option<String>,
  // The format to audio in.
  #[serde(skip_serializing_if = "Option::is_none")]
  pub response_format: Option<String>,
  // The speed of the generated audio.
  #[serde(skip_serializing_if = "Option::is_none")]
  pub speed: Option<f32>,
}

pub async fn handle(
  options: ProcessedSpeechOptions,
  client: Client,
  token: &str,
) -> Result<(HeaderMap, Bytes), AppError> {
  let body = OpenAISpeechOptions {
    input: options.input,
    model: options.model,
    voice: options.voice,
    instructions: options.instructions,
    response_format: options.response_format,
    speed: options.speed,
  };

  let res = client
    .post("https://api.openai.com/v1/audio/speech")
    .bearer_auth(token)
    .json(&body)
    .send()
    .await?;

  if !res.status().is_success() {
    let status = res.status();
    let body = res
      .text()
      .await
      .unwrap_or_else(|err| format!("Could not read error body: {}", err));
    return Err(AppError::new(
      anyhow::anyhow!("API request failed with status: {}\nBody: {}", status, body),
      None,
    ));
  }

  let bytes = res.bytes().await?;

  let mut headers = HeaderMap::new();
  // TODO: remove unwrap
  headers.insert(header::CONTENT_TYPE, "audio/mpeg".parse()?);

  Ok((headers, bytes))
}
