package deepgram

import (
	"bytes"
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

func TestHandleSpeechTransformsOpenAIRequestForDeepgram(t *testing.T) {
	previousClient := http.DefaultClient
	t.Cleanup(func() { http.DefaultClient = previousClient })

	var upstreamRequest *http.Request
	var upstreamBody []byte
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		upstreamRequest = req.Clone(req.Context())
		upstreamBody, _ = io.ReadAll(req.Body)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"audio/wav"}},
			Body:       io.NopCloser(bytes.NewBufferString("audio-bytes")),
		}, nil
	})}

	request := httptest.NewRequest(http.MethodPost, "/v1/audio/speech", bytes.NewBufferString(`{"model":"deepgram/aura","input":"Hello, World!","voice":"aura-asteria-en"}`))
	request.Header.Set("Authorization", "Bearer deepgram-key")
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
		t.Fatal("expected request to Deepgram")
	}
	if got, want := upstreamRequest.URL.String(), "https://api.deepgram.com/v1/speak?model=aura-asteria-en"; got != want {
		t.Errorf("upstream URL = %q, want %q", got, want)
	}
	if got, want := upstreamRequest.Header.Get("Authorization"), "Token deepgram-key"; got != want {
		t.Errorf("Authorization = %q, want %q", got, want)
	}
	if got, want := upstreamRequest.Header.Get("Accept"), "audio/*"; got != want {
		t.Errorf("Accept = %q, want %q", got, want)
	}
	if got, want := string(upstreamBody), `{"text":"Hello, World!"}`; got != want {
		t.Errorf("upstream body = %s, want %s", got, want)
	}
	if got, want := recorder.Header().Get("Content-Type"), "audio/wav"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}
	if got, want := recorder.Body.String(), "audio-bytes"; got != want {
		t.Errorf("response body = %q, want %q", got, want)
	}
}

func TestHandleSpeechDefaultsDeepgramResponseContentType(t *testing.T) {
	previousClient := http.DefaultClient
	t.Cleanup(func() { http.DefaultClient = previousClient })
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("audio-bytes")),
		}, nil
	})}

	request := httptest.NewRequest(http.MethodPost, "/v1/audio/speech", bytes.NewBufferString(`{"model":"deepgram/aura","input":"Hello","voice":"aura-asteria-en"}`))
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
	if got, want := recorder.Header().Get("Content-Type"), "audio/mpeg"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}
}
