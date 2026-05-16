package volcengine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/samber/mo"

	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/utils"
)

// v3BidirectionEndpoint is the Volcengine v3 bidirectional TTS endpoint.
// Declared as `var` (not `const`) so the in-package mock-upstream
// integration tests can swap it for a local httptest server. The runtime
// value is never mutated outside tests.
var v3BidirectionEndpoint = "wss://openspeech.bytedance.com/api/v3/tts/bidirection" //nolint:gochecknoglobals

type v3ReqParamsAudio struct {
	Format          string  `json:"format,omitempty"`
	SampleRate      int     `json:"sample_rate,omitempty"`
	BitRate         int     `json:"bit_rate,omitempty"`
	Emotion         string  `json:"emotion,omitempty"`
	EmotionScale    float64 `json:"emotion_scale,omitempty"`
	SpeechRate      int     `json:"speech_rate,omitempty"`
	LoudnessRate    int     `json:"loudness_rate,omitempty"`
	EnableTimestamp bool    `json:"enable_timestamp,omitempty"`
	EnableSubtitle  bool    `json:"enable_subtitle,omitempty"`
}

type v3ReqParams struct {
	Text         string            `json:"text"`
	Model        string            `json:"model,omitempty"`
	Speaker      string            `json:"speaker"`
	AudioParams  *v3ReqParamsAudio `json:"audio_params,omitempty"`
	Additions    string            `json:"additions,omitempty"`
	SectionID    string            `json:"section_id,omitempty"`
	ContextTexts []string          `json:"context_texts,omitempty"`
}

type v3SessionPayload struct {
	User      map[string]any `json:"user,omitempty"`
	Event     v3Event        `json:"event"`
	Namespace string         `json:"namespace,omitempty"`
	ReqParams v3ReqParams    `json:"req_params"`
}

// HandleSpeechStream bridges a client WebSocket to Volcengine V3 bidirectional
// TTS. The caller is expected to have already upgraded the HTTP connection and
// consumed the `start` frame.
func HandleSpeechStream(c echo.Context, clientWS *websocket.Conn, options mo.Option[types.SpeechRequestOptions]) mo.Result[any] {
	opts := options.MustGet()

	// Post-upgrade error surface: all failures from here on go to the client
	// as JSON `error` events plus a clean close frame. Echo's HTTP error
	// middleware MUST NOT see them (the connection is hijacked; writing an
	// HTTP response over it dumps a stack trace into the ws bytes).
	failClient := func(code, message string) mo.Result[any] {
		errEvent, marshalErr := json.Marshal(types.SpeechStreamServerEvent{
			Event:   types.SpeechStreamServerEventError,
			Code:    code,
			Message: message,
		})
		if marshalErr == nil {
			_ = clientWS.WriteMessage(websocket.TextMessage, errEvent)
		}

		_ = clientWS.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, truncateWsCloseReason(message)),
		)

		return mo.Ok[any](nil)
	}

	apiKey := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
	if apiKey == "" {
		return failClient("missing_api_key", "missing X-Api-Key in Authorization header")
	}

	resourceID := utils.GetByJSONPath[string](opts.ExtraBody, "{ .api_resource_id }")
	if resourceID == "" {
		resourceID = "seed-tts-2.0"
	}

	connectID := utils.GetByJSONPath[string](opts.ExtraBody, "{ .api_connect_id }")
	if connectID == "" {
		connectID = uuid.New().String()
	}

	upstreamHeaders := http.Header{}
	upstreamHeaders.Set("X-Api-Key", apiKey)
	upstreamHeaders.Set("X-Api-Resource-Id", resourceID)
	upstreamHeaders.Set("X-Api-Connect-Id", connectID)
	// Request usage in SessionFinished so downstream billing proxies can meter
	// the session by text_words. Documented at
	// https://www.volcengine.com/docs/6561/1329505 ("X-Control-Require-Usage-Tokens-Return").
	upstreamHeaders.Set("X-Control-Require-Usage-Tokens-Return", "*")

	upstream, resp, err := websocket.DefaultDialer.DialContext(c.Request().Context(), v3BidirectionEndpoint, upstreamHeaders)
	if err != nil {
		if resp == nil {
			return failClient("upstream_dial_failed", err.Error())
		}

		defer func() { _ = resp.Body.Close() }()

		errResult := utils.NewJSONResponseError(resp.StatusCode, resp.Body)
		jsonErr, parseErr := errResult.Get()

		detail := err.Error()
		if parseErr == nil {
			detail = jsonErr.Error()
		}

		return failClient(fmt.Sprintf("upstream_handshake_%d", resp.StatusCode), detail)
	}

	defer func() {
		_ = resp.Body.Close()
		_ = upstream.Close()
	}()

	sessionID := uuid.New().String()
	bridge := &v3Bridge{
		client:    clientWS,
		upstream:  upstream,
		sessionID: sessionID,
		startOpts: opts,
		startedCh: make(chan struct{}, 1),
		doneCh:    make(chan struct{}),
	}

	bridgeErr := bridge.start(c.Request().Context())
	if bridgeErr != nil {
		bridge.sendClientError("upstream_error", bridgeErr.Error())
		_ = clientWS.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, truncateWsCloseReason(bridgeErr.Error())),
		)

		return mo.Ok[any](nil)
	}

	return mo.Ok[any](nil)
}

