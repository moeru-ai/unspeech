package minimax

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/samber/mo"
)

const (
	testInput        = "hello"
	testVoice        = "test-voice"
	testMinimaxModel = "speech-2.8-turbo"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

type panicReader struct{}

func (panicReader) Read([]byte) (int, error) {
	panic("reader must not be consumed while selecting an SSE decoder")
}

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type signalingRecorder struct {
	*httptest.ResponseRecorder

	wrote chan struct{}
	once  sync.Once
}

func (r *signalingRecorder) Write(data []byte) (int, error) {
	written, err := r.ResponseRecorder.Write(data)
	r.once.Do(func() { close(r.wrote) })

	return written, err
}

func TestHandleSpeechMapsOpenAIFields(t *testing.T) {
	var captured TTSRequest

	setDefaultClient(t, func(req *http.Request) (*http.Response, error) {
		decodeErr := json.NewDecoder(req.Body).Decode(&captured)
		if decodeErr != nil {
			t.Fatalf("decode upstream request: %v", decodeErr)
		}

		return jsonResponse(http.StatusOK, `{"data":{"audio":"0102"},"base_resp":{"status_code":0}}`), nil
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/audio/speech", nil)
	context := echo.New().NewContext(request, recorder)
	result := HandleSpeech(context, mo.Some(types.SpeechRequestOptions{
		OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{
			Input:          testInput,
			Voice:          testVoice,
			Speed:          1.25,
			ResponseFormat: audioFormatWAV,
			ExtraBody:      map[string]any{"sample_rate": 24000},
		},
		Model: testMinimaxModel,
	}))

	if result.IsError() {
		t.Fatalf("HandleSpeech returned error: %v", result.Error())
	}
	if captured.VoiceSetting == nil || captured.VoiceSetting.Speed != 1.25 {
		t.Fatalf("VoiceSetting = %+v, want speed 1.25", captured.VoiceSetting)
	}
	if captured.AudioSetting == nil || captured.AudioSetting.Format != "wav" || captured.AudioSetting.SampleRate != 24000 {
		t.Fatalf("AudioSetting = %+v, want format wav and sample rate 24000", captured.AudioSetting)
	}
	if got := recorder.Header().Get(echo.HeaderContentType); got != contentTypeWAV {
		t.Fatalf("Content-Type = %q, want audio/wav", got)
	}
}

func TestHandleSpeechMapsChannelOption(t *testing.T) {
	var captured TTSRequest

	setDefaultClient(t, func(req *http.Request) (*http.Response, error) {
		decodeErr := json.NewDecoder(req.Body).Decode(&captured)
		if decodeErr != nil {
			t.Fatalf("decode upstream request: %v", decodeErr)
		}

		return jsonResponse(http.StatusOK, `{"data":{"audio":"0102"},"base_resp":{"status_code":0}}`), nil
	})

	context := echo.New().NewContext(
		httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/audio/speech", nil),
		httptest.NewRecorder(),
	)
	result := HandleSpeech(context, mo.Some(types.SpeechRequestOptions{
		OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{
			Input:     testInput,
			Voice:     testVoice,
			ExtraBody: map[string]any{"channel": 2},
		},
		Model: testMinimaxModel,
	}))

	if result.IsError() {
		t.Fatalf("HandleSpeech returned error: %v", result.Error())
	}
	if captured.AudioSetting == nil || captured.AudioSetting.Channel != 2 {
		t.Fatalf("AudioSetting = %+v, want channel 2", captured.AudioSetting)
	}
}

func TestHandleStreamingSpeechRejectsIncompleteResponseBeforeAudio(t *testing.T) {
	setDefaultClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{"data":{"status":1},"base_resp":{"status_code":0}}`), nil
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/audio/speech", nil)
	context := echo.New().NewContext(request, recorder)
	result := handleStreamingSpeech(context, "test-token", types.SpeechRequestOptions{
		OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{Input: testInput, Voice: testVoice},
		Model:                      testMinimaxModel,
	})

	if result.IsOk() {
		t.Fatal("handleStreamingSpeech accepted EOF before status=2")
	}
}

func TestHandleStreamingSpeechAbortsIncompleteResponseAfterAudio(t *testing.T) {
	setDefaultClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{"data":{"audio":"0102","status":1},"base_resp":{"status_code":0}}`), nil
	})

	e := echo.New()
	e.POST("/", func(c echo.Context) error {
		result := handleStreamingSpeech(c, "test-token", types.SpeechRequestOptions{
			OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{Input: testInput, Voice: testVoice},
			Model:                      testMinimaxModel,
		})
		if result.IsError() {
			return result.Error()
		}

		return nil
	})

	server := httptest.NewServer(e)
	t.Cleanup(server.Close)

	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("create streaming speech request: %v", err)
	}

	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatalf("request streaming speech: %v", err)
	}

	defer func() { _ = response.Body.Close() }()

	_, readErr := io.ReadAll(response.Body)
	if readErr == nil {
		t.Fatal("incomplete committed stream ended cleanly instead of aborting the response body")
	}
}

