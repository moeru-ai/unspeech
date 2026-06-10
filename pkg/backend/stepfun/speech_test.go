package stepfun

import (
	"strings"
	"testing"

	"github.com/moeru-ai/unspeech/pkg/backend/types"
)

func TestBuildSpeechRequestForStepAudio25(t *testing.T) {
	t.Parallel()

	req, err := buildSpeechRequest(types.SpeechRequestOptions{
		OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{
			Model:          "stepfun/" + modelStepAudio25TTS,
			Input:          "（轻声）你好",
			Voice:          "cixingnansheng",
			ResponseFormat: "mp3",
			Speed:          1.2,
			ExtraBody: map[string]any{
				"instruction": "温柔、克制、有一点笑意",
				"volume":      1.1,
				"sample_rate": 24000,
			},
		},
		Model: modelStepAudio25TTS,
	})

	if err != nil {
		t.Fatalf("buildSpeechRequest returned error: %v", err)
	}
	if req.Model != modelStepAudio25TTS {
		t.Fatalf("Model = %q, want %q", req.Model, modelStepAudio25TTS)
	}
	if req.Instruction != "温柔、克制、有一点笑意" {
		t.Fatalf("Instruction = %q", req.Instruction)
	}
	if req.Speed != 1.2 {
		t.Fatalf("Speed = %v, want 1.2", req.Speed)
	}
	if req.Volume == nil || *req.Volume != 1.1 {
		t.Fatalf("Volume = %v, want 1.1", req.Volume)
	}
	if req.SampleRate == nil || *req.SampleRate != 24000 {
		t.Fatalf("SampleRate = %v, want 24000", req.SampleRate)
	}
}

func TestBuildSpeechRequestRejectsStepAudio25VoiceLabel(t *testing.T) {
	t.Parallel()

	_, err := buildSpeechRequest(types.SpeechRequestOptions{
		OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{
			Input: "hello",
			Voice: "cixingnansheng",
			ExtraBody: map[string]any{
				"voice_label": map[string]any{"emotion": "高兴"},
			},
		},
		Model: modelStepAudio25TTS,
	})

	if err == nil {
		t.Fatal("buildSpeechRequest returned nil error")
	}
	if !strings.Contains(err.Error(), "does not support voice_label") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestBuildSpeechRequestRejectsInstructionForLegacyModels(t *testing.T) {
	t.Parallel()

	_, err := buildSpeechRequest(types.SpeechRequestOptions{
		OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{
			Input: "hello",
			Voice: "cixingnansheng",
			ExtraBody: map[string]any{
				"instruction": "温柔",
			},
		},
		Model: modelStepTTS2,
	})

	if err == nil {
		t.Fatal("buildSpeechRequest returned nil error")
	}
	if !strings.Contains(err.Error(), "instruction is only supported") {
		t.Fatalf("error = %q", err.Error())
	}
}