// truncateWsCloseReason caps a WebSocket close reason to the 123-byte protocol
// limit (RFC 6455 §5.5: control frame payload max is 125 bytes; close frame
// reserves 2 bytes for the status code).
func truncateWsCloseReason(s string) string {
	const maxReasonBytes = 123
	if len(s) <= maxReasonBytes {
		return s
	}

	return s[:maxReasonBytes]
}

type v3Bridge struct {
	client    *websocket.Conn
	upstream  *websocket.Conn
	sessionID string
	startOpts types.SpeechRequestOptions
	startedCh chan struct{}
	// doneCh is closed when `start` returns, signalling the client loop that
	// the bridge is tearing down so it can exit the post-`finish` drain state.
	doneCh chan struct{}

	writeMu sync.Mutex
}

func (b *v3Bridge) start(ctx context.Context) error {
	// Signal client loop to exit its post-`finish` drain state when this
	// function returns (either because the upstream session finished or
	// because we hit an error tear-down). Without this, a client that sent
	// `finish` would block forever in the drain select even after we close
	// the bridge.
	defer close(b.doneCh)

	err := b.sendUpstream(v3Frame{
		msgType: v3MsgTypeFullClient,
		flags:   v3FlagWithEvent,
		serial:  v3SerialJSON,
		event:   v3EventStartConnection,
		payload: []byte(`{}`),
	})
	if err != nil {
		return err
	}

	upstreamErrCh := make(chan error, 1)
	clientErrCh := make(chan error, 1)

	go func() { upstreamErrCh <- b.readUpstreamLoop(ctx) }()
	go func() { clientErrCh <- b.readClientLoop(ctx) }()

	// Whichever loop exits first ends the bridge:
	//
	// - Upstream loop exits naturally when `SessionFinished` /
	//   `SessionFailed` / `ConnectionFailed` arrives, which is the **normal**
	//   completion path. The client loop is in its post-`finish` drain
	//   waiting for this exact moment (see readClientLoop).
	// - Client loop exits on `cancel` (forced tear-down) or on a non-graceful
	//   client ws error (network drop). Both are abort paths.
	select {
	case err := <-upstreamErrCh:
		return err
	case err := <-clientErrCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *v3Bridge) sendUpstream(f v3Frame) error {
	b.writeMu.Lock()
	defer b.writeMu.Unlock()

	return b.upstream.WriteMessage(websocket.BinaryMessage, encodeV3Frame(f))
}

func (b *v3Bridge) sendClientJSON(event types.SpeechStreamServerEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		slog.Warn("v3 bridge: failed to marshal client event", slog.String("err", err.Error()))

		return
	}

	err = b.client.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		slog.Warn("v3 bridge: failed to write client event", slog.String("err", err.Error()))
	}
}

func (b *v3Bridge) sendClientError(code, message string) {
	b.sendClientJSON(types.SpeechStreamServerEvent{
		Event:   types.SpeechStreamServerEventError,
		Code:    code,
		Message: message,
	})
}

func (b *v3Bridge) sendClientAudio(audio []byte) {
	err := b.client.WriteMessage(websocket.BinaryMessage, audio)
	if err != nil {
		slog.Warn("v3 bridge: failed to write client audio", slog.String("err", err.Error()))
	}
}

