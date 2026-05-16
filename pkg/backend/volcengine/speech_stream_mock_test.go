package volcengine

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/samber/mo"
	"github.com/stretchr/testify/require"

	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/ho"
)

// mockUpstream stubs the Volcengine v3 bidirectional TTS server so the
// bridge can be driven end-to-end without real credentials. It speaks the
// minimal subset of the v3 protocol the bridge actually uses.
//
// The hooks let each test customise post-TaskRequest behavior — e.g. one
// scenario sends a few audio frames then SessionFinished, another stalls
// audio to give the test time to send `cancel` before the upstream is
// "done".
type mockUpstream struct {
	t *testing.T

	// onTaskRequest is invoked after a TaskRequest text frame arrives. It
	// drives the response side of the protocol (audio + SessionFinished /
	// SessionFailed / etc.). The implementation may block to let the test
	// inject more client-side events.
	onTaskRequest func(conn *websocket.Conn, sessionID string)

	// canceled is set when a CancelSession upstream frame is observed; tests
	// can read it via .CanceledOK() / .DidCancel().
	mu       sync.Mutex
	canceled bool
	finished bool

	server *httptest.Server
}

func (m *mockUpstream) URL() string {
	return strings.Replace(m.server.URL, "http://", "ws://", 1)
}

func (m *mockUpstream) DidCancel() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.canceled
}

func (m *mockUpstream) DidFinish() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.finished
}

func (m *mockUpstream) Close() {
	m.server.Close()
}

func newMockUpstream(t *testing.T, onTaskRequest func(conn *websocket.Conn, sessionID string)) *mockUpstream {
	t.Helper()

	m := &mockUpstream{t: t, onTaskRequest: onTaskRequest}
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	handler := func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("mock upstream upgrade: %v", err)

			return
		}

		defer conn.Close()

		// Drive the v3 protocol manually.
		sessionID := ""

		for {
			_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))

			msgType, data, readErr := conn.ReadMessage()
			if readErr != nil {
				return
			}

			if msgType != websocket.BinaryMessage {
				continue
			}

			frame, decodeErr := decodeV3Frame(data)
			if decodeErr != nil {
				t.Errorf("mock upstream decode: %v", decodeErr)

				return
			}

			switch frame.event {
			case v3EventStartConnection:
				// Reply ConnectionStarted with a synthetic connection id.
				replyFrame := v3Frame{
					msgType: v3MsgTypeFullClient | 0b1000, // server-response flag bit
					flags:   v3FlagWithEvent,
					serial:  v3SerialJSON,
					event:   v3EventConnectionStarted,
					payload: []byte(`{}`),
					// Mock upstream uses a fixed connection id; decodeV3Frame
					// reads it but does not retain it past handshake. The
					// connection_id field in encodeV3Frame is keyed off
					// hasConnectionID(event) which is true for
					// ConnectionStarted/Failed/Finished — but our encoder
					// looks at sessionID, not connectionID. Send no
					// connection-id payload prefix and let the decoder treat
					// the next 4 bytes as the connection-id size (0). See
					// encodeV3Frame: we set sessionID to "" and the
					// hasSessionID check is false for connection events, so
					// the encoder skips the session_id field and we manually
					// inject a zero-length connection id. We work around the
					// public encoder by hand-building this frame.
				}

				// encodeV3Frame doesn't handle connection_id, so build the
				// frame inline matching the wire layout (header + event +
				// conn_id_size=0 + payload_size + payload).
				bin := buildConnEventFrame(replyFrame.event, []byte(`{}`))
				if writeErr := conn.WriteMessage(websocket.BinaryMessage, bin); writeErr != nil {
					t.Errorf("mock upstream write: %v", writeErr)

					return
				}

			case v3EventStartSession:
				sessionID = frame.sessionID
				replyBin := encodeV3Frame(v3Frame{
					msgType:   v3MsgTypeFullClient | 0b1000,
					flags:     v3FlagWithEvent,
					serial:    v3SerialJSON,
					event:     v3EventSessionStarted,
					sessionID: sessionID,
					payload:   []byte(`{}`),
				})
				if writeErr := conn.WriteMessage(websocket.BinaryMessage, replyBin); writeErr != nil {
					t.Errorf("mock upstream write: %v", writeErr)

					return
				}

			case v3EventTaskRequest:
				if m.onTaskRequest != nil {
					m.onTaskRequest(conn, sessionID)
				}

			case v3EventFinishSession:
				m.mu.Lock()
				m.finished = true
				m.mu.Unlock()

			case v3EventCancelSession:
				m.mu.Lock()
				m.canceled = true
				m.mu.Unlock()

				// Reply SessionCanceled then close.
				replyBin := encodeV3Frame(v3Frame{
					msgType:   v3MsgTypeFullClient | 0b1000,
					flags:     v3FlagWithEvent,
					serial:    v3SerialJSON,
					event:     v3EventSessionCanceled,
					sessionID: sessionID,
					payload:   []byte(`{}`),
				})
				_ = conn.WriteMessage(websocket.BinaryMessage, replyBin)

				return

			case v3EventFinishConnection:
				return
			}
		}
	}

	m.server = httptest.NewServer(http.HandlerFunc(handler))

	return m
}

