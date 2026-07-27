package elevenlabs

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/utils"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestHandleSpeechTransformsOpenAIRequestForElevenLabs(t *testing.T) {
	previousClient := http.DefaultClient
	t.Cleanup(func() { http.DefaultClient = previousClient })

	var upstreamRequest *http.Request
	var upstreamBody []byte
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		upstreamRequest = req.Clone(req.Context())
		upstreamBody, _ = io.ReadAll(req.Body)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"audio/mpeg"}},
			Body:       io.NopCloser(bytes.NewBufferString("audio-bytes")),
		}, nil
	})}

	request := httptest.NewRequest(http.MethodPost, "/v1/audio/speech", bytes.NewBufferString(`{
		"model":"elevenlabs/eleven_multilingual_v2",
		"input":"Hello, World!",
		"voice":"voice-id",
		"speed":1.2,
		"extra_body":{"voice_settings":{"stability":0.5}}
	}`))
	request.Header.Set("Authorization", "Bearer elevenlabs-key")
	options := types.NewSpeechRequestOptions(request.Body)
	if options.IsError() {
		t.Fatalf("NewSpeechRequestOptions returned error: %v", options.Error())
	}

	e := echo.New()
	recorder := httptest.NewRecorder()
	result := HandleSpeech(e.NewContext(request, recorder), utils.ResultToOption(options))
	if result.IsError() {
		t.Fatalf("HandleSpeech returned error: %v", result.Error())
	}

	if upstreamRequest == nil {
		t.Fatal("expected request to ElevenLabs")
	}
	if got, want := upstreamRequest.URL.String(), "https://api.elevenlabs.io/v1/text-to-speech/voice-id"; got != want {
		t.Errorf("upstream URL = %q, want %q", got, want)
	}
	if got, want := upstreamRequest.Header.Get("xi-api-key"), "elevenlabs-key"; got != want {
		t.Errorf("xi-api-key = %q, want %q", got, want)
	}
	var payload map[string]any
	if err := json.Unmarshal(upstreamBody, &payload); err != nil {
		t.Fatalf("unmarshal upstream body: %v", err)
	}
	if got, want := payload["text"], "Hello, World!"; got != want {
		t.Errorf("text = %#v, want %#v", got, want)
	}
	if got, want := payload["model_id"], "eleven_multilingual_v2"; got != want {
		t.Errorf("model_id = %#v, want %#v", got, want)
	}
	if _, present := payload["model"]; present {
		t.Error("upstream payload includes OpenAI-only model field")
	}
	if _, present := payload["voice"]; present {
		t.Error("upstream payload includes OpenAI-only voice field")
	}
	if got, want := payload["voice_settings"], map[string]any{"stability": 0.5}; !mapsEqual(got, want) {
		t.Errorf("voice_settings = %#v, want %#v", got, want)
	}
	if got, want := recorder.Code, http.StatusOK; got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
	if got, want := recorder.Body.String(), "audio-bytes"; got != want {
		t.Errorf("response body = %q, want %q", got, want)
	}
}

func mapsEqual(got, want any) bool {
	gotJSON, gotErr := json.Marshal(got)
	wantJSON, wantErr := json.Marshal(want)
	return gotErr == nil && wantErr == nil && bytes.Equal(gotJSON, wantJSON)
}

func TestHandleSpeechReturnsErrorForUpstreamFailure(t *testing.T) {
	previousClient := http.DefaultClient
	t.Cleanup(func() { http.DefaultClient = previousClient })
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewBufferString(`{"detail":"invalid API key"}`)),
		}, nil
	})}

	request := httptest.NewRequest(http.MethodPost, "/v1/audio/speech", bytes.NewBufferString(`{"model":"elevenlabs/model","input":"Hello","voice":"voice-id"}`))
	options := types.NewSpeechRequestOptions(request.Body)
	if options.IsError() {
		t.Fatalf("NewSpeechRequestOptions returned error: %v", options.Error())
	}

	e := echo.New()
	result := HandleSpeech(e.NewContext(request, httptest.NewRecorder()), utils.ResultToOption(options))
	if !result.IsError() {
		t.Fatal("HandleSpeech returned success for an upstream 401")
	}
}
