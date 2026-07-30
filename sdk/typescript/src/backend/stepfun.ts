import type { SpeechProviderWithExtraOptions } from '@xsai-ext/providers/utils'

import type { UnSpeechOptions, VoiceProviderWithExtraOptions } from '../types'

import { merge } from '@xsai-ext/providers/utils'
import { objCamelToSnake } from '@xsai/shared'

export interface UnStepfunOptions {
  /**
   * Selects a StepFun-defined API endpoint without exposing an arbitrary URL.
   *
   * Use `step-plan` for requests covered by a Step Plan subscription. Omit
   * this option, or use `default`, for StepFun's default speech endpoint.
   *
   * @default 'default'
   */
  endpointProfile?: 'default' | 'step-plan'
  /**
   * Audio volume. Range: 0.1 to 2.0.
   *
   * @default 1
   */
  volume?: number
  /**
   * Voice labels for `step-tts-2` and `step-tts-mini`.
   *
   * Not supported by `stepaudio-2.5-tts`; use `instruction` instead.
   */
  voiceLabel?: {
    language?: '粤语' | '四川话' | '日语' | string
    emotion?: string
    style?: string
  }
  /**
   * Global natural-language guidance for `stepaudio-2.5-tts`.
   */
  instruction?: string
  /**
   * Output sample rate.
   *
   * @default 24000
   */
  sampleRate?: 8000 | 16000 | 22050 | 24000 | 48000 | number
  pronunciationMap?: {
    tone: string[]
  }
  streamFormat?: 'audio' | 'sse'
  markdownFilter?: boolean
}

function toUnSpeechOptions(options: UnStepfunOptions): UnSpeechOptions {
  return { extraBody: objCamelToSnake({ ...options }) }
}

/**
 * [StepFun / 阶跃星辰](https://platform.stepfun.com/) provider for
 * [UnSpeech](https://github.com/moeru-ai/unspeech).
 *
 * @param apiKey - StepFun API key
 * @param baseURL - UnSpeech Instance URL
 * @returns SpeechProviderWithExtraOptions
 */
export function createUnStepfun(apiKey: string, baseURL = 'http://localhost:5933/v1/') {
  const speechProvider: SpeechProviderWithExtraOptions<
    'stepfun/stepaudio-2.5-tts' | 'stepfun/step-tts-2' | 'stepfun/step-tts-mini' | `stepfun/${string}`,
    UnStepfunOptions
  > = {
    speech: (model, options) => ({
      ...(options ? toUnSpeechOptions(options) : {}),
      apiKey,
      baseURL,
      model,
    }),
  }

  const voiceProvider: VoiceProviderWithExtraOptions<
    UnStepfunOptions
  > = {
    voice: (options) => {
      if (baseURL.endsWith('v1/')) {
        baseURL = baseURL.slice(0, -3)
      }
      else if (baseURL.endsWith('v1')) {
        baseURL = baseURL.slice(0, -2)
      }

      return {
        query: 'provider=stepfun',
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
