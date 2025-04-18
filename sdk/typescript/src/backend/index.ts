import { createSpeechProviderWithExtraOptions, merge } from '@xsai-ext/shared-providers'

import { MicrosoftRegions } from './microsoft'
import { UnSpeechOptions, VoiceProviderWithExtraOptions } from '../types'

export * from './elevenlabs'
export * from './microsoft'
export * from './volcengine'
export * from './alibabacloud'

/** @see {@link https://github.com/moeru-ai/unspeech} */
export const createUnSpeech = (apiKey: string, baseURL = 'http://localhost:5933/v1/') => {
  const voiceProvider: VoiceProviderWithExtraOptions<
    {
  backend:
  | 'elevenlabs'
  | 'koemotion'
  | 'openai'
  | 'alibaba' | 'aliyun' | 'ali' | 'bailian' | 'alibaba-model-studio'
} | {
  backend: 'microsoft' | 'azure'
  region: MicrosoftRegions | string
} | {
  backend: 'volcengine'
  appId: string
} | {
  backend: 'volcano'
  appId: string
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
          query: `region=${options.region}&provider=${options.backend}`,
          baseURL,
          apiKey,
        }
      }

      return {
        query: `provider=${options?.backend}`,
        baseURL,
        apiKey,
      }
    },
  }

  return merge(
  createSpeechProviderWithExtraOptions<
    | `elevenlabs/${string}`
    | `koemotion/${string}`
    | `openai/${string}`
    | `volcengine/${string}`
    | `volcano/${string}`
    | `aliyun/${string}`
    | `alibaba/${string}`,
    UnSpeechOptions
      >({ apiKey, baseURL }),
    voiceProvider
  )
}
