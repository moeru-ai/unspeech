use axum::body::Bytes;
use reqwest::Client;
use serde_json::json;
use unspeech_shared::{
  AppError,
  speech::ProcessedSpeechOptions,
};

pub async fn handle(
  options: ProcessedSpeechOptions,
  client: Client,
  token: &str,
) -> Result<Bytes, AppError> {
  let body = json!({
    "input": options.input,
    "model": options.model,
    "voice": options.voice,
  });

  let res = client.post("https://api.openai.com/v1/audio/speech")
    .bearer_auth(token)
    .json(&body)
    .send()
    .await?;

  if !res.status().is_success() {
    let status = res.status();
    let body = res.text().await.unwrap_or_else(|_| "Could not read error body".to_string());
    return Err(AppError::anyhow(format!("API request failed with status: {}\nBody: {}", status, body)));
  }

  let bytes = res
    .bytes()
    .await?;

  Ok(bytes)
}