func TestHandleVoicesPreservesUnknownContentTypeErrorStatus(t *testing.T) {
	setDefaultClient(t, func(req *http.Request) (*http.Response, error) {
		response := jsonResponse(http.StatusTooManyRequests, "rate limited")
		response.Header.Set(echo.HeaderContentType, "application/octet-stream")

		return response, nil
	})

	context := echo.New().NewContext(
		httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/voices?provider=minimax", nil),
		httptest.NewRecorder(),
	)

	result := HandleVoices(context, mo.Some(types.VoicesRequestOptions{Backend: "minimax"}))
	if result.IsOk() {
		t.Fatal("HandleVoices accepted an upstream 429 response")
	}

	var apiErr *apierrors.Error
	if !errors.As(result.Error(), &apiErr) {
		t.Fatalf("error type = %T, want *apierrors.Error", result.Error())
	}
	if apiErr.Status != http.StatusTooManyRequests {
		t.Fatalf("error status = %d, want %d", apiErr.Status, http.StatusTooManyRequests)
	}
}

func TestHandleStreamingSpeechWritesEachChunkBeforeUpstreamCompletes(t *testing.T) {
	upstreamReader, upstreamWriter := io.Pipe()
	firstChunkSent := make(chan struct{})
	releaseFinalChunk := make(chan struct{})

	setDefaultClient(t, func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       upstreamReader,
		}, nil
	})

	go func() {
		_, _ = fmt.Fprintln(upstreamWriter, `{"data":{"audio":"0102","status":1},"base_resp":{"status_code":0}}`)

		close(firstChunkSent)
		<-releaseFinalChunk

		_, _ = fmt.Fprintln(upstreamWriter, `{"data":{"audio":"0304","status":1},"base_resp":{"status_code":0}}`)
		_, _ = fmt.Fprintln(upstreamWriter, `{"data":{"audio":"01020304","status":2},"base_resp":{"status_code":0}}`)
		_ = upstreamWriter.Close()
	}()

	recorder := &signalingRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		wrote:            make(chan struct{}),
	}
	e := echo.New()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/audio/speech", nil)
	context := e.NewContext(request, recorder)
	resultDone := make(chan error, 1)

	go func() {
		result := handleStreamingSpeech(context, "test-token", types.SpeechRequestOptions{
			OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{
				Input: testInput,
				Voice: testVoice,
			},
			Model: testMinimaxModel,
		})
		if result.IsError() {
			resultDone <- result.Error()
			return
		}

		resultDone <- nil
	}()

	<-firstChunkSent

	select {
	case <-recorder.wrote:
	case <-time.After(500 * time.Millisecond):
		close(releaseFinalChunk)
		<-resultDone
		t.Fatal("first audio chunk was buffered until the upstream response completed")
	}

	if got := recorder.Body.Bytes(); string(got) != string([]byte{0x01, 0x02}) {
		close(releaseFinalChunk)
		<-resultDone
		t.Fatalf("first streamed chunk = %v, want [1 2]", got)
	}

	close(releaseFinalChunk)

	resultErr := <-resultDone
	if resultErr != nil {
		t.Fatalf("handleStreamingSpeech returned error: %v", resultErr)
	}
	if got := recorder.Body.Bytes(); string(got) != string([]byte{0x01, 0x02, 0x03, 0x04}) {
		t.Fatalf("streamed audio = %v, want [1 2 3 4]", got)
	}
}

func TestHandleStreamingSpeechDecodesSSEDataFrames(t *testing.T) {
	setDefaultClient(t, func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{echo.HeaderContentType: []string{"text/event-stream"}},
			Body: io.NopCloser(strings.NewReader(
				"data: {\"data\":{\"audio\":\"0102\",\"status\":1},\"base_resp\":{\"status_code\":0}}\n\n" +
					"data: {\"data\":{\"audio\":\"0304\",\"status\":1},\"base_resp\":{\"status_code\":0}}\n\n" +
					"data: {\"data\":{\"audio\":\"01020304\",\"status\":2},\"base_resp\":{\"status_code\":0}}\n\n",
			)),
		}, nil
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/audio/speech", nil)
	context := echo.New().NewContext(request, recorder)
	result := handleStreamingSpeech(context, "test-token", types.SpeechRequestOptions{
		OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{Input: testInput, Voice: testVoice},
		Model:                      testMinimaxModel,
	})

	if result.IsError() {
		t.Fatalf("handleStreamingSpeech returned error: %v", result.Error())
	}
	if got := recorder.Body.Bytes(); string(got) != string([]byte{0x01, 0x02, 0x03, 0x04}) {
		t.Fatalf("streamed SSE audio = %v, want [1 2 3 4]", got)
	}
}

func TestBuildTTSRequestExcludesAggregatedStreamAudio(t *testing.T) {
	reqBody := buildTTSRequest(types.SpeechRequestOptions{
		OpenAISpeechRequestOptions: types.OpenAISpeechRequestOptions{Input: testInput, Voice: testVoice},
		Model:                      testMinimaxModel,
	}, true)

	encoded, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal TTS request: %v", err)
	}
	if !strings.Contains(string(encoded), `"stream_options":{"exclude_aggregated_audio":true}`) {
		t.Fatalf("streaming request = %s, want exclude_aggregated_audio=true", encoded)
	}
}

func TestNewTTSResponseDecoderUsesContentTypeWithoutReading(t *testing.T) {
	decoder := newTTSResponseDecoder(panicReader{}, "text/event-stream; charset=utf-8")
	if _, ok := decoder.(*sseTTSResponseDecoder); !ok {
		t.Fatalf("decoder type = %T, want *sseTTSResponseDecoder", decoder)
	}
}

func setDefaultClient(t *testing.T, roundTrip roundTripFunc) {
	t.Helper()

	originalClient := http.DefaultClient

	t.Cleanup(func() { http.DefaultClient = originalClient })

	http.DefaultClient = &http.Client{Transport: roundTrip}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{echo.HeaderContentType: []string{echo.MIMEApplicationJSON}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
