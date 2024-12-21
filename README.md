# unSpeech

> Your Text-to-Speech Services, All-in-One.

## Features

unSpeech lets you use other online TTS in OpenAI-compatible format.

- OpenAI
- ElevenLabs

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

## License

[AGPL-3.0](./LICENSE)
