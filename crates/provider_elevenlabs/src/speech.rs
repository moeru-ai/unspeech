use axum::{
  body::Bytes,
  http::{HeaderMap, header},
};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use unspeech_shared::{AppError, speech::ProcessedSpeechOptions};

#[derive(Serialize)]
// https://elevenlabs.io/docs/api-reference/text-to-speech/convert
pub struct ElevenLabsSpeechOptions {
  pub text: String,
  pub model_id: String,
  pub voice_settings: ElevenLabsSpeechOptionsVoiceSettings,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub seed: Option<f64>,
}

#[derive(Deserialize, Serialize)]
pub struct ElevenLabsSpeechOptionsVoiceSettings {
  #[serde(skip_serializing_if = "Option::is_none")]
  pub stability: Option<f64>,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub similarity_boost: Option<f64>,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub style: Option<f64>,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub use_speaker_boost: Option<bool>,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub speed: Option<f64>,
}

pub async fn handle(
  options: ProcessedSpeechOptions,
  client: Client,
  token: &str,
) -> Result<(HeaderMap, Bytes), AppError> {
  let voice_settings: ElevenLabsSpeechOptionsVoiceSettings = match options.extra.get("voice_settings") {
    Some(v) => ElevenLabsSpeechOptionsVoiceSettings {
      speed: options.speed,
      ..serde_json::from_value::<ElevenLabsSpeechOptionsVoiceSettings>(v.clone())?
    },
    None => ElevenLabsSpeechOptionsVoiceSettings {
      stability: None,
      similarity_boost: None,
      style: None,
      use_speaker_boost: None,
      speed: options.speed,
    },
  };

  let body = ElevenLabsSpeechOptions {
    text: options.input,
    model_id: options.model,
    voice_settings,
    seed: options.extra.get("seed").and_then(|v| v.as_f64()),
  };

  // TODO: output_format
  // TODO: query parameters
  // https://elevenlabs.io/docs/api-reference/text-to-speech/convert#request.query
  let url = format!("https://api.elevenlabs.io/v1/text-to-speech/{}?output_format=mp3_44100_128", options.voice);

  let res = client
    .post(&url)
    .header("xi-api-key", token)
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
  headers.insert(header::CONTENT_TYPE, "audio/mpeg".parse()?);

  Ok((headers, bytes))
}
