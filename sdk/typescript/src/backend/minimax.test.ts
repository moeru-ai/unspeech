/* eslint-disable test/no-import-node-test -- Tests use the existing tsx dependency instead of adding a new runner. */
import assert from 'node:assert/strict'

import { describe, it } from 'node:test'

import { createUnMinimax } from './minimax'

describe('createUnMinimax', () => {
  it('preserves the provider-prefixed model name', () => {
    const provider = createUnMinimax('test-key')

    assert.equal(provider.speech('minimax/speech-2.8-turbo').model, 'minimax/speech-2.8-turbo')
  })

  it('does not mutate the speech base URL after listing voices', () => {
    const provider = createUnMinimax('test-key', 'http://localhost:5933/v1/')

    assert.equal(provider.voice().baseURL, 'http://localhost:5933/')
    assert.equal(provider.speech('minimax/speech-2.8-turbo').baseURL, 'http://localhost:5933/v1/')
  })
})
