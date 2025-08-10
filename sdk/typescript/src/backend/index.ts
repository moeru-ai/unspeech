import { createSpeechProviderWithExtraOptions, merge } from '@xsai-ext/shared-providers'

import type { UnSpeechOptions, VoiceProviderWithExtraOptions } from '../types'
import type { MicrosoftRegions } from './microsoft'

export * from './alibabacloud'
export * from './elevenlabs'
export * from './microsoft'
export * from './volcengine'

/** @see {@link https://github.com/moeru-ai/unspeech} */
export const createUnSpeech = (apiKey: string, baseURL = 'http://localhost:5933/v1/') => {
  const voiceProvider: VoiceProviderWithExtraOptions<
    {
      appId: string
      backend: 'volcano'
    } | {
      appId: string
      backend: 'volcengine'
    } | {
      backend: 'azure' | 'microsoft'
      region: MicrosoftRegions | string
    } | {
      backend:
        | 'ali'
        | 'alibaba'
        | 'alibaba-model-studio'
        | 'aliyun' | 'bailian' | 'elevenlabs' | 'koemotion' | 'openai'
    }
  > = {
    voice: (options) => {
      if (baseURL.endsWith('v1/')) {
        baseURL = baseURL.slice(0, -3)
      }
      else if (baseURL.endsWith('v1')) {
        baseURL = baseURL.slice(0, -2)
      }

      if (options?.backend === 'microsoft' || options?.backend === 'azure') {
        return {
          apiKey,
          baseURL,
          query: `region=${options.region}&provider=${options.backend}`,
        }
      }

      return {
        apiKey,
        baseURL,
        query: `provider=${options?.backend}`,
      }
    },
  }

  return merge(
    createSpeechProviderWithExtraOptions<
      | `alibaba/${string}`
      | `aliyun/${string}`
      | `elevenlabs/${string}`
      | `koemotion/${string}`
      | `openai/${string}`
      | `volcano/${string}`
      | `volcengine/${string}`,
    UnSpeechOptions
      >({ apiKey, baseURL }),
    voiceProvider,
  )
}
