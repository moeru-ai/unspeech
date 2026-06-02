// Compile-only regression test: spreading provider.voice() into listVoices() must type-check.
import type { Fetch } from '@xsai/shared'
import type { ListVoicesOptions } from '../list-voices'

// Simulate what provider.voice() returns (CommonRequestOptions minus 'model')
declare const voiceResult: {
  baseURL: string
  apiKey?: string
  fetch?: Fetch | typeof globalThis.fetch
}

// This assignment must not produce TS2345
const opts: ListVoicesOptions = { ...voiceResult }
void opts
