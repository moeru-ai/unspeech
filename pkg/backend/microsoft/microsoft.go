package microsoft

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

type voice struct {
	XMLName xml.Name `xml:"voice"`
	Lang    string   `xml:"lang,attr"`
	Gender  string   `xml:"gender,attr"`
	Name    string   `xml:"name,attr"`
	Text    string   `xml:",chardata"`
}

// SSML structures for parsing
type ssml struct {
	XMLName  xml.Name `xml:"speak"`
	Version  string   `xml:"version,attr"`
	Lang     string   `xml:"lang,attr"`
	Voice    voice    `xml:"voice"`
	TextOnly string   `xml:",chardata"`
}

type extraBody struct {
	DisableSSML  mo.Option[bool]   `json:"disable_ssml,omitempty"`
	Region       string            `json:"region"`
	DeploymentID mo.Option[string] `json:"deployment_id,omitempty"`
	Lang         mo.Option[string] `json:"lang,omitempty"`
	Gender       mo.Option[string] `json:"gender,omitempty"`
	SampleRate   mo.Option[uint]   `json:"sample_rate,omitempty"`
}

var (
	// Text to speech API reference (REST) - Speech service - Azure AI services | Microsoft Learn
	// https://learn.microsoft.com/en-us/azure/ai-services/speech-service/rest-text-to-speech?tabs=streaming#audio-outputs
	supportedOutputFormats = map[string]map[uint][]string{
		"mp3": {
			16000: {
				"audio-16khz-32kbitrate-mono-mp3",
				"audio-16khz-64kbitrate-mono-mp3",
				"audio-16khz-128kbitrate-mono-mp3",
			},
			24000: {
				"audio-24khz-48kbitrate-mono-mp3",
				"audio-24khz-96kbitrate-mono-mp3",
				"audio-24khz-160kbitrate-mono-mp3",
			},
			48000: {
				"audio-48khz-96kbitrate-mono-mp3",
				"audio-48khz-192kbitrate-mono-mp3",
			},
		},
		"opus": {
			16000: {
				"audio-16khz-16bit-32kbps-mono-opus",
				"ogg-16khz-16bit-mono-opus",
				"webm-16khz-16bit-mono-opus",
			},
			24000: {
				"audio-24khz-16bit-24kbps-mono-opus",
				"audio-24khz-16bit-48kbps-mono-opus",
				"ogg-24khz-16bit-mono-opus",
				"webm-24khz-16bit-24kbps-mono-opus",
				"webm-24khz-16bit-mono-opus",
			},
			48000: {
				"ogg-48khz-16bit-mono-opus",
			},
		},
		"wav": {
			8000: {
				"raw-8khz-16bit-mono-pcm",
				"raw-8khz-8bit-mono-alaw",
				"raw-8khz-8bit-mono-mulaw",
			},
			16000: {
				"raw-16khz-16bit-mono-pcm",
				"raw-16khz-16bit-mono-truesilk",
			},
			22050: {
				"raw-22050hz-16bit-mono-pcm",
			},
			24000: {
				"raw-24khz-16bit-mono-pcm",
				"raw-24khz-16bit-mono-truesilk",
			},
			44100: {
				"raw-44100hz-16bit-mono-pcm",
			},
			48000: {
				"raw-48khz-16bit-mono-pcm",
			},
		},
	}
)

func getOutputFormat(format string, sampleRate uint) mo.Option[string] {
	formatsWithSampleRate, ok := supportedOutputFormats[format]
	if !ok {
		return mo.None[string]()
	}

	formatFull, ok := formatsWithSampleRate[sampleRate]
	if !ok {
		return mo.None[string]()
	}

	return mo.Some(formatFull[0])
}

// Helper function to process input text and SSML
func processSSML(input string, option types.SpeechRequestOptions, extraBody mo.Option[extraBody]) string {
	// Check if SSML formatting is disabled
	if extraBody.OrEmpty().DisableSSML.OrEmpty() {
		return input
	}

	// Default SSML values
	defaultLang := "en-US"
	defaultGender := "Male"
	defaultVoiceName := lo.CoalesceOrEmpty(option.Voice, "en-US-ChristopherNeural")

	// Override defaults with values from ExtraBody if provided
	if extraBody.OrEmpty().Lang.IsPresent() {
		defaultLang = extraBody.MustGet().Lang.MustGet()
	}
	if extraBody.OrEmpty().Gender.IsPresent() {
		defaultGender = extraBody.MustGet().Gender.MustGet()
	}

	// Check if input already contains SSML
	if !strings.Contains(input, "<speak") {
		// For plain text, format with defaults
		return formatAsSSML(input, defaultLang, defaultGender, defaultVoiceName)
	}

	// Try to parse the existing SSML
	var s ssml

	err := xml.Unmarshal([]byte(input), &s)
	if err != nil {
		// If parsing fails, treat as plain text and format with defaults
		slog.Warn("failed to parse SSML, using as plain text", slog.String("error", err.Error()))
		return formatAsSSML(input, defaultLang, defaultGender, defaultVoiceName)
	}

	// Extract values from parsed SSML, using defaults for missing values
	lang := defaultLang
	if s.Lang != "" {
		lang = s.Lang
	}

	gender := defaultGender
	if s.Voice.Gender != "" {
		gender = s.Voice.Gender
	}

	voiceName := defaultVoiceName
	if s.Voice.Name != "" {
		voiceName = s.Voice.Name
	}

	// If there's no voice tag or it's empty, use the text from the speak tag
	text := s.Voice.Text
	if strings.TrimSpace(text) == "" {
		text = strings.TrimSpace(s.TextOnly)
	}

	// If we have valid text, format it with the extracted/default values
	if strings.TrimSpace(text) != "" {
		return formatAsSSML(text, lang, gender, voiceName)
	}

	// If we couldn't extract text, return the original input
	return input
}

