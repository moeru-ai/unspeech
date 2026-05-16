package types

import (
	"encoding/json"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/mo"

	"github.com/moeru-ai/unspeech/pkg/apierrors"
)

// SpeechStreamClientEventType is the discriminator for events the client
// sends to unspeech on a streaming TTS WebSocket (JSON text frames).
type SpeechStreamClientEventType string

const (
	SpeechStreamClientEventStart  SpeechStreamClientEventType = "start"
	SpeechStreamClientEventText   SpeechStreamClientEventType = "text"
	SpeechStreamClientEventFinish SpeechStreamClientEventType = "finish"
	SpeechStreamClientEventCancel SpeechStreamClientEventType = "cancel"
)

type SpeechStreamClientEvent struct {
	Event SpeechStreamClientEventType `json:"event"`
	Text  string                      `json:"text,omitempty"`
}

// SpeechStreamServerEventType is the discriminator for events unspeech sends
// back to the client (JSON text frames). Audio bytes travel as binary frames
// and do not use this envelope.
type SpeechStreamServerEventType string

const (
	SpeechStreamServerEventSessionStarted  SpeechStreamServerEventType = "session.started"
	SpeechStreamServerEventSentenceStart   SpeechStreamServerEventType = "sentence.start"
	SpeechStreamServerEventSentenceEnd     SpeechStreamServerEventType = "sentence.end"
	SpeechStreamServerEventSubtitle        SpeechStreamServerEventType = "subtitle"
	SpeechStreamServerEventSessionFinished SpeechStreamServerEventType = "session.finished"
	SpeechStreamServerEventError           SpeechStreamServerEventType = "error"
)

type SpeechStreamServerEvent struct {
	Event   SpeechStreamServerEventType `json:"event"`
	Text    string                      `json:"text,omitempty"`
	Code    string                      `json:"code,omitempty"`
	Message string                      `json:"message,omitempty"`
	Payload map[string]any              `json:"payload,omitempty"`
}

// NewSpeechStreamStartOptions parses the WS `start` frame. Unlike
// NewSpeechRequestOptions, `input` is optional because text flows through later
// `text` frames.
func NewSpeechStreamStartOptions(payload []byte) mo.Result[SpeechRequestOptions] {
	var optionsMap map[string]any

	err := json.Unmarshal(payload, &optionsMap)
	if err != nil {
		return mo.Err[SpeechRequestOptions](apierrors.NewErrBadRequest().WithDetail(err.Error()))
	}

	var options OpenAISpeechRequestOptions

	err = json.Unmarshal(payload, &options)
	if err != nil {
		return mo.Err[SpeechRequestOptions](apierrors.NewErrBadRequest().WithDetail(err.Error()))
	}

	if options.Model == "" || options.Voice == "" {
		return mo.Err[SpeechRequestOptions](apierrors.NewErrInvalidArgument().WithDetail("model and voice parameter are required"))
	}

	backendAndModel := lo.Ternary(
		strings.Contains(options.Model, "/"),
		strings.SplitN(options.Model, "/", 2), //nolint:mnd
		[]string{options.Model, ""},
	)

	return mo.Ok(SpeechRequestOptions{
		OpenAISpeechRequestOptions: options,
		Backend:                    backendAndModel[0],
		Model:                      backendAndModel[1],
		bodyParsedMap:              optionsMap,
	})
}
