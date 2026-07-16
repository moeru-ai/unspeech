package fishaudio

import (
	"strings"
	"testing"

	"github.com/moeru-ai/unspeech/pkg/backend/types"
)

func TestBuildSpeechRequestMapsOpenAIFields(t *testing.T) {
	t.Parallel()

	req, err := buildSpeechRequest(types.SpeechRequestOptions{
		OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{
			Model:          "fishaudio/" + modelS1,
			Input:          "Hello from unSpeech",
			Voice:          "802e3bc2b27e49c2995d23ef70e6ac89",
			ResponseFormat: "mp3",
			Speed:          1.2,
		},
		Model: modelS1,
	})

	if err != nil {
		t.Fatalf("buildSpeechRequest returned error: %v", err)
	}
	if req.Text != "Hello from unSpeech" {
		t.Fatalf("Text = %q", req.Text)
	}
	if req.ReferenceID != "802e3bc2b27e49c2995d23ef70e6ac89" {
		t.Fatalf("ReferenceID = %q", req.ReferenceID)
	}
	if req.Format != "mp3" {
		t.Fatalf("Format = %q, want mp3", req.Format)
	}
	if req.Prosody == nil || req.Prosody.Speed != 1.2 {
		t.Fatalf("Prosody = %+v, want speed 1.2", req.Prosody)
	}
}

func TestBuildSpeechRequestDefaultsToMP3(t *testing.T) {
	t.Parallel()

	req, err := buildSpeechRequest(types.SpeechRequestOptions{
		OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{
			Input: "hello",
			Voice: "some-reference-id",
		},
		Model: modelS1,
	})

	if err != nil {
		t.Fatalf("buildSpeechRequest returned error: %v", err)
	}
	if req.Format != "mp3" {
		t.Fatalf("Format = %q, want mp3", req.Format)
	}
	if req.Prosody != nil {
		t.Fatalf("Prosody = %+v, want nil when speed is unset", req.Prosody)
	}
}

func TestBuildSpeechRequestMapsExtraBody(t *testing.T) {
	t.Parallel()

	req, err := buildSpeechRequest(types.SpeechRequestOptions{
		OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{
			Input: "hello",
			Voice: "some-reference-id",
			ExtraBody: map[string]any{
				"chunk_length": 150,
				"normalize":    false,
				"latency":      "balanced",
				"temperature":  0.8,
				"top_p":        0.9,
				"mp3_bitrate":  192,
			},
		},
		Model: modelS1,
	})

	if err != nil {
		t.Fatalf("buildSpeechRequest returned error: %v", err)
	}
	if req.ChunkLength == nil || *req.ChunkLength != 150 {
		t.Fatalf("ChunkLength = %v, want 150", req.ChunkLength)
	}
	if req.Normalize == nil || *req.Normalize != false {
		t.Fatalf("Normalize = %v, want false", req.Normalize)
	}
	if req.Latency != "balanced" {
		t.Fatalf("Latency = %q, want balanced", req.Latency)
	}
	if req.Temperature == nil || *req.Temperature != 0.8 {
		t.Fatalf("Temperature = %v, want 0.8", req.Temperature)
	}
	if req.TopP == nil || *req.TopP != 0.9 {
		t.Fatalf("TopP = %v, want 0.9", req.TopP)
	}
	if req.MP3Bitrate == nil || *req.MP3Bitrate != 192 {
		t.Fatalf("MP3Bitrate = %v, want 192", req.MP3Bitrate)
	}
}

func TestBuildSpeechRequestExtraProsodyOverridesSpeed(t *testing.T) {
	t.Parallel()

	req, err := buildSpeechRequest(types.SpeechRequestOptions{
		OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{
			Input: "hello",
			Voice: "some-reference-id",
			Speed: 1.5,
			ExtraBody: map[string]any{
				"prosody": map[string]any{
					"speed":  0.8,
					"volume": -3.0,
				},
			},
		},
		Model: modelS1,
	})

	if err != nil {
		t.Fatalf("buildSpeechRequest returned error: %v", err)
	}
	if req.Prosody == nil {
		t.Fatal("Prosody = nil")
	}
	if req.Prosody.Speed != 0.8 {
		t.Fatalf("Prosody.Speed = %v, want extra_body prosody to win over OpenAI speed", req.Prosody.Speed)
	}
	if req.Prosody.Volume != -3.0 {
		t.Fatalf("Prosody.Volume = %v, want -3.0", req.Prosody.Volume)
	}
}

func TestBuildSpeechRequestRejectsUnsupportedFormats(t *testing.T) {
	t.Parallel()

	for _, format := range []string{"aac", "flac"} {
		_, err := buildSpeechRequest(types.SpeechRequestOptions{
			OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{
				Input:          "hello",
				Voice:          "some-reference-id",
				ResponseFormat: format,
			},
			Model: modelS1,
		})

		if err == nil {
			t.Fatalf("buildSpeechRequest(%q) returned nil error", format)
		}
		if !strings.Contains(err.Error(), "supports only") {
			t.Fatalf("error = %q", err.Error())
		}
	}
}

func TestModelForRequestFallsBackToS1(t *testing.T) {
	t.Parallel()

	if got := modelForRequest(""); got != modelS1 {
		t.Fatalf("modelForRequest(\"\") = %q, want %q", got, modelS1)
	}
	if got := modelForRequest(modelS21Pro); got != modelS21Pro {
		t.Fatalf("modelForRequest(%q) = %q", modelS21Pro, got)
	}
}
