import { describe, expect, it } from 'vitest'

import { createUnMinimax } from './minimax'

describe('createUnMinimax', () => {
  it('preserves the provider-prefixed model name', () => {
    const provider = createUnMinimax('test-key')

    expect(provider.speech('minimax/speech-2.8-turbo').model).toBe('minimax/speech-2.8-turbo')
  })

  it('does not mutate the speech base URL after listing voices', () => {
    const provider = createUnMinimax('test-key', 'http://localhost:5933/v1/')

    expect(provider.voice().baseURL).toBe('http://localhost:5933/')
    expect(provider.speech('minimax/speech-2.8-turbo').baseURL).toBe('http://localhost:5933/v1/')
  })
})
