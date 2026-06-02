import type { CommonRequestOptions } from '@xsai/shared'

import type { Voice } from '../types/voice'

import { requestHeaders, requestURL } from '@xsai/shared'

import { UnSpeechAPIError } from './generate-speech-response'

type StringFetch = (input: string, init: RequestInit) => Promise<Response>

export interface ListVoicesOptions extends Omit<CommonRequestOptions, 'fetch' | 'model'> {
  fetch?: StringFetch | typeof globalThis.fetch
  query?: string
}

export interface ListVoicesResponse {
  voices: Voice[]
}

export async function listVoices(options: ListVoicesOptions): Promise<Voice[]> {
  const fetchImpl = options.fetch ?? globalThis.fetch
  const response = await fetchImpl(requestURL(options.query ? `api/voices?${options.query}` : 'api/voices', options.baseURL).toString(), {
    headers: requestHeaders({ ...options.headers }, options.apiKey),
    method: 'GET',
    signal: options.abortSignal,
  })

  if (!response.ok) {
    const responseBody = await response.text().catch(() => '')
    throw new UnSpeechAPIError(`unspeech voices ${response.status}: ${responseBody.slice(0, 256)}`, {
      requestBody: '',
      response,
      responseBody,
    })
  }

  const data = await response.json() as ListVoicesResponse
  return data.voices
}
