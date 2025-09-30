import type { SpeechProviderWithExtraOptions } from '@xsai-ext/shared-providers'

import { merge } from '@xsai-ext/shared-providers'
import { objCamelToSnake } from '@xsai/shared'

import type { UnSpeechOptions, VoiceProviderWithExtraOptions } from '../types'

export interface UnVolcengineOptions {
  app?: {
    appId?: string
    cluster?: 'volcano_tts' | string
  }
  audio?: {
    /**
     * @default 160
     */
    bitRate?: 160 | number
    /**
     * Languages that contextual to the model
     */
    contextLanguage?: 'es' | 'id' | 'pt' | string
    emotion?: 'angry' | string
    /**
     * After calling emotion to set the emotion parameter you can use emotion_scale to
     * further set the emotion value, the range is 1~5, the default value is 4 when not
     * set.
     *
     * Note: Theoretically, the larger the emotion value is, the more obvious the emotion
     * is. However, the emotion value 1~5 is actually non-linear growth, there may be
     * more than a certain value, the increase in emotion is not obvious, for example,
     * set 3 and 5 when the emotion value may be close.
     *
     * 1~5
     *
     * @default 4
     */
    emotionScale?: number
    enableEmotion?: boolean
    /**
     * @default 'mp3'
     */
    encoding?: 'mp3' | 'ogg_opus' | 'pcm' | 'wav'
    /**
     * - undefined: General mixed bilingual
     * - crosslingual: mix with zh/en/ja/es-ms/id/pt-br
     * - zh: primarily Chinese, supports mixed Chinese and English
     * - en: only English
     * - ja: only Japanese
     * - es-mx: only Mexican Spanish
     * - id: only Indonesian
     * - pt-br: only Brazilian Portuguese
     *
     * @default 'en'
     */
    explicitLanguage?: 'crosslingual' | 'en' | 'es-mx' | 'id' | 'jp' | 'pt-br' | 'zh' | string
    /**
     * 0.5 ~ 2
     *
     * @default 1
     */
    loudnessRatio?: number
    /**
     * @default 24000
     */
    rate?: 8000 | 16000 | 24000 | number
    /**
     * 0.8~2
     *
     * @default 1
     */
    speedRatio?: number
  }
  request?: {
    cacheConfig?: Record<string, unknown>
    disableMarkdownFilter?: boolean
    enableLatexTone?: boolean
    extraParam?: string
    reqid?: string
    /**
     * 0 ~ 30000ms
     */
    silenceDuration?: number
    /**
     * - set to `ssml` to use SSML
     */
    textType?: 'ssml' | string
    useCache?: boolean
    withTimestamp?: string
  }
  user?: {
    uid?: string
  }
}

/**
 * [Volcengine / 火山引擎](https://www.volcengine.com/docs/6561/162929) provider for [UnSpeech](https://github.com/moeru-ai/unspeech)
 * only.
 *
 * [UnSpeech](https://github.com/moeru-ai/unspeech) is a open-source project that provides a
 * OpenAI-compatible audio & speech related API that can be used with various providers such
 * as ElevenLabs, Azure TTS, Google TTS, etc.
 *
 * @param apiKey - Volcano Engine Speech Service Token
 * @param baseURL - UnSpeech Instance URL
 * @returns SpeechProviderWithExtraOptions
 */
export const createUnVolcengine = (apiKey: string, baseURL = 'http://localhost:5933/v1/') => {
  const toUnSpeechOptions = (options: UnVolcengineOptions): UnSpeechOptions => {
    const extraBody: Record<string, unknown> = {
      app: {
        appid: options.app?.appId,
        token: apiKey,
      },
    }

    if (typeof options.app !== 'undefined') {
      extraBody.app = {
        ...options.app,
        appid: options.app?.appId,
        token: apiKey,
      }
    }
    if (typeof options.user !== 'undefined') {
      extraBody.user = options.user
    }
    if (typeof options.audio !== 'undefined') {
      extraBody.audio = options.audio
    }

    return { extraBody: objCamelToSnake(extraBody) }
  }

  const speechProvider: SpeechProviderWithExtraOptions<
    /** @see Currently, only v1 is available */
    'volcengine/v1',
    UnVolcengineOptions
  > = {
    speech: (model, options) => ({
      ...(options ? toUnSpeechOptions(options) : {}),
      apiKey,
      baseURL,
      model: `volcengine/${model}`,
    }),
  }

  const voiceProvider: VoiceProviderWithExtraOptions<
    UnVolcengineOptions
  > = {
    voice: (options) => {
      if (baseURL.endsWith('v1/')) {
        baseURL = baseURL.slice(0, -3)
      }
      else if (baseURL.endsWith('v1')) {
        baseURL = baseURL.slice(0, -2)
      }

      return {
        query: 'provider=volcengine',
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
