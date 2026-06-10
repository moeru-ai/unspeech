package backend

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/samber/mo"

	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/alibaba"
	"github.com/moeru-ai/unspeech/pkg/backend/deepgram"
	"github.com/moeru-ai/unspeech/pkg/backend/elevenlabs"
	"github.com/moeru-ai/unspeech/pkg/backend/koemotion"
	"github.com/moeru-ai/unspeech/pkg/backend/microsoft"
	"github.com/moeru-ai/unspeech/pkg/backend/openai"
	"github.com/moeru-ai/unspeech/pkg/backend/stepfun"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/backend/volcengine"
	"github.com/moeru-ai/unspeech/pkg/utils"
)

const (
	backendVolcengine = "volcengine"
	backendVolcano    = "volcano"
)

var speechStreamUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Speech(c echo.Context) mo.Result[any] {
	options := types.NewSpeechRequestOptions(c.Request().Body)
	if options.IsError() {
		return mo.Err[any](options.Error())
	}

	switch options.MustGet().Backend {
	case "openai":
		return openai.HandleSpeech(c, utils.ResultToOption(options))
	case "stepfun", "step":
		return stepfun.HandleSpeech(c, utils.ResultToOption(options))
	case "deepgram":
		return deepgram.HandleSpeech(c, utils.ResultToOption(options))
	case "elevenlabs":
		return elevenlabs.HandleSpeech(c, utils.ResultToOption(options))
	case "koemotion":
		return koemotion.HandleSpeech(c, utils.ResultToOption(options))
	case "microsoft", "azure":
		return microsoft.HandleSpeech(c, utils.ResultToOption(options))
	case backendVolcengine, backendVolcano:
		return volcengine.HandleSpeech(c, utils.ResultToOption(options))
	case "ali", "aliyun", "alibaba", "bailian", "alibaba-model-studio":
		return alibaba.HandleSpeech(c, utils.ResultToOption(options))
	default:
		return mo.Err[any](apierrors.NewErrBadRequest().WithDetail("unsupported backend"))
	}
}

func SpeechStream(c echo.Context) mo.Result[any] {
	ws, err := speechStreamUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		// Upgrade itself writes a response on failure (echo can still write
		// because the conn is not yet hijacked when upgrade fails).
		return mo.Err[any](apierrors.NewErrBadRequest().WithDetail(err.Error()))
	}

	defer func() { _ = ws.Close() }()

	// Post-upgrade, the underlying conn is hijacked: echo's error middleware
	// MUST NOT try to write an HTTP response on it (would dump a stack trace
	// over the websocket bytes and the client sees an abnormal close with no
	// reason). All failures from here on are surfaced via JSON `error` events
	// on the open ws, followed by a clean close. We always return mo.Ok so
	// the echo handler chain treats the request as completed normally.
	failClient := func(code, message string) mo.Result[any] {
		errEvent, marshalErr := json.Marshal(types.SpeechStreamServerEvent{
			Event:   types.SpeechStreamServerEventError,
			Code:    code,
			Message: message,
		})
		if marshalErr == nil {
			_ = ws.WriteMessage(websocket.TextMessage, errEvent)
		}
		// Reason length is limited to 123 bytes by the protocol; truncate to
		// be safe. Use policy-violation code so the browser surfaces it
		// distinctly from network errors.
		closeMsg := websocket.FormatCloseMessage(websocket.ClosePolicyViolation, truncateCloseReason(message))
		_ = ws.WriteMessage(websocket.CloseMessage, closeMsg)

		return mo.Ok[any](nil)
	}

	msgType, payload, err := ws.ReadMessage()
	if err != nil {
		return failClient("read_first_frame_failed", err.Error())
	}

	if msgType != websocket.TextMessage {
		return failClient("invalid_first_frame", "first frame must be a JSON text frame")
	}

	var envelope struct {
		Event string `json:"event"`
	}

	unmarshalErr := json.Unmarshal(payload, &envelope)
	if unmarshalErr != nil {
		return failClient("invalid_first_frame", unmarshalErr.Error())
	}

	if envelope.Event != string(types.SpeechStreamClientEventStart) {
		return failClient("invalid_first_frame", "first frame must be event=start")
	}

	options := types.NewSpeechStreamStartOptions(payload)
	if options.IsError() {
		return failClient("invalid_start_options", options.Error().Error())
	}

	switch options.MustGet().Backend {
	case backendVolcengine, backendVolcano:
		// Backend takes over the ws from here. It also surfaces any further
		// errors via `error` events on the same ws (see HandleSpeechStream).
		result := volcengine.HandleSpeechStream(c, ws, utils.ResultToOption(options))
		if result.IsError() {
			// Backend already sent its own error event; just absorb the error
			// so the echo chain doesn't try to write to the hijacked conn.
			return mo.Ok[any](nil)
		}

		return result
	default:
		return failClient("unsupported_backend", "streaming is only supported for backend=volcengine")
	}
}

// truncateCloseReason caps a WebSocket close reason to the 123-byte protocol
// limit (RFC 6455 §5.5: control frame payload max is 125 bytes; close frame
// reserves 2 bytes for the status code).
func truncateCloseReason(s string) string {
	const maxReasonBytes = 123
	if len(s) <= maxReasonBytes {
		return s
	}

	return s[:maxReasonBytes]
}

func Voices(c echo.Context) mo.Result[any] {
	options := types.NewVoicesRequestOptions(c.Request())
	if options.IsError() {
		return mo.Err[any](options.Error())
	}

	switch options.MustGet().Backend {
	case "openai":
		return openai.HandleVoices(c, utils.ResultToOption(options))
	case "stepfun", "step":
		return stepfun.HandleVoices(c, utils.ResultToOption(options))
	case "deepgram":
		return deepgram.HandleVoices(c, utils.ResultToOption(options))
	case "elevenlabs":
		return elevenlabs.HandleVoices(c, utils.ResultToOption(options))
	case "koemotion":
		return koemotion.HandleVoices(c, utils.ResultToOption(options))
	case "microsoft", "azure":
		return microsoft.HandleVoices(c, utils.ResultToOption(options))
	case backendVolcengine, backendVolcano:
		return volcengine.HandleVoices(c, utils.ResultToOption(options))
	case "ali", "aliyun", "alibaba", "bailian", "alibaba-model-studio":
		return alibaba.HandleVoices(c, utils.ResultToOption(options))
	default:
		return mo.Err[any](apierrors.NewErrBadRequest().WithDetail("unsupported backend"))
	}
}
