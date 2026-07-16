import type { SpeechProviderWithExtraOptions } from '@xsai-ext/providers/utils'

import type { UnSpeechOptions, VoiceProviderWithExtraOptions } from '../types'

import { merge } from '@xsai-ext/providers/utils'
import { objCamelToSnake } from '@xsai/shared'

/** @see {@link https://docs.fish.audio/api-reference/endpoint/openapi-v1/text-to-speech} */
export interface UnFishAudioOptions {
  /**
   * Text segment length used when splitting long input.
   * Range: 100 to 300.
   *
   * @default 200
   */
  chunkLength?: number
  /**
   * Latency mode. `balanced` reduces latency at a slight cost to stability.
   *
   * @default 'normal'
   */
  latency?: 'balanced' | 'normal'
  /**
   * MP3 bitrate in kbps, only used with the mp3 format.
   *
   * @default 128
   */
  mp3Bitrate?: 64 | 128 | 192
  /**
   * Normalize text (e.g. spell out numbers) for better stability.
   *
   * @default true
   */
  normalize?: boolean
  /**
   * Speech pacing controls.
   */
  prosody?: {
    /**
     * Speed multiplier, 1.0 is normal speed.
     *
     * @default 1
     */
    speed?: number
    /**
     * Volume gain in dB, 0 is unchanged.
     *
     * @default 0
     */
    volume?: number
  }
  /**
   * Output sample rate in Hz.
   */
  sampleRate?: number
  /**
   * Sampling temperature; lower is more deterministic.
   *
   * @default 0.7
   */
  temperature?: number
  /**
   * Nucleus sampling threshold.
   *
   * @default 0.7
   */
  topP?: number
}

function toUnSpeechOptions(options: UnFishAudioOptions): UnSpeechOptions {
  return { extraBody: objCamelToSnake({ ...options }) }
}

/**
 * [Fish Audio](https://fish.audio/) provider for
 * [UnSpeech](https://github.com/moeru-ai/unspeech).
 *
 * The OpenAI `voice` parameter maps to a Fish Audio reference voice model id
 * (the id in a fish.audio voice page URL).
 *
 * @param apiKey - Fish Audio API key
 * @param baseURL - UnSpeech Instance URL
 * @returns SpeechProviderWithExtraOptions
 */
export function createUnFishAudio(apiKey: string, baseURL = 'http://localhost:5933/v1/') {
  const speechProvider: SpeechProviderWithExtraOptions<
    /** @see {@link https://docs.fish.audio/api-reference/endpoint/openapi-v1/text-to-speech} */
    'fishaudio/s1' | 'fishaudio/s2-pro' | 'fishaudio/s2.1-pro' | 'fishaudio/s2.1-pro-free' | `fishaudio/${string}`,
    UnFishAudioOptions
  > = {
    speech: (model, options) => ({
      ...(options ? toUnSpeechOptions(options) : {}),
      apiKey,
      baseURL,
      model,
    }),
  }

  const voiceProvider: VoiceProviderWithExtraOptions<
    UnFishAudioOptions
  > = {
    voice: (options) => {
      if (baseURL.endsWith('v1/')) {
        baseURL = baseURL.slice(0, -3)
      }
      else if (baseURL.endsWith('v1')) {
        baseURL = baseURL.slice(0, -2)
      }

      return {
        query: 'provider=fishaudio',
        ...(options ? toUnSpeechOptions(options) : {}),
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