// Helper to format text as SSML
//
// See also: Speech Synthesis Markup Language (SSML) overview - Speech service - Azure AI services | Microsoft Learn
// https://learn.microsoft.com/en-us/azure/ai-services/speech-service/speech-synthesis-markup
func formatAsSSML(text string, lang string, gender string, voiceName string) string {
	return fmt.Sprintf(`<speak version='1.0' xml:lang='%s'>
  <voice xml:lang='%s' xml:gender='%s' name='%s'>
    %s
  </voice>
</speak>`, lang, lang, gender, voiceName, text)
}

func Handle(c echo.Context, options mo.Option[types.SpeechRequestOptions]) mo.Result[any] {
	opts := options.MustGet()

	// Text to speech API reference (REST) - Speech service - Azure AI services | Microsoft Learn
	reqURL := lo.Must(url.Parse(fmt.Sprintf(
		"https://%s.tts.speech.microsoft.com/cognitiveservices/v1",
		// Text to speech API reference (REST) - Speech service - Azure AI services | Microsoft Learn
		// https://learn.microsoft.com/en-us/azure/ai-services/speech-service/rest-text-to-speech?tabs=streaming#prebuilt-neural-voices
		//
		// NOTICE: Voices in preview are available in only these three regions: East US, West Europe, and Southeast Asia.
		lo.Must(lo.Coalesce(opts.ExtraBody["region"], "eastasia")))),
	)

	var extra mo.Option[extraBody]

	{
		if opts.ExtraBody != nil {
			extraBodyJSON, err := json.Marshal(opts.ExtraBody)
			if err != nil {
				return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
			}

			var body extraBody

			err = json.Unmarshal(extraBodyJSON, &body)
			if err != nil {
				return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
			}

			extra = mo.Some(body)
		} else {
			extra = mo.None[extraBody]()
		}
	}

	reqSearchParams := url.Values{}

	// Text to speech API reference (REST) - Speech service - Azure AI services | Microsoft Learn
	// https://learn.microsoft.com/en-us/azure/ai-services/speech-service/rest-text-to-speech?tabs=streaming#custom-neural-voices
	if extra.IsPresent() && extra.OrEmpty().DeploymentID.IsPresent() {
		reqSearchParams.Add("deploymentId", extra.MustGet().DeploymentID.MustGet())
	}

	// Process the input text with SSML handling
	inputText := opts.Input
	formattedText := processSSML(inputText, opts, extra)

	// Create request with the SSML content directly
	req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodPost, reqURL.String()+lo.Ternary(len(reqSearchParams) > 0, "?"+reqSearchParams.Encode(), ""), bytes.NewBufferString(formattedText))
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithCaller())
	}

	var format string
	if opts.ResponseFormat == "" {
		format = "audio-48khz-192kbitrate-mono-mp3"
	} else {
		format = getOutputFormat(opts.ResponseFormat, extra.MustGet().SampleRate.OrElse(48000)).OrEmpty() //nolint:mnd
		if format == "" {
			return mo.Err[any](apierrors.NewErrBadRequest().WithDetail("unsupported output format, check https://learn.microsoft.com/en-us/azure/ai-services/speech-service/rest-text-to-speech?tabs=streaming#audio-outputs for full list of supported formats"))
		}
	}

	req.Header.Set("Ocp-Apim-Subscription-Key", strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer "))
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("X-Microsoft-OutputFormat", format) //nolint:canonicalheader

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return mo.Err[any](apierrors.NewErrBadGateway().WithDetail(err.Error()).WithError(err).WithCaller())
	}

	defer func() { _ = res.Body.Close() }()

	if res.StatusCode >= 400 && res.StatusCode < 600 {
		return mo.Err[any](apierrors.NewUpstreamError(res.StatusCode).WithDetail(res.Status))
	}

	return mo.Ok[any](c.Stream(http.StatusOK, res.Header.Get("Content-Type"), res.Body))
}
