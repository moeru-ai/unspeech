import type { SpeechProviderWithExtraOptions } from '@xsai-ext/providers/utils'

import type { UnSpeechOptions, VoiceProviderWithExtraOptions } from '../types'

import { merge } from '@xsai-ext/providers/utils'
import { objCamelToSnake } from '@xsai/shared'

export interface UnAlibabaCloudOptions {
  /**
   * Speech pitch. Range: 0.5 to 2.0.
   * @default 1.0
   */
  pitch?: number
  /**
   * Speech rate. Range: 0.5 to 2.0.
   * @default 1.0
   */
  rate?: number
  /**
   * Sampling rate of the synthesized audio.
   * @default 22050
   */
  sampleRate?: 8000 | 16000 | 22050 | 24000 | 44100 | 48000 | number
  /**
   * Volume of the synthesized audio. Range: 0 to 100.
   * @default 50
   */
  volume?: number
}

/**
 * [Alibaba Cloud / 阿里云 通义听悟](https://tingwu.aliyun.com/) provider for [UnSpeech](https://github.com/moeru-ai/unspeech)
 * only.
 *
 * [UnSpeech](https://github.com/moeru-ai/unspeech) is a open-source project that provides a
 * OpenAI-compatible audio & speech related API that can be used with various providers such
 * as ElevenLabs, Azure TTS, Google TTS, etc.
 *
 * @param apiKey - Alibaba Cloud AccessKey Token (see https://help.aliyun.com/document_detail/72153.html)
 * @param baseURL - UnSpeech Instance URL
 * @returns SpeechProviderWithExtraOptions & VoiceProviderWithExtraOptions
 */
export function createUnAlibabaCloud(apiKey: string, baseURL = 'http://localhost:5933/v1/') {
  const toUnSpeechOptions = (options: UnAlibabaCloudOptions): UnSpeechOptions => {
    const { pitch, rate, sampleRate, volume } = options

    const extraBody: Record<string, unknown> = {
      pitch,
      rate,
      sampleRate,
      volume,
    }

    // Remove undefined values before converting keys
    Object.keys(extraBody).forEach(key => extraBody[key] === undefined && delete extraBody[key])

    return { extraBody: objCamelToSnake(extraBody) }
  }

  const speechProvider: SpeechProviderWithExtraOptions<
    /** @see https://help.aliyun.com/document_detail/451271.html */
    'alibaba/v1',
    UnAlibabaCloudOptions
  > = {
    speech: (model, options) => ({
      ...(options ? toUnSpeechOptions(options) : {}),
      apiKey,
      baseURL,
      // Model is used directly, e.g., "sifeng"
      // The backend prepends the necessary task group/task/function info
      model: `alibaba/${model}`,
    }),
  }

  const voiceProvider: VoiceProviderWithExtraOptions<
    UnAlibabaCloudOptions
  > = {
    voice: (options) => {
      let adjustedBaseURL = baseURL
      if (adjustedBaseURL.endsWith('v1/')) {
        adjustedBaseURL = adjustedBaseURL.slice(0, -3)
      }
      else if (adjustedBaseURL.endsWith('v1')) {
        adjustedBaseURL = adjustedBaseURL.slice(0, -2)
      }

      return {
        query: 'provider=alibaba',
        ...(options ? toUnSpeechOptions(options) : {}),
        apiKey,
        baseURL: adjustedBaseURL,
      }
    },
  }

  return merge(
    speechProvider,
    voiceProvider,
  )
}
