import type { SpeechProviderWithExtraOptions } from '@xsai-ext/providers/utils'

import type { UnSpeechOptions, VoiceProviderWithExtraOptions } from '../types'

import { merge } from '@xsai-ext/providers/utils'
import { objCamelToSnake } from '@xsai/shared'

/**
 * MiniMax TTS API options
 * @see https://platform.minimaxi.com/docs/guides/speech-t2a-http
 */
export interface UnMinimaxOptions {
  /**
   * Speech speed. Range: 0.5-2.0
   * @default 1.0
   */
  speed?: number
  /**
   * Volume. Range: 0-10
   * @default 1.0
   */
  vol?: number
  /**
   * Pitch adjustment. Range: -12 to 12
   * @default 0
   */
  pitch?: number
  /**
   * Emotion setting
   * @see https://platform.minimaxi.com/docs/guides/speech-t2a-http#emotion
   * @example "happy" | "sad" | "angry" | "fearful" | "disgusted" | "surprised" | "calm" | "fluent" | "whisper"
   */
  emotion?: string
  /**
   * Enable streaming output
   * @default false
   */
  stream?: boolean
  /**
   * Sample rate for audio output
   * @example 8000 | 16000 | 22050 | 24000 | 32000 | 44100
   */
  sampleRate?: number
  /**
   * Audio bitrate
   * @example 32000 | 64000 | 128000 | 256000
   */
  bitrate?: number
  /**
   * Audio format
   * @example "mp3" | "pcm" | "flac" | "wav"
   * @default "mp3"
   */
  format?: 'mp3' | 'pcm' | 'flac' | 'wav'
  /**
   * Audio channel
   * @example 1 | 2
   * @default 1
   */
  channel?: number
}

/**
 * MiniMax TTS models
 * @see https://platform.minimaxi.com/docs/guides/speech-t2a-http#model
 */
export type MinimaxModel
  = | 'speech-2.8-hd'
    | 'speech-2.8-turbo'
    | 'speech-2.6-hd'
    | 'speech-2.6-turbo'
    | 'speech-02-hd'
    | 'speech-02-turbo'
    | 'speech-01-hd'
    | 'speech-01-turbo'

/**
 * [MiniMax](https://platform.minimaxi.com/) provider for [UnSpeech](https://github.com/moeru-ai/unspeech)
 *
 * @param apiKey - MiniMax API Key
 * @param baseURL - UnSpeech Instance URL
 * @returns SpeechProviderWithExtraOptions & VoiceProviderWithExtraOptions
 */
export function createUnMinimax(apiKey: string, baseURL = 'http://localhost:5933/v1/') {
  const toUnSpeechOptions = ({
    speed,
    vol,
    pitch,
    emotion,
    stream,
    sampleRate,
    bitrate,
    format,
    channel,
  }: UnMinimaxOptions): UnSpeechOptions => ({
    extraBody: objCamelToSnake({
      speed,
      vol,
      pitch,
      emotion,
      stream,
      sampleRate,
      bitrate,
      format,
      channel,
    }),
  })

  const speechProvider: SpeechProviderWithExtraOptions<
    `minimax/${MinimaxModel}`,
    UnMinimaxOptions
  > = {
    speech: (model, options) => ({
      ...(options ? toUnSpeechOptions(options) : {}),
      apiKey,
      baseURL,
      model,
    }),
  }

  const voiceProvider: VoiceProviderWithExtraOptions<
    UnMinimaxOptions
  > = {
    voice: (options) => {
      let adjustedBaseURL = baseURL
      if (adjustedBaseURL.endsWith('v1/')) {
        adjustedBaseURL = adjustedBaseURL.slice(0, -3)
      }
      else if (adjustedBaseURL.endsWith('v1')) {
        adjustedBaseURL = adjustedBaseURL.slice(0, -2)
      }

      return {
        query: 'provider=minimax',
        ...(options ? toUnSpeechOptions(options) : {}),
        apiKey,
        baseURL: adjustedBaseURL,
      }
    },
  }

  return merge(
    speechProvider,
    voiceProvider,
  )
}
