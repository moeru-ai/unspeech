use std::collections::HashMap;

use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Deserialize)]
// https://platform.openai.com/docs/api-reference/audio/createSpeech
pub struct SpeechOptions {
  // The text to generate audio for. The maximum length is 4096 characters.
  pub input: String,
  // One of the available TTS models.
  pub model: String,
  // The voice to use when generating the audio.
  pub voice: String,
  // instructions
  // response_format
  // speed
  #[serde(flatten)]
  pub extra: HashMap<String, Value>,
}

#[derive(Serialize)]
pub struct ProcessedSpeechOptions {
  // The text to generate audio for. The maximum length is 4096 characters.
  pub input: String,
  // One of the available TTS models.
  pub model: String,
  // The voice to use when generating the audio.
  pub voice: String,
  // instructions
  // response_format
  // speed
  #[serde(flatten)]
  pub extra: HashMap<String, Value>,
  // One of the available TTS providers.
  pub provider: String,
}

pub fn process_speech_options(options: SpeechOptions) -> ProcessedSpeechOptions {
  let vec: Vec<&str> = options.model.split('/').collect();

  ProcessedSpeechOptions {
    input: options.input,
    model: vec[0].to_string(),
    voice: options.voice,
    extra: options.extra,
    provider: vec[1].to_string(),
  }
}