// buildConnEventFrame is a one-off helper for connection-level frames the
// public encoder doesn't handle (it only emits session_id-bearing
// frames). Returns the wire bytes for ConnectionStarted/Failed/Finished
// with a zero-length connection id and the supplied payload.
func buildConnEventFrame(event v3Event, payload []byte) []byte {
	header := []byte{
		(1 << 4) | 1, // version=1, header_size=1*4
		(byte(v3MsgTypeFullClient|0b1000) << 4) | v3FlagWithEvent,
		(v3SerialJSON << 4) | 0,
		0,
	}

	buf := make([]byte, 0, len(header)+4+4+4+len(payload))
	buf = append(buf, header...)

	// event (int32 big-endian)
	buf = append(buf, byte(event>>24), byte(event>>16), byte(event>>8), byte(event))

	// connection_id size = 0
	buf = append(buf, 0, 0, 0, 0)

	// payload size + payload
	pSize := uint32(len(payload)) //nolint:gosec
	buf = append(buf, byte(pSize>>24), byte(pSize>>16), byte(pSize>>8), byte(pSize))
	buf = append(buf, payload...)

	return buf
}

// dialBridge fronts unspeech's SpeechStream route in an httptest server
// and returns a client ws connection ready to send `start`.
func dialBridge(t *testing.T, upstreamURL string) *websocket.Conn {
	t.Helper()

	prev := v3BidirectionEndpoint
	v3BidirectionEndpoint = upstreamURL

	t.Cleanup(func() { v3BidirectionEndpoint = prev })

	e := echo.New()
	e.HideBanner = true
	// We need backend.SpeechStream wired but importing the public package
	// here would be circular; the unspeech command wires it at startup, so
	// for the in-package test we use the volcengine bridge directly via a
	// thin echo handler that mimics what SpeechStream does post-upgrade.
	e.GET("/stream", ho.MonadEcho1(testSpeechStreamRoute))

	srv := httptest.NewServer(e)

	t.Cleanup(srv.Close)

	wsURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	wsURL.Scheme = "ws"
	wsURL.Path = "/stream"

	hdr := http.Header{}
	hdr.Set("Authorization", "Bearer mock-api-key-for-test")

	dialer := *websocket.DefaultDialer
	dialer.HandshakeTimeout = 5 * time.Second
	conn, resp, dialErr := dialer.Dial(wsURL.String(), hdr)
	if resp != nil {
		_ = resp.Body.Close()
	}

	require.NoError(t, dialErr)

	t.Cleanup(func() { _ = conn.Close() })

	return conn
}

