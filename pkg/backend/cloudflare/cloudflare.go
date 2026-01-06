package cloudflareai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/utils"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// CloudflareSpeechRequest defines the payload for the Workers AI TTS API.
// Based on official documentation for models like @cf/myshell-ai/melotts.
type CloudflareSpeechRequest struct {
	Text string `json:"text"`
	// 'lang' is another potential field for models that support it.
	// Lang string `json:"lang,omitempty"`
}

// HandleSpeechCloudflare processes a TTS request using the Cloudflare Workers AI API.
// It requires the Cloudflare Account ID to be passed in.
func HandleSpeechCloudflare(c echo.Context, accountID string, options mo.Option[types.SpeechRequestOptions]) mo.Result[any] {
	// Extract options safely once
	opt := options.MustGet()

	// --- 1. Select Model ---
	// Choose a Cloudflare TTS model.
	// You could make this dynamic based on opt.Model if you map your internal
	// model names (e.g., "tts-1") to Cloudflare's model names.
	//
	// Available models include:
	// - @cf/myshell-ai/melotts (supports 'text' and 'lang' params)
	// - @cf/deepgram/aura-1 (supports 'text' param)
	//
	// We'll use @cf/myshell-ai/melotts as an example.
	const modelName = "@cf/myshell-ai/melotts"

	// --- 2. Build Cloudflare Payload ---
	// Note: Cloudflare's TTS models (like melotts) do not support the
	// 'voice', 'speed', or 'response_format' parameters from the OpenAI API.
	// The input text field is 'text' (or 'prompt' for some models), not 'input'.
	values := CloudflareSpeechRequest{
		Text: opt.Input,
	}
	payload := lo.Must(json.Marshal(values))

	// --- 3. Build HTTP Request ---
	// The endpoint format is:
	// https://api.cloudflare.com/client/v4/accounts/{ACCOUNT_ID}/ai/run/{MODEL_NAME}
	endpoint := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/accounts/%s/ai/run/%s",
		accountID,
		modelName,
	)

	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		endpoint,
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithCaller())
	}

	// Set Headers
	// The Authorization header must contain a Cloudflare API Token (Bearer)
	req.Header.Set("Authorization", c.Request().Header.Get("Authorization"))
	req.Header.Set("Content-Type", "application/json")
	// Requesting a specific audio format is good practice, though models
	// often default to mp3.
	req.Header.Set("Accept", "audio/mpeg")

	// --- 4. Execute Request ---
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return mo.Err[any](
			apierrors.NewErrBadGateway().
				WithDetail(err.Error()).
				WithError(err).
				WithCaller(),
		)
	}

	defer func() { _ = res.Body.Close() }()

	// --- 5. Handle Errors (Same as before, this logic is solid) ---
	if res.StatusCode >= http.StatusBadRequest {
		ct := res.Header.Get("Content-Type")

		switch {
		case strings.HasPrefix(ct, "application/json"):
			// Cloudflare errors are returned as JSON
			return mo.Err[any](
				apierrors.NewUpstreamError(res.StatusCode).
					WithDetail(utils.NewJSONResponseError(res.StatusCode, res.Body).OrEmpty().Error()),
			)
		case strings.HasPrefix(ct, "text/"):
			return mo.Err[any](
				apierrors.NewUpstreamError(res.StatusCode).
					WithDetail(utils.NewTextResponseError(res.StatusCode, res.Body).OrEmpty().Error()),
			)
		default:
			slog.Warn("unknown upstream error",
				slog.Int("status", res.StatusCode),
				slog.String("content_type", ct),
				slog.String("content_length", res.Header.Get("Content-Length")),
			)

			return mo.Err[any](
				apierrors.NewUpstreamError(res.StatusCode).
					WithDetail("unknown Content-Type: " + ct),
			)
		}
	}

	// --- 6. Stream Successful Audio Response ---
	// On success, Cloudflare returns the raw audio stream directly in the body.
	// The Content-Type (e.g., "audio/mpeg") is correctly proxied.
	return mo.Ok[any](c.Stream(http.StatusOK, res.Header.Get("Content-Type"), res.Body))
}
