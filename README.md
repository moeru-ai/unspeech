# unSpeech

> Your Text-to-Speech Services, All-in-One.

## Features

unSpeech lets you use other online TTS in OpenAI-compatible format.

- [OpenAI](https://platform.openai.com/docs/api-reference/audio/createSpeech)
- [ElevenLabs](https://elevenlabs.io/docs/api-reference/text-to-speech/convert)

## Getting Started

### Build

unSpeech is not released yet, you need to build it from source code:

> require `go` 1.23+

```bash
git clone https://github.com/moeru-ai/unspeech.git
cd unspeech
go build -o ./result/unspeech ./cmd/unspeech
```

### Run

```bash
# http server started on [::]:5933
./result/unspeech
```

### Client

You can use unSpeech with most OpenAI clients.

###### [`@xsai/generate-speech`](https://github.com/moeru-ai/xsai)

```ts
import { generateSpeech } from '@xsai/generate-speech'

const speech = await generateSpeech({
  apiKey: 'YOUR_API_KEY',
  baseURL: 'http://localhost:5933/v1/',
  input: 'Hello, World!',
  model: 'elevenlabs/eleven_multilingual_v2',
  voice: '9BWtsMINqrJLrRacOk9x',
})
```

## Related Projects

Looking for something like unSpeech, but for local TTS? check it out:

- [erew123/alltalk_tts/alltalkbeta](https://github.com/erew123/alltalk_tts/tree/alltalkbeta)
- [astramind-ai/Auralis](https://github.com/astramind-ai/Auralis)
- [matatonic/openedai-speech](https://github.com/matatonic/openedai-speech)

Or to use Edge TTS, which unSpeech is not intended to support:

- [travisvn/openai-edge-tts](https://github.com/travisvn/openai-edge-tts)

## License

[AGPL-3.0](./LICENSE)