// readClientEvents drains ws frames until either session.finished, an
// error event, or the conn closes. Returns observed events in order plus
// a count of binary (audio) bytes.
type clientEvent struct {
	Event   string `json:"event"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func drainClient(t *testing.T, conn *websocket.Conn, deadline time.Time) (events []clientEvent, audioBytes int) {
	t.Helper()

	_ = conn.SetReadDeadline(deadline)

	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			var closeErr *websocket.CloseError
			if errors.As(err, &closeErr) {
				return events, audioBytes
			}

			return events, audioBytes
		}

		switch msgType {
		case websocket.BinaryMessage:
			audioBytes += len(data)
		case websocket.TextMessage:
			var ev clientEvent

			if err := json.Unmarshal(data, &ev); err == nil {
				events = append(events, ev)

				if ev.Event == "session.finished" || ev.Event == "error" {
					return events, audioBytes
				}
			}
		}
	}
}

// TestBridge_FinishWaitsForUpstreamCompletion regression-tests the
// codex CRITICAL fix: `finish` from the client must NOT tear the bridge
// down before upstream drains audio + SessionFinished. Pre-fix the
// client received zero or a partial audio prefix then close 1006.
func TestBridge_FinishWaitsForUpstreamCompletion(t *testing.T) {
	const totalAudioBytes = 4096
	chunks := 4
	chunkSize := totalAudioBytes / chunks

	upstream := newMockUpstream(t, func(conn *websocket.Conn, sessionID string) {
		// Send N audio frames, then SessionFinished. The audio frames have
		// staggered writes to simulate streaming over wall-clock time —
		// if the bridge tears down on `finish` instead of waiting, only
		// the first few chunks would be forwarded.
		for i := 0; i < chunks; i++ {
			frame := encodeV3Frame(v3Frame{
				msgType:   v3MsgTypeFullClient | 0b1000, // server-response
				flags:     v3FlagWithEvent,
				serial:    0, // raw audio
				event:     v3EventTTSResponse,
				sessionID: sessionID,
				payload:   make([]byte, chunkSize),
			})
			_ = conn.WriteMessage(websocket.BinaryMessage, frame)

			time.Sleep(50 * time.Millisecond)
		}

		usagePayload := `{"status_code":20000000,"message":"ok","usage":{"text_words":42}}`
		finalFrame := encodeV3Frame(v3Frame{
			msgType:   v3MsgTypeFullClient | 0b1000,
			flags:     v3FlagWithEvent,
			serial:    v3SerialJSON,
			event:     v3EventSessionFinished,
			sessionID: sessionID,
			payload:   []byte(usagePayload),
		})
		_ = conn.WriteMessage(websocket.BinaryMessage, finalFrame)
	})

	defer upstream.Close()

	conn := dialBridge(t, upstream.URL())

	require.NoError(t, conn.WriteJSON(map[string]any{
		"event":           "start",
		"model":           "volcengine/seed-tts-2.0",
		"voice":           "mock-voice",
		"response_format": "mp3",
	}))
	require.NoError(t, conn.WriteJSON(map[string]any{"event": "text", "text": "hello"}))
	require.NoError(t, conn.WriteJSON(map[string]any{"event": "finish"}))

	events, audioBytes := drainClient(t, conn, time.Now().Add(5*time.Second))

	require.NotEmpty(t, events, "expected at least one server event")

	var sawStarted, sawFinished bool
	for _, ev := range events {
		switch ev.Event {
		case "session.started":
			sawStarted = true
		case "session.finished":
			sawFinished = true
		case "error":
			t.Fatalf("unexpected error event: code=%s message=%s", ev.Code, ev.Message)
		}
	}

	require.True(t, sawStarted, "expected session.started before finish")
	require.True(t, sawFinished, "expected session.finished — finish must wait for upstream completion")
	require.GreaterOrEqual(t, audioBytes, totalAudioBytes,
		"expected all %d audio bytes; got %d — `finish` likely truncated the upstream", totalAudioBytes, audioBytes)
}

// TestBridge_CancelAfterFinish regression-tests codex follow-up #2.
// Stage.vue sends `finish` immediately after `text`, so user-initiated
// aborts later in the session land as `cancel` in the post-finish drain
// state. Pre-fix the drain state silently swallowed it.
//
// Wire ordering this test exercises:
//
//	client → bridge:    start
//	client → bridge:    text("abortable")
//	bridge → upstream:  TaskRequest
//	client → bridge:    finish              ← bridge enters drain state
//	bridge → upstream:  FinishSession
//	client → bridge:    cancel              ← pre-fix: swallowed by drain
//	bridge → upstream:  CancelSession        ← must reach upstream (the assertion)
func TestBridge_CancelAfterFinish(t *testing.T) {
	// Upstream signals TaskRequest receipt but does not stall (so the
	// mock's read loop keeps draining subsequent FinishSession +
	// CancelSession frames from the bridge). We don't actually need any
	// audio for this test — we're only verifying that CancelSession
	// reaches upstream.
	taskReached := make(chan struct{})
	upstream := newMockUpstream(t, func(_ *websocket.Conn, _ string) {
		close(taskReached)
	})

	defer upstream.Close()

	conn := dialBridge(t, upstream.URL())

	require.NoError(t, conn.WriteJSON(map[string]any{
		"event": "start", "model": "volcengine/seed-tts-2.0", "voice": "mock-voice",
	}))
	require.NoError(t, conn.WriteJSON(map[string]any{"event": "text", "text": "abortable"}))
	require.NoError(t, conn.WriteJSON(map[string]any{"event": "finish"}))

	// Wait for upstream to confirm it got the TaskRequest before issuing
	// cancel, otherwise we race the bridge's startedCh handshake.
	select {
	case <-taskReached:
	case <-time.After(2 * time.Second):
		t.Fatal("upstream never received TaskRequest")
	}

	// Now send cancel — this lands in the post-finish drain branch.
	require.NoError(t, conn.WriteJSON(map[string]any{"event": "cancel"}))

	// The bridge should forward CancelSession upstream. Poll up to 3s.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && !upstream.DidCancel() {
		time.Sleep(10 * time.Millisecond)
	}

	require.True(t, upstream.DidCancel(),
		"upstream did not receive CancelSession — `cancel` was swallowed in post-finish drain")
}

// TestBridge_ErrorEventOnMissingApiKey is a fast regression check that
// post-upgrade auth failures surface as JSON `error` events rather than
// HTTP 500 stack traces written into the hijacked socket (smoke test
// finding from this session).
func TestBridge_ErrorEventOnMissingApiKey(t *testing.T) {
	// Mock upstream is never reached; bridge fails at auth-header read.
	upstream := newMockUpstream(t, func(_ *websocket.Conn, _ string) {})
	defer upstream.Close()

	prev := v3BidirectionEndpoint
	v3BidirectionEndpoint = upstream.URL()

	t.Cleanup(func() { v3BidirectionEndpoint = prev })

	e := echo.New()
	e.HideBanner = true
	e.GET("/stream", ho.MonadEcho1(testSpeechStreamRoute))

	srv := httptest.NewServer(e)

	defer srv.Close()

	wsURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	wsURL.Scheme = "ws"
	wsURL.Path = "/stream"

	dialer := *websocket.DefaultDialer
	dialer.HandshakeTimeout = 5 * time.Second
	// NO Authorization header — should produce a `missing_api_key` event.
	conn, resp, dialErr := dialer.Dial(wsURL.String(), nil)
	if resp != nil {
		_ = resp.Body.Close()
	}

	require.NoError(t, dialErr)

	defer func() { _ = conn.Close() }()

	require.NoError(t, conn.WriteJSON(map[string]any{
		"event": "start", "model": "volcengine/seed-tts-2.0", "voice": "mock-voice",
	}))

	events, _ := drainClient(t, conn, time.Now().Add(2*time.Second))

	require.NotEmpty(t, events)
	require.Equal(t, "error", events[0].Event)
	require.Equal(t, "missing_api_key", events[0].Code)
}

// testSpeechStreamRoute is the in-package entry point for the tests.
// It mirrors what `backend.SpeechStream` does post-upgrade, but lives
// inside `package volcengine` so it can swap v3BidirectionEndpoint
// without import cycles (backend imports volcengine; volcengine cannot
// import backend).
func testSpeechStreamRoute(c echo.Context) mo.Result[any] {
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return mo.Ok[any](nil)
	}

	defer func() { _ = ws.Close() }()

	_, payload, err := ws.ReadMessage()
	if err != nil {
		return mo.Ok[any](nil)
	}

	opts := types.NewSpeechStreamStartOptions(payload)
	if opts.IsError() {
		return mo.Ok[any](nil)
	}

	HandleSpeechStream(c, ws, mo.Some(opts.MustGet()))

	return mo.Ok[any](nil)
}
