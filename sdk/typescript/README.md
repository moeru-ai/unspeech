# unSpeech TypeScript Client

> unSpeech, Your Text-to-Speech Services, All-in-One.

## Install

```bash
npm i unspeech
```

## Getting Started

### List voices

Besides of the `/audio/speech` endpoint, we support listing all the available voices from providers as well:

```ts
import { createUnSpeech, listVoices } from 'unspeech'

const unspeech = createUnSpeech('YOUR_EXTERNAL_PROVIDER_API_KEY', 'http://localhost:5933/v1/')

const voices = await listVoices({
  ...unspeech.voice({ backend: 'elevenlabs' })
})
```

### Speech synthesis

For general purpose `/audio/speech` requests, `@xsai/generate-speech` or xsAI can be used as it's compatible:

```bash
npm i @xsai/generate-speech
```

```ts
import { generateSpeech } from '@xsai/generate-speech'
import { createUnSpeech } from 'unspeech'

const unspeech = createUnSpeech('YOUR_EXTERNAL_PROVIDER_API_KEY', 'http://localhost:5933/v1/')
const speech = await generateSpeech(
  ...unspeech.speech('elevenlabs/eleven_multilingual_v2'),
  voice: '9BWtsMINqrJLrRacOk9x',
  input: 'Hello, World!',
)
```

For the other providers, you can import them as needed

```ts
import {
  createUnElevenLabs,
  createUnMicrosoft,
  createUnSpeech,
  createUnAlibabaCloud,
  createUnVolcengine
} from 'unspeech'
```

When using

- [Microsoft / Azure AI Speech service](https://learn.microsoft.com/en-us/azure/ai-services/speech-service/text-to-speech)
- [Alibaba Cloud Model Studio / 阿里云百炼 / CosyVoice](https://www.alibabacloud.com/en/product/modelstudio)
- [Volcano Engine / 火山引擎语音技术](https://www.volcengine.com/product/voice-tech)
- [ElevenLabs](https://elevenlabs.io/docs/api-reference/text-to-speech/convert)

providers, [SSML](https://learn.microsoft.com/en-us/azure/ai-services/speech-service/speech-synthesis-markup) is supported to control in fine grain level for pitch, volume, rate, etc.

## Related Projects

Looking for something like unSpeech, but for local TTS? check it out:

- [erew123/alltalk_tts/alltalkbeta](https://github.com/erew123/alltalk_tts/tree/alltalkbeta)
- [astramind-ai/Auralis](https://github.com/astramind-ai/Auralis)
- [matatonic/openedai-speech](https://github.com/matatonic/openedai-speech)

Or to use free Edge TTS:

- [travisvn/openai-edge-tts](https://github.com/travisvn/openai-edge-tts)

## Similar Projects

- [orate](https://github.com/haydenbleasel/orate)

## License

[AGPL-3.0](./LICENSE)
