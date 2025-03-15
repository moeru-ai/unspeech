import type { CommonRequestOptions } from '@xsai/shared'

import { requestHeaders, requestURL, responseJSON } from '@xsai/shared'

import type { Voice } from '../types/voice'

export interface ListVoicesOptions extends Omit<CommonRequestOptions, 'model'> {
  query?: string
}

export interface ListVoicesResponse {
  voices: Voice[]
}

export const listVoices = async (options: ListVoicesOptions): Promise<Voice[]> =>
  (options.fetch ?? globalThis.fetch)(requestURL(options.query ? `api/voices?${options.query}` : 'api/voices', options.baseURL), {
    headers: requestHeaders({ ...options.headers }, options.apiKey),
    method: 'GET',
    signal: options.abortSignal,
  })
    .then(responseJSON<ListVoicesResponse>)
    .then(({ voices }) => voices)
