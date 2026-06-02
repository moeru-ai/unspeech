import type { CommonRequestOptions, WithUnknown } from '@xsai/shared'

import { requestBody, requestHeaders, requestURL } from '@xsai/shared'

type StringFetch = (input: string, init: RequestInit) => Promise<Response>

export interface GenerateSpeechResponseOptions extends Omit<CommonRequestOptions, 'fetch'> {
  fetch?: StringFetch | typeof globalThis.fetch
  input: string
  /** @default `mp3` */
  responseFormat?: 'aac' | 'flac' | 'mp3' | 'opus' | 'pcm' | 'wav' | string
  /** @default `1.0` */
  speed?: number
  voice: string
}

export interface GenerateSpeechResponseResult {
  body: ArrayBuffer
  contentType: null | string
  response: Response
}

export interface UnSpeechAPIErrorOptions {
  requestBody: string
  response: Response
  responseBody: string
}

/**
 * Error thrown when unspeech returns a non-2xx response.
 *
 * Use when:
 * - Callers need to preserve unspeech/upstream status codes for fallback,
 *   retries, or diagnostics.
 *
 * Expects:
 * - The caller has already read the response body text.
 *
 * Returns:
 * - An Error carrying status, response body, headers, and URL metadata.
 */
export class UnSpeechAPIError extends Error {
  requestBody: string
  response: Response
  responseBody: string
  responseHeaders: Record<string, string>
  status: number
  statusCode: number
  statusText: string
  url: string

  constructor(message: string, options: UnSpeechAPIErrorOptions) {
    super(message)
    this.name = 'UnSpeechAPIError'
    this.requestBody = options.requestBody
    this.response = options.response
    this.responseBody = options.responseBody
    this.responseHeaders = Object.fromEntries(options.response.headers.entries())
    this.status = options.response.status
    this.statusCode = options.response.status
    this.statusText = options.response.statusText
    this.url = options.response.url
  }
}

/**
 * Generates speech through unspeech and preserves response metadata.
 *
 * Use when:
 * - A server proxy needs the binary audio bytes plus headers such as
 *   `content-type`.
 * - A router needs non-2xx status/body metadata for fallback decisions.
 *
 * Expects:
 * - `baseURL` points at an unspeech v1 endpoint.
 * - Provider-specific options are passed through as xsAI-compatible fields
 *   such as `extraBody`, which are serialized into OpenAI-style snake_case.
 *
 * Returns:
 * - The generated audio bytes, the `content-type` header when present, and
 *   the original `Response`.
 */
export async function generateSpeechResponse(options: WithUnknown<GenerateSpeechResponseOptions>): Promise<GenerateSpeechResponseResult> {
  const body = requestBody(options)
  const fetchImpl = options.fetch ?? globalThis.fetch
  const response = await fetchImpl(requestURL('audio/speech', options.baseURL).toString(), {
    body,
    headers: requestHeaders({
      'Content-Type': 'application/json',
      ...options.headers,
    }, options.apiKey),
    method: 'POST',
    signal: options.abortSignal,
  })

  if (!response.ok) {
    const responseBody = await response.text().catch(() => '')
    throw new UnSpeechAPIError(`unspeech audio speech ${response.status}: ${responseBody.slice(0, 256)}`, {
      requestBody: body,
      response,
      responseBody,
    })
  }

  return {
    body: await response.arrayBuffer(),
    contentType: response.headers.get('content-type'),
    response,
  }
}
