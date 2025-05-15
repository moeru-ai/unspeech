use axum::{
  body::Bytes,
  http::{HeaderMap, header},
};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use unspeech_shared::{AppError, speech::ProcessedSpeechOptions};
use url::Url;

#[derive(serde::Serialize)]
// https://elevenlabs.io/docs/api-reference/text-to-speech/convert#request.query
pub struct ElevenLabsSpeechQuery {
  #[serde(skip_serializing_if = "Option::is_none")]
  pub enable_logging: Option<bool>,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub output_format: Option<String>,
}

#[derive(Serialize)]
// https://elevenlabs.io/docs/api-reference/text-to-speech/convert
pub struct ElevenLabsSpeechOptions {
  pub text: String,
  pub model_id: String,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub language_code: Option<String>,
  pub voice_settings: ElevenLabsSpeechVoiceSettings,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub pronunciation_dictionary_locators: Option<Vec<ElevenLabsSpeechPronunciationDictionaryLocator>>,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub seed: Option<f64>,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub previous_text: Option<String>,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub next_text: Option<String>,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub previous_request_ids: Option<Vec<String>>,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub next_request_ids: Option<Vec<String>>,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub apply_text_normalization: Option<String>,
  #[serde(skip_serializing_if = "Option::is_none")]
  pub apply_language_text_normalization: Option<bool>,
}

#[derive(Deserialize, Serialize)]
// https://elevenlabs.io/docs/api-reference/text-to-speech/convert#request.body.voice_settings
pub struct ElevenLabsSpeechVoiceSettings {
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

#[derive(Deserialize, Serialize)]
// https://elevenlabs.io/docs/api-reference/text-to-speech/convert#request.body.voice_settings
pub struct ElevenLabsSpeechPronunciationDictionaryLocator {
  pub pronunciation_dictionary_id: String,
  pub version_id: Option<String>,
}

pub async fn handle(
  options: ProcessedSpeechOptions,
  client: Client,
  token: &str,
) -> Result<(HeaderMap, Bytes), AppError> {
  let voice_settings: ElevenLabsSpeechVoiceSettings = match options.extra.get("voice_settings") {
    Some(v) => ElevenLabsSpeechVoiceSettings {
      speed: options.speed,
      ..serde_json::from_value::<ElevenLabsSpeechVoiceSettings>(v.clone())?
    },
    None => ElevenLabsSpeechVoiceSettings {
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
    language_code: options.extra.get("language_code").and_then(|v| Some(v.to_string())),
    voice_settings,
    pronunciation_dictionary_locators: options.extra.get("language_code").and_then(|v| serde_json::from_value(v.clone()).ok()?),
    seed: options.extra.get("seed").and_then(|v| v.as_f64()),
    previous_text: options.extra.get("previous_text").and_then(|v| Some(v.to_string())),
    next_text: options.extra.get("next_text").and_then(|v| Some(v.to_string())),
    previous_request_ids: options.extra.get("previous_request_ids").and_then(|v| serde_json::from_value(v.clone()).ok()?),
    next_request_ids: options.extra.get("next_request_ids").and_then(|v| serde_json::from_value(v.clone()).ok()?),
    apply_text_normalization: options.extra.get("apply_text_normalization").and_then(|v| Some(v.to_string())),
    apply_language_text_normalization: options.extra.get("apply_language_text_normalization").and_then(|v| v.as_bool())
  };

  let query = serde_html_form::to_string(ElevenLabsSpeechQuery {
    enable_logging: options.extra.get("enable_logging").and_then(|v| v.as_bool()),
    output_format: options.response_format,
  })?;

  let url = Url::parse(&format!("https://api.elevenlabs.io/v1/text-to-speech/{}?{}", options.voice, query))?;

  let res = client
    .post(url.as_str())
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
