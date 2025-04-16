package alibaba

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/utils"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

type ServerEventEvent string

const (
	ServerEventEventTaskStarted     ServerEventEvent = "task-started"
	ServerEventEventResultGenerated ServerEventEvent = "result-generated"
	ServerEventEventTaskFinished    ServerEventEvent = "task-finished"
	ServerEventEventTaskFailed      ServerEventEvent = "task-failed"
)

type ClientEventAction string

const (
	ClientEventActionContinueTask ClientEventAction = "continue-task"
	ClientEventActionFinishTask   ClientEventAction = "finish-task"
	ClientEventActionRunTask      ClientEventAction = "run-task"
)

type ServerEventHeader struct {
	TaskID     string           `json:"task_id"`
	Event      ServerEventEvent `json:"event"`
	Attributes map[string]any   `json:"attributes"`

	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

type ClientEventHeaderStreaming string

const (
	ClientEventHeaderStreamingDuplex ClientEventHeaderStreaming = "duplex"
)

type ClientEventHeader struct {
	TaskID    string                     `json:"task_id"`
	Action    ClientEventAction          `json:"action"`
	Streaming ClientEventHeaderStreaming `json:"streaming"`
}

type Event struct {
	Header  ServerEventHeader `json:"header"`
	Payload json.RawMessage   `json:"payload"`
}

type ClientEvent[E any] struct {
	Header  ClientEventHeader `json:"header"`
	Payload E                 `json:"payload"`
}

type ClientEventPayloadTaskGroup string

const (
	ClientEventPayloadTaskGroupAudio ClientEventPayloadTaskGroup = "audio"
)

type ClientEventPayloadTask string

const (
	ClientEventPayloadTaskTTS ClientEventPayloadTask = "tts"
)

type ClientEventPayloadFunction string

const (
	ClientEventPayloadFunctionSpeechSynthesizer ClientEventPayloadFunction = "SpeechSynthesizer"
)

type ClientEventRunTaskPayloadParametersTextType string

const (
	ClientEventRunTaskPayloadParametersTextTypePlainText ClientEventRunTaskPayloadParametersTextType = "PlainText"
)

type ClientEventRunTaskPayloadParameters struct {
	TextType   ClientEventRunTaskPayloadParametersTextType `json:"text_type"`
	Voice      string                                      `json:"voice"`
	Format     string                                      `json:"format"`
	SampleRate int                                         `json:"sample_rate"`
	Volume     int                                         `json:"volume"`
	Rate       float64                                     `json:"rate"`
	Pitch      float64                                     `json:"pitch"`
}

type ClientEventRunTaskPayload struct {
	TaskGroup  ClientEventPayloadTaskGroup         `json:"task_group"`
	Task       ClientEventPayloadTask              `json:"task"`
	Function   ClientEventPayloadFunction          `json:"function"`
	Model      string                              `json:"model"`
	Input      map[string]any                      `json:"input"`
	Parameters ClientEventRunTaskPayloadParameters `json:"parameters"`
}

type ClientEventContinueTaskPayloadInput struct {
	Text string `json:"text"`
}

type ClientEventContinueTaskPayload struct {
	TaskGroup ClientEventPayloadTaskGroup         `json:"task_group"`
	Task      ClientEventPayloadTask              `json:"task"`
	Function  ClientEventPayloadFunction          `json:"function"`
	Input     ClientEventContinueTaskPayloadInput `json:"input"`
}

type ClientEventFinishTaskPayload struct {
	Input map[string]any `json:"input"`
}

type EventStructured[P any] struct {
	Header  ServerEventHeader `json:"header"`
	Payload P                 `json:"payload"`
}

func HandleSpeech(c echo.Context, options mo.Option[types.SpeechRequestOptions]) mo.Result[any] {
	taskID := uuid.New().String()
	headers := http.Header{}

	headers.Add("Authorization", strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer "))
	headers.Add("X-DashScope-DataInspection", "enable") //nolint:canonicalheader

	conn, resp, err := websocket.DefaultDialer.Dial("wss://dashscope.aliyuncs.com/api-ws/v1/inference", headers)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	defer func() {
		_ = resp.Body.Close()
		_ = conn.Close()
	}()

	audioBinary := new(bytes.Buffer)
	chanResult := make(chan struct{}, 1)
	chanError := make(chan error, 1)

	go func() {
		defer close(chanResult)
		defer close(chanError)

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				chanError <- apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller()
				return
			}

			switch messageType {
			case websocket.BinaryMessage:
				_, err = audioBinary.Write(message)
				if err != nil {
					chanError <- apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller()
					return
				}
			default:
				var event Event

				err := json.Unmarshal(message, &event)
				if err != nil {
					chanError <- apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller()
					return
				}

				switch event.Header.Event {
				case ServerEventEventTaskStarted:
					err = conn.WriteJSON(ClientEvent[ClientEventContinueTaskPayload]{
						Header: ClientEventHeader{
							TaskID:    taskID,
							Action:    ClientEventActionContinueTask,
							Streaming: ClientEventHeaderStreamingDuplex,
						},
						Payload: ClientEventContinueTaskPayload{
							TaskGroup: ClientEventPayloadTaskGroupAudio,
							Task:      ClientEventPayloadTaskTTS,
							Function:  ClientEventPayloadFunctionSpeechSynthesizer,
							Input: ClientEventContinueTaskPayloadInput{
								Text: options.MustGet().Input,
							},
						},
					})
					if err != nil {
						chanError <- apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller()
						return
					}

					err = conn.WriteJSON(ClientEvent[ClientEventFinishTaskPayload]{
						Header: ClientEventHeader{
							TaskID:    taskID,
							Action:    ClientEventActionFinishTask,
							Streaming: ClientEventHeaderStreamingDuplex,
						},
						Payload: ClientEventFinishTaskPayload{
							Input: make(map[string]any),
						},
					})
					if err != nil {
						chanError <- apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller()
						return
					}
				case ServerEventEventTaskFailed:
					chanError <- apierrors.NewErrBadRequest().WithDetailf("failed to run task, task-failed event received, error_code: %s, error_message: %s", event.Header.ErrorCode, event.Header.ErrorMessage)
				case ServerEventEventResultGenerated:
					// skip as what https://help.aliyun.com/zh/model-studio/cosyvoice-websocket-api has stated that `result-generated` event was reserved for now.
					continue
				case ServerEventEventTaskFinished:
					chanResult <- struct{}{}
				}
			}
		}
	}()

	volume := utils.GetByJSONPath[*int](options.MustGet().ExtraBody, "{ .volume }")
	if volume == nil {
		volume = lo.ToPtr(50) //nolint:mnd
	}

	rate := utils.GetByJSONPath[*float64](options.MustGet().ExtraBody, "{ .rate }")
	if rate == nil {
		rate = lo.ToPtr(float64(1))
	}

	pitch := utils.GetByJSONPath[*float64](options.MustGet().ExtraBody, "{ .pitch }")
	if pitch == nil {
		pitch = lo.ToPtr(float64(1))
	}

	sampleRate := utils.GetByJSONPath[*int](options.MustGet().ExtraBody, "{ .sample_rate }")
	if sampleRate == nil {
		sampleRate = lo.ToPtr(22050) //nolint:mnd
	}

	err = conn.WriteJSON(ClientEvent[ClientEventRunTaskPayload]{
		Header: ClientEventHeader{
			TaskID:    taskID,
			Action:    ClientEventActionRunTask,
			Streaming: ClientEventHeaderStreamingDuplex,
		},
		Payload: ClientEventRunTaskPayload{
			TaskGroup: ClientEventPayloadTaskGroupAudio,
			Task:      ClientEventPayloadTaskTTS,
			Function:  ClientEventPayloadFunctionSpeechSynthesizer,
			Model:     options.MustGet().Model,
			Input:     make(map[string]any),
			Parameters: ClientEventRunTaskPayloadParameters{
				TextType:   ClientEventRunTaskPayloadParametersTextTypePlainText,
				Voice:      options.MustGet().Voice,
				Format:     lo.Ternary(options.MustGet().ResponseFormat == "", "mp3", options.MustGet().ResponseFormat),
				SampleRate: lo.FromPtr(sampleRate),
				Volume:     lo.FromPtr(volume),
				Rate:       lo.FromPtr(rate),
				Pitch:      lo.FromPtr(pitch),
			},
		},
	})
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	slog.Info("task started", "task_id", taskID)

	select {
	case err := <-chanError:
		return mo.Err[any](err)
	case <-chanResult:
		return mo.Ok[any](c.Blob(http.StatusOK, "audio/mp3", audioBinary.Bytes()))
	}
}
