package volcengine_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/moeru-ai/unspeech/pkg/backend"
	"github.com/moeru-ai/unspeech/pkg/ho"
)

// TestBidirectionalStream_Integration is gated on VOLCENGINE_API_KEY and
// VOLCENGINE_VOICE. Without both, the test is skipped — it makes a real call
// to openspeech.bytedance.com and bills the project.
//
// Run with:
//
//	VOLCENGINE_API_KEY=sk-... VOLCENGINE_VOICE=zh_female_shuangkuaisisi_moon_bigtts \
//	  go test -tags=integration -run TestBidirectionalStream_Integration ./pkg/backend/volcengine/...
func TestBidirectionalStream_Integration(t *testing.T) {
	apiKey := os.Getenv("VOLCENGINE_API_KEY")
	voice := os.Getenv("VOLCENGINE_VOICE")

	if apiKey == "" || voice == "" {
		t.Skip("VOLCENGINE_API_KEY and VOLCENGINE_VOICE must be set")
	}

	resourceID := os.Getenv("VOLCENGINE_RESOURCE_ID")
	if resourceID == "" {
		resourceID = "seed-tts-2.0"
	}

	e := echo.New()
	e.GET("/v1/audio/speech/stream", ho.MonadEcho1(backend.SpeechStream))

	srv := httptest.NewServer(e)
	defer srv.Close()

	wsURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	wsURL.Scheme = "ws"
	wsURL.Path = "/v1/audio/speech/stream"

	hdr := http.Header{}
	hdr.Set("Authorization", "Bearer "+apiKey)

	dialer := *websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, resp, err := dialer.Dial(wsURL.String(), hdr)
	if resp != nil {
		_ = resp.Body.Close()
	}

	require.NoError(t, err)

	defer func() { _ = conn.Close() }()

	startMsg := map[string]any{
		"event":           "start",
		"model":           "volcengine/seed-tts-2.0",
		"voice":           voice,
		"response_format": "mp3",
		"extra_body": map[string]any{
			"api_resource_id": resourceID,
			"audio": map[string]any{
				"sample_rate": 24000,
				"bit_rate":    160000,
			},
		},
	}

	require.NoError(t, conn.WriteJSON(startMsg))
	require.NoError(t, conn.WriteJSON(map[string]any{"event": "text", "text": "unspeech bidirectional streaming smoke test."}))
	require.NoError(t, conn.WriteJSON(map[string]any{"event": "finish"}))

	deadline := time.Now().Add(30 * time.Second)
	require.NoError(t, conn.SetReadDeadline(deadline))

	var (
		sawSessionStarted  bool
		sawSessionFinished bool
		audioBytes         int
	)

	for !sawSessionFinished {
		msgType, data, readErr := conn.ReadMessage()
		if readErr != nil {
			if websocket.IsCloseError(readErr, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				break
			}

			t.Fatalf("read message: %v (saw_started=%v audio_bytes=%d)", readErr, sawSessionStarted, audioBytes)
		}

		switch msgType {
		case websocket.BinaryMessage:
			audioBytes += len(data)

		case websocket.TextMessage:
			var event struct {
				Event   string `json:"event"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}

			require.NoError(t, json.Unmarshal(data, &event))
			t.Logf("event: %s", strings.TrimSpace(string(data)))

			switch event.Event {
			case "session.started":
				sawSessionStarted = true
			case "session.finished":
				sawSessionFinished = true
			case "error":
				t.Fatalf("server error: code=%s message=%s", event.Code, event.Message)
			}
		}
	}

	require.True(t, sawSessionStarted, "expected session.started")
	require.True(t, sawSessionFinished, "expected session.finished")
	require.Positive(t, audioBytes, "expected at least one audio byte")
}