func (b *v3Bridge) buildStartSessionPayload() ([]byte, error) {
	opts := b.startOpts

	audio := v3ReqParamsAudio{
		Format:     lo.Ternary(opts.ResponseFormat != "", opts.ResponseFormat, "mp3"),
		SampleRate: lo.FromPtr(utils.GetByJSONPath[*int](opts.ExtraBody, "{ .audio.sample_rate }")),
		BitRate:    lo.FromPtr(utils.GetByJSONPath[*int](opts.ExtraBody, "{ .audio.bit_rate }")),
		Emotion:    utils.GetByJSONPath[string](opts.ExtraBody, "{ .audio.emotion }"),
	}

	if v := utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .audio.emotion_scale }"); v != nil {
		audio.EmotionScale = *v
	}

	if v := utils.GetByJSONPath[*int](opts.ExtraBody, "{ .audio.speech_rate }"); v != nil {
		audio.SpeechRate = *v
	}

	if v := utils.GetByJSONPath[*int](opts.ExtraBody, "{ .audio.loudness_rate }"); v != nil {
		audio.LoudnessRate = *v
	}

	if v := utils.GetByJSONPath[*bool](opts.ExtraBody, "{ .audio.enable_timestamp }"); v != nil {
		audio.EnableTimestamp = *v
	}

	if v := utils.GetByJSONPath[*bool](opts.ExtraBody, "{ .audio.enable_subtitle }"); v != nil {
		audio.EnableSubtitle = *v
	}

	additionsMap := utils.GetByJSONPath[map[string]any](opts.ExtraBody, "{ .additions }")

	additionsRaw, err := json.Marshal(additionsMap)
	if err != nil {
		return nil, err
	}

	params := v3ReqParams{
		Speaker:     opts.Voice,
		Model:       utils.GetByJSONPath[string](opts.ExtraBody, "{ .model_variant }"),
		AudioParams: &audio,
	}

	if string(additionsRaw) != "null" {
		params.Additions = string(additionsRaw)
	}

	if v := utils.GetByJSONPath[string](opts.ExtraBody, "{ .section_id }"); v != "" {
		params.SectionID = v
	}

	if v := utils.GetByJSONPath[[]any](opts.ExtraBody, "{ .context_texts }"); len(v) > 0 {
		params.ContextTexts = lo.FilterMap(v, func(item any, _ int) (string, bool) {
			s, ok := item.(string)

			return s, ok
		})
	}

	payload := v3SessionPayload{
		Event:     v3EventStartSession,
		Namespace: "BidirectionalTTS",
		ReqParams: params,
	}

	if uid := utils.GetByJSONPath[string](opts.ExtraBody, "{ .user.uid }"); uid != "" {
		payload.User = map[string]any{"uid": uid}
	}

	return json.Marshal(payload)
}

func (b *v3Bridge) buildTaskRequestPayload(text string) ([]byte, error) {
	payload := v3SessionPayload{
		Event:     v3EventTaskRequest,
		Namespace: "BidirectionalTTS",
		ReqParams: v3ReqParams{
			Text:    text,
			Speaker: b.startOpts.Voice,
		},
	}

	return json.Marshal(payload)
}

func (b *v3Bridge) readUpstreamLoop(_ context.Context) error {
	for {
		msgType, data, err := b.upstream.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return nil
			}

			return apierrors.NewErrBadGateway().WithDetail(err.Error())
		}

		if msgType != websocket.BinaryMessage {
			continue
		}

		f, decodeErr := decodeV3Frame(data)
		if decodeErr != nil {
			slog.Warn("v3 bridge: decode upstream frame", slog.String("err", decodeErr.Error()))

			continue
		}

		dispatchErr := b.dispatchUpstream(f)
		if dispatchErr != nil {
			return dispatchErr
		}

		if f.event == v3EventSessionFinished || f.event == v3EventSessionFailed || f.event == v3EventConnectionFailed {
			return nil
		}
	}
}

func (b *v3Bridge) dispatchUpstream(f v3Frame) error {
	if f.msgType == v3MsgTypeError {
		b.sendClientError("upstream_error", string(f.payload))

		return apierrors.NewErrBadGateway().WithDetailf("upstream error code=%d payload=%s", f.errorCode, string(f.payload))
	}

	switch f.event { //nolint:exhaustive
	case v3EventConnectionStarted:
		payload, err := b.buildStartSessionPayload()
		if err != nil {
			return err
		}

		return b.sendUpstream(v3Frame{
			msgType:   v3MsgTypeFullClient,
			flags:     v3FlagWithEvent,
			serial:    v3SerialJSON,
			event:     v3EventStartSession,
			sessionID: b.sessionID,
			payload:   payload,
		})

	case v3EventConnectionFailed:
		b.sendClientError("connection_failed", string(f.payload))

		return apierrors.NewErrBadGateway().WithDetail(string(f.payload))

	case v3EventSessionStarted:
		b.sendClientJSON(types.SpeechStreamServerEvent{Event: types.SpeechStreamServerEventSessionStarted})

		select {
		case b.startedCh <- struct{}{}:
		default:
		}

	case v3EventTTSResponse:
		b.sendClientAudio(f.payload)

	case v3EventTTSSentenceStart:
		b.sendClientJSON(types.SpeechStreamServerEvent{
			Event:   types.SpeechStreamServerEventSentenceStart,
			Payload: decodeJSONObject(f.payload),
		})

	case v3EventTTSSentenceEnd:
		b.sendClientJSON(types.SpeechStreamServerEvent{
			Event:   types.SpeechStreamServerEventSentenceEnd,
			Payload: decodeJSONObject(f.payload),
		})

	case v3EventSessionFinished:
		b.sendClientJSON(types.SpeechStreamServerEvent{
			Event:   types.SpeechStreamServerEventSessionFinished,
			Payload: decodeJSONObject(f.payload),
		})

		_ = b.sendUpstream(v3Frame{
			msgType: v3MsgTypeFullClient,
			flags:   v3FlagWithEvent,
			serial:  v3SerialJSON,
			event:   v3EventFinishConnection,
			payload: []byte(`{}`),
		})

	case v3EventSessionFailed:
		b.sendClientError("session_failed", string(f.payload))

		return apierrors.NewErrBadGateway().WithDetail(string(f.payload))
	}

	return nil
}

