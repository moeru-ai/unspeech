package deepgram

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/utils"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

func HandleSpeech(c echo.Context, options mo.Option[types.SpeechRequestOptions]) mo.Result[any] {
	opt := options.MustGet()

	// Deepgram uses query parameters for model/voice configuration
	// https://developers.deepgram.com/docs/text-to-speech
	u, _ := url.Parse("https://api.deepgram.com/v1/speak")
	q := u.Query()

	if opt.Voice != "" {
		q.Set("model", opt.Voice)
	}

	u.RawQuery = q.Encode()

	// Request body only needs text
	payload := lo.Must(json.Marshal(map[string]string{
		"text": opt.Input,
	}))

	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		u.String(),
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithCaller())
	}

	auth := c.Request().Header.Get("Authorization")
	if after, ok := strings.CutPrefix(auth, "Bearer "); ok {
		auth = "Token " + after
	}

	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/*")

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

	if res.StatusCode >= http.StatusBadRequest {
		return mo.Err[any](
			apierrors.NewUpstreamError(res.StatusCode).
				WithDetail(utils.NewJSONResponseError(res.StatusCode, res.Body).OrEmpty().Error()),
		)
	}

	// Stream audio response
	// Deepgram returns the audio content directly
	contentType := res.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "audio/mpeg" // Default to mp3 if not specified
	}

	return mo.Ok[any](c.Stream(http.StatusOK, contentType, res.Body))
}
