use axum::{
  Json,
  http::StatusCode,
  debug_handler,
};
use serde::{Deserialize, Serialize};

#[derive(Deserialize)]
// https://platform.openai.com/docs/api-reference/audio/createSpeech
pub struct SpeechOptions {
  // The text to generate audio for. The maximum length is 4096 characters.
  input: String,
  // One of the available TTS models.
  model: String,
  // The voice to use when generating the audio.
  voice: String,
  // instructions
  // response_format
  // speed
}

#[derive(Serialize)]
pub struct SpeechResult {
  // The text to generate audio for. The maximum length is 4096 characters.
  input: String,
  // One of the available TTS models.
  model: String,
  // The voice to use when generating the audio.
  voice: String,
  // One of the available TTS providers.
  provider: String,
}

#[debug_handler]
pub async fn speech(
  Json(body): Json<SpeechOptions>,
) -> (StatusCode, Json<SpeechResult>) {
  let vec: Vec<&str> = body.model.split('/').collect();

  let result = SpeechResult {
    input: body.input,
    model: vec[0].to_string(),
    voice: body.voice,
    provider: vec[1].to_string(),
  };

  (StatusCode::OK, Json(result))
}