func (b *v3Bridge) readClientLoop(ctx context.Context) error {
	// Wait for upstream session before forwarding text — sending TaskRequest
	// before SessionStarted yields an upstream error.
	select {
	case <-b.startedCh:
	case <-ctx.Done():
		return ctx.Err()
	}

	// Tracks whether the client has signalled `finish`. After finish, we MUST
	// NOT return — the upstream loop owns normal completion (it drains
	// remaining audio and emits `SessionFinished`). If we returned here, the
	// outer `start` select would close the upstream ws before that audio
	// arrives, truncating the session (codex review: CRITICAL #1).
	clientFinished := false

	for {
		msgType, data, err := b.client.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				if clientFinished {
					// Client closed after finish — normal teardown. Let the
					// upstream loop drive completion; block until it does.
					<-b.doneCh

					return nil
				}

				return b.finishSession()
			}

			if clientFinished {
				// Underlying ws read errored after the client said finish.
				// This is expected when the bridge tears down (upstream loop
				// closed the connections via the outer defer). Treat as
				// normal completion driven by the upstream loop.
				<-b.doneCh

				return nil
			}

			return apierrors.NewErrBadRequest().WithDetail(err.Error())
		}

		if msgType != websocket.TextMessage {
			continue
		}

		var event types.SpeechStreamClientEvent

		unmarshalErr := json.Unmarshal(data, &event)
		if unmarshalErr != nil {
			b.sendClientError("invalid_event", unmarshalErr.Error())

			continue
		}

		// Post-finish drain: only `cancel` is still meaningful. Users abort
		// mid-synthesis by routing a `cancel` after `finish` (Stage.vue's
		// pipeline sends finish immediately after text, so AbortController
		// firing later can only land here). Swallowing it would leave the
		// upstream session running until SessionFinished, defeating abort.
		// Codex review verified call #2.
		if clientFinished {
			if event.Event == types.SpeechStreamClientEventCancel {
				_ = b.sendUpstream(v3Frame{
					msgType:   v3MsgTypeFullClient,
					flags:     v3FlagWithEvent,
					serial:    v3SerialJSON,
					event:     v3EventCancelSession,
					sessionID: b.sessionID,
					payload:   []byte(`{}`),
				})

				return nil
			}

			continue
		}

		switch event.Event { //nolint:exhaustive
		case types.SpeechStreamClientEventText:
			if event.Text == "" {
				continue
			}

			payload, buildErr := b.buildTaskRequestPayload(event.Text)
			if buildErr != nil {
				return buildErr
			}

			sendErr := b.sendUpstream(v3Frame{
				msgType:   v3MsgTypeFullClient,
				flags:     v3FlagWithEvent,
				serial:    v3SerialJSON,
				event:     v3EventTaskRequest,
				sessionID: b.sessionID,
				payload:   payload,
			})
			if sendErr != nil {
				return sendErr
			}

		case types.SpeechStreamClientEventFinish:
			if finishErr := b.finishSession(); finishErr != nil {
				return finishErr
			}

			clientFinished = true

		case types.SpeechStreamClientEventCancel:
			// Cancel is an explicit abort — do NOT enter drain state. Send
			// CancelSession and return so the outer select tears the bridge
			// down immediately.
			_ = b.sendUpstream(v3Frame{
				msgType:   v3MsgTypeFullClient,
				flags:     v3FlagWithEvent,
				serial:    v3SerialJSON,
				event:     v3EventCancelSession,
				sessionID: b.sessionID,
				payload:   []byte(`{}`),
			})

			return nil

		default:
			b.sendClientError("unknown_event", string(event.Event))
		}
	}
}

func (b *v3Bridge) finishSession() error {
	return b.sendUpstream(v3Frame{
		msgType:   v3MsgTypeFullClient,
		flags:     v3FlagWithEvent,
		serial:    v3SerialJSON,
		event:     v3EventFinishSession,
		sessionID: b.sessionID,
		payload:   []byte(`{}`),
	})
}

func decodeJSONObject(data []byte) map[string]any {
	if len(data) == 0 {
		return nil
	}

	var out map[string]any

	err := json.Unmarshal(data, &out)
	if err != nil {
		return nil
	}

	return out
}
