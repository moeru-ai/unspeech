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
  // Control the voice of your generated audio with additional instructions.
  pub instructions: Option<String>,
  // The format to audio in.
  pub response_format: Option<String>,
  // The speed of the generated audio.
  pub speed: Option<f32>,
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
  // Control the voice of your generated audio with additional instructions.
  pub instructions: Option<String>,
  // The format to audio in.
  pub response_format: Option<String>,
  // The speed of the generated audio.
  pub speed: Option<f32>,
  #[serde(flatten)]
  pub extra: HashMap<String, Value>,
  // One of the available TTS providers.
  pub provider: String,
}

pub fn process_speech_options(options: SpeechOptions) -> ProcessedSpeechOptions {
  let vec: Vec<&str> = options.model.split('/').collect();

  ProcessedSpeechOptions {
    input: options.input,
    model: vec[1].to_string(),
    voice: options.voice,
    instructions: options.instructions,
    response_format: options.response_format,
    speed: options.speed,
    extra: options.extra,
    provider: vec[0].to_string(),
  }
}
