import type { SpeechProviderWithExtraOptions } from '@xsai-ext/shared-providers'
import type { VoiceProviderWithExtraOptions } from '../types'

import { merge } from '@xsai-ext/shared-providers'

/** @see {@link https://developers.deepgram.com/docs/text-to-speech} */
export interface UnDeepgramOptions {}

/**
 * [Deepgram](https://deepgram.com/) provider for [UnSpeech](https://github.com/moeru-ai/unspeech)
 * only.
 *
 * [UnSpeech](https://github.com/moeru-ai/unspeech) is a open-source project that provides a
 * OpenAI-compatible audio & speech related API that can be used with various providers such
 * as ElevenLabs, Azure TTS, Google TTS, etc.
 *
 * @param apiKey - Deepgram API Key
 * @param baseURL - UnSpeech Instance URL
 * @returns SpeechProviderWithExtraOptions
 */
export function createUnDeepgram(apiKey: string, baseURL = 'http://localhost:5933/v1/') {
  const speechProvider: SpeechProviderWithExtraOptions<
    /** @see {@link https://developers.deepgram.com/docs/tts-models} */
    `deepgram/${string}` | string,
    UnDeepgramOptions
  > = {
    speech: (model, _options) => ({
      apiKey,
      baseURL,
      model: model.startsWith('deepgram/') ? model : `deepgram/${model}`,
    }),
  }

  const voiceProvider: VoiceProviderWithExtraOptions<
    UnDeepgramOptions
  > = {
    voice: (_options) => {
      if (baseURL.endsWith('v1/')) {
        baseURL = baseURL.slice(0, -3)
      }
      else if (baseURL.endsWith('v1')) {
        baseURL = baseURL.slice(0, -2)
      }

      return {
        query: 'provider=deepgram',
        apiKey,
        baseURL,
      }
    },
  }

  return merge(
    speechProvider,
    voiceProvider,
  )
}
