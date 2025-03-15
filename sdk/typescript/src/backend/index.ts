import { createMetadataProvider, createSpeechProviderWithExtraOptions, merge } from '@xsai-ext/shared-providers'

import { MicrosoftRegions } from './microsoft'
import { UnSpeechOptions, VoiceProviderWithExtraOptions } from '../types'

export * from './elevenlabs'
export * from './microsoft'

/** @see {@link https://github.com/moeru-ai/unspeech} */
export const createUnSpeech = (apiKey: string, baseURL = 'http://localhost:5933/v1/') => {
  const voiceProvider: VoiceProviderWithExtraOptions<
    {
  backend: 'elevenlabs' | 'koemotion' | 'openai'
} | {
  backend: 'microsoft' | 'azure'
  region: MicrosoftRegions | string
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
  createMetadataProvider('unspeech'),
  createSpeechProviderWithExtraOptions<
    | `elevenlabs/${string}`
    | `koemotion/${string}`
    | `openai/${string}`,
    UnSpeechOptions
      >({ apiKey, baseURL }),
    voiceProvider
  )
}
