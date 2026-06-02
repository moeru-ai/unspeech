# Bidirectional Streaming TTS Wire Protocol — v1

Endpoint: `GET wss://<host>/v1/audio/speech/stream`

This document defines the WebSocket wire protocol unSpeech exposes for
bidirectional streaming text-to-speech. It is the contract every client
(browsers, Electron apps, server-side proxies) talks against, regardless
of which upstream provider backs the session.

The v1 protocol is currently implemented against the Volcengine
bidirectional TTS upstream
(<https://www.volcengine.com/docs/6561/1329505>). Additional backends
will be added behind the same wire.

## Status

Stable for v1. Additive changes (new optional fields, new server event
types that older clients can ignore) ship under v1. Breaking changes
will be released under a different path (`/v2/audio/speech/stream`).

## Transport

WebSocket over TLS. The endpoint is an HTTP `GET` that the server
upgrades. Both text frames and binary frames are used:

| Frame   | Direction        | Purpose                              |
| ------- | ---------------- | ------------------------------------ |
| Text    | both             | Control events, JSON-encoded         |
| Binary  | server → client  | Raw audio bytes, no framing header   |

Audio is sent as raw WebSocket binary frames rather than base64 inside
JSON to avoid the 33% size overhead of base64 in a continuously
streaming payload.

## Authentication

The HTTP upgrade request must carry the upstream API key in the
`Authorization` header:

```
Authorization: Bearer <upstream_api_key>
```

For the Volcengine backend, this value is forwarded upstream as the
`X-Api-Key` header (new-console auth). Old-console auth
(`X-Api-App-Id` + `X-Api-Access-Key`) is not supported on the streaming
endpoint.

### Browser clients

Browsers cannot set custom headers on `new WebSocket(...)`. Browser
clients must connect through a server-side proxy that injects the
`Authorization` header on their behalf, holding the upstream key
server-side. The proxy can stream its own WebSocket onward to unSpeech
without re-buffering audio.

## Routing

The upstream backend is selected via the `model` field on the first
client frame, using the same `<backend>/<model>` convention as the REST
`POST /v1/audio/speech` endpoint:

| `model` prefix                                 | Backend         |
| ---------------------------------------------- | --------------- |
| `volcengine/<anything>`, `volcano/<anything>`  | Volcengine v3   |

For Volcengine the trailing `<anything>` is informational; the actual
upstream model (resource id) is selected via
`extra_body.api_resource_id` (defaults to `seed-tts-2.0`).

A request whose backend does not implement streaming is rejected with
HTTP 400 during the upgrade handshake.

## Session Lifecycle

```
client                  unspeech                       upstream
  │                        │                              │
  ├── ws upgrade ─────────▶│                              │
  │                        ├── ws upgrade ───────────────▶│
  │                        │                              │
  ├── {event:"start"} ────▶│                              │
  │                        ├── StartConnection ──────────▶│
  │                        │◀── ConnectionStarted ────────┤
  │                        ├── StartSession ─────────────▶│
  │                        │◀── SessionStarted ───────────┤
  │◀── {event:"session.started"} ┤                        │
  │                        │                              │
  ├── {event:"text", text:"Hello "} ─▶                    │
  │                        ├── TaskRequest("Hello ") ────▶│
  │                        │◀── TTSSentenceStart ─────────┤
  │◀── {event:"sentence.start"} ─┤                        │
  │                        │◀── TTSResponse (audio) ──────┤
  │◀── <binary audio>──────┤                              │
  │◀── <binary audio>──────┤                              │
  │                        │◀── TTSSentenceEnd ───────────┤
  │◀── {event:"sentence.end"} ───┤                        │
  │                        │                              │
  ├── {event:"text", text:"world."} ─▶ …                  │
  │                        │                              │
  ├── {event:"finish"} ───▶│                              │
  │                        ├── FinishSession ────────────▶│
  │                        │◀── SessionFinished ──────────┤
  │◀── {event:"session.finished"} ┤                       │
  │                        ├── FinishConnection ─────────▶│
  │                        │◀── close ────────────────────┤
  │◀── close ──────────────┤                              │
```

A single WebSocket connection hosts exactly one session. To synthesize
in parallel, open multiple connections.

The single-session model intentionally mirrors the upstream constraint
(Volcengine v3 forbids concurrent sessions on one upstream connection).
A future v2 may add `context_id` multiplexing similar to Cartesia's
design.

## Client → Server Events

All client-to-server frames are WebSocket **text** frames carrying JSON
with a top-level `event` discriminator.

### `start`

The first frame on every connection. Carries an OpenAI-shape speech
request body with `input` omitted (text arrives later via `text`
frames).

| Field             | Type     | Required | Notes                                                                      |
| ----------------- | -------- | -------- | -------------------------------------------------------------------------- |
| `event`           | string   | yes      | `"start"`                                                                  |
| `model`           | string   | yes      | `<backend>/<id>`, e.g. `volcengine/seed-tts-2.0`                           |
| `voice`           | string   | yes      | Upstream voice/speaker id, e.g. `zh_female_shuangkuaisisi_moon_bigtts`     |
| `response_format` | string   | no       | OpenAI-style: `mp3` (default), `opus`, `aac`, `flac`, `wav`, `pcm`. See below |
| `extra_body`      | object   | no       | Backend-specific knobs (see "Volcengine extra_body" below)                 |

`wav` is rejected by the Volcengine upstream in streaming mode (the
upstream re-emits a WAV header on every chunk). Use `mp3` or `pcm`.

Example:

```json
{
  "event": "start",
  "model": "volcengine/seed-tts-2.0",
  "voice": "zh_female_shuangkuaisisi_moon_bigtts",
  "response_format": "mp3",
  "extra_body": {
    "api_resource_id": "seed-tts-2.0",
    "audio": { "sample_rate": 24000, "bit_rate": 64000, "enable_subtitle": true }
  }
}
```

#### Volcengine `extra_body`

All fields are optional.

| Path                       | Type    | Notes                                                                       |
| -------------------------- | ------- | --------------------------------------------------------------------------- |
| `api_resource_id`          | string  | `seed-tts-2.0` (default), `seed-tts-1.0`, `seed-icl-2.0`, etc.              |
| `api_connect_id`           | string  | Trace id forwarded as `X-Api-Connect-Id`; defaults to a generated UUID      |
| `model_variant`            | string  | TTS 2.0 only: `seed-tts-2.0-standard` or `-expressive`                      |
| `audio.sample_rate`        | int     | 8000 / 16000 / 22050 / 24000 / 32000 / 44100 / 48000 — default 24000        |
| `audio.bit_rate`           | int     | bps; mp3/ogg recommended ≥64000                                             |
| `audio.emotion`            | string  | Upstream emotion tag (see Volcengine voice list)                            |
| `audio.emotion_scale`      | number  | 1–5                                                                         |
| `audio.speech_rate`        | int     | -50 to 100                                                                  |
| `audio.loudness_rate`      | int     | -50 to 100                                                                  |
| `audio.enable_timestamp`   | bool    | Word timestamps on `sentence.end` (TTS 1.0 / ICL 1.0)                       |
| `audio.enable_subtitle`    | bool    | Subtitle events (TTS 2.0 / ICL 2.0)                                         |
| `additions`                | object  | Opaque JSON forwarded as upstream `req_params.additions`                    |
| `section_id`               | string  | Multi-turn session id (TTS 2.0)                                             |
| `context_texts`            | array   | Voice instruction phrases (TTS 2.0 expressive)                              |
| `user.uid`                 | string  | Optional caller id forwarded upstream                                       |

### `text`

Append a chunk of text. Each `text` frame triggers exactly one upstream
`TaskRequest`. Multiple `text` frames within a single session share
voice state — the upstream model treats them as one synthesis pass,
which yields more natural prosody than re-opening a session per chunk.

```json
{ "event": "text", "text": "明朝开国皇帝朱元璋也称这本书为，万物之根" }
```

| Field   | Type   | Required | Notes                                |
| ------- | ------ | -------- | ------------------------------------ |
| `event` | string | yes      | `"text"`                             |
| `text`  | string | yes      | Non-empty; empty values are ignored  |

The Volcengine docs recommend feeding raw LLM streaming output through
this without client-side sentence boundary detection. The upstream
fragments internally to balance latency against naturalness.

### `finish`

End the session cleanly. The server emits any remaining audio, then a
final `session.finished` event, then closes the connection.

```json
{ "event": "finish" }
```

### `cancel`

Abort the in-flight session. The server forwards `CancelSession`
upstream and closes the connection immediately. **No
`session.finished` will follow.** Use this when the user interrupts
mid-stream.

```json
{ "event": "cancel" }
```

> Known limitation: v1 does not wait for the upstream `SessionCanceled`
> ack before closing. Clients learn that cancellation succeeded only
> via the absence of further frames and the WebSocket close. This is
> acceptable because no useful client action depends on the ack — a
> future v2 may surface `session.canceled` if a use case appears.

## Server → Client Events

All server-to-client control frames are WebSocket **text** frames with
the same `event` discriminator. Audio data uses **binary** frames with
no envelope.

### `session.started` (text)

Sent once, after the upstream session is established. After this event
the client may begin sending `text` frames.

```json
{ "event": "session.started" }
```

### `sentence.start` (text)

Marks the beginning of a sentence in the upstream output. The
`payload` field carries the upstream event payload verbatim.

```jsonc
{ "event": "sentence.start", "payload": { /* upstream payload */ } }
```

### `sentence.end` (text)

Marks the end of a sentence. When
`extra_body.audio.enable_timestamp = true` (TTS 1.0 / ICL 1.0), the
payload includes word-level timestamps aligned to the synthesized audio
(`startTime` / `endTime` in seconds, relative to the session start).

```json
{
  "event": "sentence.end",
  "payload": {
    "text": "明朝开国皇帝朱元璋",
    "words": [
      { "word": "明", "startTime": 0.155, "endTime": 0.295 },
      { "word": "朝", "startTime": 0.295, "endTime": 0.425 }
    ]
  }
}
```

### `subtitle` (text)

TTS 2.0 / ICL 2.0 only. Original-text-aligned timestamps that arrive
asynchronously from the audio. May lag the audio for the same
sentence; clients that drive captions should buffer audio playback to
align with subtitle arrival, or accept that captions are best-effort.

```jsonc
{ "event": "subtitle", "payload": { /* upstream payload */ } }
```

### `session.finished` (text)

The final event of a normal session. The payload mirrors the upstream
`SessionFinished` body and may include usage when requested upstream.

```json
{
  "event": "session.finished",
  "payload": {
    "status_code": 20000000,
    "message": "ok",
    "usage": { "text_words": 195 }
  }
}
```

> The bridge sets `X-Control-Require-Usage-Tokens-Return: *` on the upstream
> handshake, so Volcengine returns the character count alongside the session
> close. Providers other than Volcengine may not surface `usage` — clients
> should treat the field as optional.

### `error` (text)

```json
{ "event": "error", "code": "upstream_error", "message": "…" }
```

| `code`              | Trigger                                              |
| ------------------- | ---------------------------------------------------- |
| `upstream_error`    | Upstream sent an error frame (binary `msgType=1111`) |
| `connection_failed` | Upstream `ConnectionFailed` event                    |
| `session_failed`    | Upstream `SessionFailed` event                       |
| `invalid_event`     | Client sent malformed JSON                           |
| `unknown_event`     | Client used an unrecognized `event` value            |

After a fatal `error` (`upstream_error`, `connection_failed`,
`session_failed`) the server closes the connection. `invalid_event`
and `unknown_event` are recoverable — the connection stays open.

### Binary audio frames (binary)

Each binary WebSocket frame is a chunk of raw audio in the format
selected by `response_format` (default `mp3`). There is **no framing
header**: concatenating every binary frame of a session in arrival
order yields the complete audio.

For `pcm`, the byte stream is signed 16-bit little-endian samples at
the requested sample rate. For `mp3` and `ogg_opus`, each frame is a
self-contained chunk that decoders can be fed incrementally (e.g. via
MediaSource Extensions or `mpg123`-style chunked decoding).

## Error Handling

* The WebSocket handshake itself only fails for transport-level reasons
  (TLS, route not mounted). Application-level errors — missing
  `Authorization`, unsupported `model` backend, malformed `start` frame,
  upstream connect failure — surface **after** upgrade: the server emits
  an `error` JSON frame on the open WebSocket and then closes. Clients
  should treat receipt of `error` followed by close as the error signal,
  not the HTTP handshake.
* Once the WebSocket is open, all errors flow through `error` events.
* The client should always handle the WebSocket close event as the
  final ground truth — `session.finished` and `error` are advisory.
* A close arriving without a preceding `session.finished` means the
  session was truncated. Clients must NOT treat partially accumulated
  audio as a success.

## Sample Client (TypeScript, browser)

For browser environments, replace `Authorization` with a proxy that
injects it server-side.

```ts
const ws = new WebSocket('wss://unspeech.example.com/v1/audio/speech/stream')
ws.binaryType = 'arraybuffer'

const audioChunks: Uint8Array[] = []

ws.addEventListener('open', () => {
  ws.send(JSON.stringify({
    event: 'start',
    model: 'volcengine/seed-tts-2.0',
    voice: 'zh_female_shuangkuaisisi_moon_bigtts',
    response_format: 'mp3',
    extra_body: {
      api_resource_id: 'seed-tts-2.0',
      audio: { sample_rate: 24000, bit_rate: 64000 },
    },
  }))
})

ws.addEventListener('message', (e) => {
  if (typeof e.data === 'string') {
    const event = JSON.parse(e.data)
    switch (event.event) {
      case 'session.started':
        ws.send(JSON.stringify({ event: 'text', text: 'Hello, world.' }))
        ws.send(JSON.stringify({ event: 'finish' }))
        break
      case 'sentence.end':
        // event.payload.words has timestamps
        break
      case 'session.finished':
        ws.close()
        break
      case 'error':
        console.error('tts error', event.code, event.message)
        break
    }
  }
  else {
    audioChunks.push(new Uint8Array(e.data as ArrayBuffer))
  }
})
```

## Sample Client (Go, server-side proxy)

```go
header := http.Header{}
header.Set("Authorization", "Bearer "+upstreamKey)

ws, _, err := websocket.DefaultDialer.Dial(
    "wss://unspeech.example.com/v1/audio/speech/stream", header)
if err != nil {
    return err
}
defer ws.Close()

start, _ := json.Marshal(map[string]any{
    "event":           "start",
    "model":           "volcengine/seed-tts-2.0",
    "voice":           voiceID,
    "response_format": "mp3",
    "extra_body": map[string]any{
        "api_resource_id": "seed-tts-2.0",
    },
})
if err := ws.WriteMessage(websocket.TextMessage, start); err != nil {
    return err
}
```

## Compatibility & Versioning

The wire protocol is versioned by URL path. v1 is the current shipping
version.

* Additive changes (new optional fields, new event types older clients
  can ignore) ship under v1.
* Breaking changes (renamed events, removed fields, repurposed
  semantics) will live at `/v2/audio/speech/stream` and run alongside
  v1 until v1 is retired.

The `event` discriminator on both directions is deliberately
extensible: clients should ignore unknown server events rather than
fail, and servers should reject unknown client events with
`unknown_event`.

## Open Questions / Future Work

* [ ] Surface the Volcengine binary `errorCode` (from `msgType=0b1111`
      frames) inside the JSON `error` event payload so clients can
      distinguish auth, quota, and malformed-session errors without
      parsing provider-specific text.
* [ ] Add ElevenLabs / Cartesia / Azure backends behind the same wire.
* [ ] `context_id` multiplexing for parallel sessions on one
      connection (v2).
* [ ] Heartbeat / `ping` event for idle keep-alive on long pauses
      between client `text` frames.
* [ ] Surface `session.canceled` ack from upstream after `cancel`.

## References

* Volcengine v3 bidirectional TTS:
  <https://www.volcengine.com/docs/6561/1329505>
* OpenAI `/v1/audio/speech` request shape (the REST sibling this
  protocol extends): <https://platform.openai.com/docs/api-reference/audio/createSpeech>
* Cartesia TTS WebSocket (reference design for `context_id`):
  <https://docs.cartesia.ai/api-reference/tts/websocket>
* Deepgram Aura WebSocket (reference for binary audio frames):
  <https://developers.deepgram.com/docs/tts-websocket>
