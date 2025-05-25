use std::collections::HashMap;

use axum::http::StatusCode;
use serde::{Deserialize, Serialize};
use serde_json::Value;

use crate::AppError;

#[derive(Deserialize)]
// https://platform.openai.com/docs/api-reference/audio/createSpeech
pub struct SpeechOptions {
  // The text to generate audio for. The maximum length is 4096 characters.
  pub input:           String,
  // One of the available TTS models.
  pub model:           String,
  // The voice to use when generating the audio.
  pub voice:           String,
  // Control the voice of your generated audio with additional instructions.
  pub instructions:    Option<String>,
  // The format to audio in.
  pub response_format: Option<String>,
  // The speed of the generated audio.
  pub speed:           Option<f64>,
  #[serde(flatten)]
  pub extra:           HashMap<String, Value>,
}

#[derive(Serialize)]
pub struct ProcessedSpeechOptions {
  // The text to generate audio for. The maximum length is 4096 characters.
  pub input:           String,
  // One of the available TTS models.
  pub model:           String,
  // The voice to use when generating the audio.
  pub voice:           String,
  // Control the voice of your generated audio with additional instructions.
  pub instructions:    Option<String>,
  // The format to audio in.
  pub response_format: Option<String>,
  // The speed of the generated audio.
  pub speed:           Option<f64>,
  #[serde(flatten)]
  pub extra:           HashMap<String, Value>,
  // One of the available TTS providers.
  pub provider:        String,
}

pub fn process_speech_options(options: SpeechOptions) -> Result<ProcessedSpeechOptions, AppError> {
  match options.model.split_once('/') {
    Some((provider, model)) if !provider.is_empty() && !model.is_empty() => Ok(ProcessedSpeechOptions {
      input:           options.input,
      model:           model.to_string(),
      voice:           options.voice,
      instructions:    options.instructions,
      response_format: options.response_format,
      speed:           options.speed,
      extra:           options.extra,
      provider:        provider.to_string(),
    }),
    _ => Err(AppError::new(anyhow::anyhow!("Invalid model: {}", options.model), Some(StatusCode::BAD_REQUEST))),
  }
}
