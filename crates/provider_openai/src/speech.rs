use axum::body::Bytes;
use reqwest::Client;
use serde_json::json;
use unspeech_shared::speech::ProcessedSpeechOptions;

pub async fn handle(
  options: ProcessedSpeechOptions,
  client: Client,
) -> Result<Bytes, reqwest::Error> {
  let body = json!({
    "input": options.input,
    "model": options.model,
    "voice": options.voice,
  });

  let res = client.post("https://api.openai.com/v1/audio/speech")
    .bearer_auth("TODO: API_KEY")
    .json(&body)
    .send()
    .await?;

  // if !res.status().is_success() {
  //   let status = res.status();
  //   let body = res.text().await.unwrap_or_else(|_| "Could not read error body".to_string());
  //   return Err(format!("API request failed with status: {}\nBody: {}", status, body).into());
  // }

  let bytes = res.bytes().await?;

  Ok(bytes)
}
