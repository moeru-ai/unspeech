import type { SpeechProviderWithExtraOptions } from '@xsai-ext/shared-providers'

import { merge } from '@xsai-ext/shared-providers'
import { objCamelToSnake } from '@xsai/shared'

import type { UnSpeechOptions, VoiceProviderWithExtraOptions } from '../types'

export interface UnVolcengineOptions {
  app?: {
    appId?: string
    cluster?: string | 'volcano_tts'
  }
  user?: {
    uid?: string
  }
  audio?: {
    emotion?: string | 'angry'
    enableEmotion?: boolean
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
    /**
     * @default 'mp3'
     */
    encoding?: 'wav' | 'pcm' | 'ogg_opus' | 'mp3'
    /**
     * 0.8~2
     *
     * @default 1
     */
    speedRatio?: number
    /**
     * @default 24000
     */
    rate?: number | 24000 | 8000 | 16000
    /**
     * @default 160
     */
    bitRate?: number | 160
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
    explicitLanguage?: string | 'crosslingual' | 'zh' | 'en' | 'jp' | 'es-mx' | 'id' | 'pt-br'
    /**
     * Languages that contextual to the model
     */
    contextLanguage?: string | 'id' | 'es' | 'pt'
    /**
     * 0.5 ~ 2
     *
     * @default 1
     */
    loudnessRatio?: number
  }
  request?: {
    reqid?: string
    /**
     * - set to `ssml` to use SSML
     */
    textType?: string | 'ssml'
    /**
     * 0 ~ 30000ms
     */
    silenceDuration?: number
    withTimestamp?: string
    extraParam?: string
    disableMarkdownFilter?: boolean
    enableLatexTone?: boolean
    cacheConfig?: Record<string, unknown>
    useCache?: boolean
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
 * @param appId - Volcano Engine Speech Service App ID
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
        query: `provider=volcengine`,
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
