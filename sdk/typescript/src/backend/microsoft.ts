import type { SpeechProviderWithExtraOptions } from '@xsai-ext/providers/utils'

import type { UnSpeechOptions, VoiceProviderWithExtraOptions } from '../types'

import { merge } from '@xsai-ext/providers/utils'
import { objCamelToSnake } from '@xsai/shared'

export type MicrosoftRegions
  = | 'australiaeast'
    | 'brazilsouth'
    | 'canadacentral'
    | 'centralindia'
    | 'centralus'
    | 'eastasia'
    | 'eastus2'
    | 'eastus'
    | 'francecentral'
    | 'germanywestcentral'
    | 'japaneast'
    | 'japanwest'
    | 'jioindiawest'
    | 'koreacentral'
    | 'northcentralus'
    | 'northeurope'
    | 'norwayeast'
    | 'southcentralus'
    | 'southeastasia'
    | 'swedencentral'
    | 'switzerlandnorth'
    | 'switzerlandwest'
    | 'uaenorth'
    | 'uksouth'
    | 'usgovarizona'
    | 'usgovvirginia'
    | 'westcentralus'
    | 'westeurope'
    | 'westus2'
    | 'westus3'
    | 'westus'

export interface UnMicrosoftOptionAutoSSML {
  gender:
    | 'Female'
    | 'Male'
    | 'Neutral'
    | string
  lang:
    | 'en-US'
    | string
  /**
   * Speech Studio - Voice Gallery
   * https://speech.microsoft.com/portal/018ba84135d64cf79106cc99c75ffa6a/voicegallery
   */
  voice:
    | 'en-US-AndrewMultilingualNeural'
    | 'en-US-AriaNeural'
    | 'en-US-AvaMultilingualNeural'
    | 'en-US-BrianMultilingualNeural'
    | 'en-US-ChristopherMultilingualNeural'
    | 'en-US-EmmaMultilingualNeural'
    | 'en-US-JaneNeural'
    | string
}

export interface UnMicrosoftOptionCommon {
  /**
   * Text to speech API reference (REST) - Speech service - Azure AI services | Microsoft Learn
   * https://learn.microsoft.com/en-us/azure/ai-services/speech-service/rest-text-to-speech?tabs=streaming#custom-neural-voices
   */
  deploymentId?: string
  /**
   * Text to speech API reference (REST) - Speech service - Azure AI services | Microsoft Learn
   * https://learn.microsoft.com/en-us/azure/ai-services/speech-service/rest-text-to-speech?tabs=streaming#prebuilt-neural-voices
   *
   * NOTICE: Voices in preview are available in only these three regions: East US, West Europe, and Southeast Asia.
   */
  region: MicrosoftRegions | string
  sampleRate?:
    | 8000
    | 16000
    | 22050
    | 24000
    | 44100
    | 48000
    | number
}

export interface UnMicrosoftOptionCustomSSML {
  /**
   * By default, unspeech service will help you automatically convert OpenAI style plain text input
   * into SSML with lang, gender, voice parameters, but if you ever wanted to provide your own SSML
   * with all customizable parameters, you can set this option to `true` to disable the automatic
   * conversion and use your own SSML instead.
   *
   * About SSML (Speech Synthesis Markup Language), @see {@link https://learn.microsoft.com/en-us/azure/ai-services/speech-service/speech-synthesis-markup}
   */
  disableSsml?: boolean
}

/** @see {@link https://elevenlabs.io/docs/api-reference/text-to-speech/convert#request} */
export type UnMicrosoftOptions = (UnMicrosoftOptionAutoSSML | UnMicrosoftOptionCustomSSML) & UnMicrosoftOptionCommon

/**
 * Default Microsoft output format used when callers only ask for OpenAI-style
 * `mp3`. This keeps the common path on a compact 24kHz MP3 instead of the
 * backend's maximum-quality 48kHz default.
 */
export const DEFAULT_MICROSOFT_OUTPUT_FORMAT = 'audio-24khz-48kbitrate-mono-mp3'

const MICROSOFT_VOICE_ID = /^[a-z0-9-]+$/i

/**
 * Checks whether a Microsoft voice id is safe to embed in SSML attributes.
 *
 * Use when:
 * - A caller builds SSML locally before sending it to unspeech.
 *
 * Expects:
 * - Microsoft neural voice ids such as `en-US-AvaMultilingualNeural`.
 *
 * Returns:
 * - `true` for the canonical letters/digits/hyphen shape.
 */
export function isMicrosoftVoiceId(voice: string): boolean {
  return MICROSOFT_VOICE_ID.test(voice)
}

/**
 * Resolves OpenAI-style short audio format names to Microsoft output formats.
 *
 * Before:
 * - `"mp3"`
 * - `"wav"`
 * - `"audio-24khz-48kbitrate-mono-mp3"`
 *
 * After:
 * - `"audio-24khz-48kbitrate-mono-mp3"`
 * - `"riff-24khz-16bit-mono-pcm"`
 * - `"audio-24khz-48kbitrate-mono-mp3"`
 */
export function resolveMicrosoftOutputFormat(responseFormat: string | undefined): string {
  if (!responseFormat)
    return DEFAULT_MICROSOFT_OUTPUT_FORMAT
  if (responseFormat.includes('-'))
    return responseFormat
  if (responseFormat === 'mp3')
    return DEFAULT_MICROSOFT_OUTPUT_FORMAT
  if (responseFormat === 'wav')
    return 'riff-24khz-16bit-mono-pcm'
  if (responseFormat === 'opus')
    return 'ogg-24khz-16bit-mono-opus'
  return responseFormat
}

/**
 * Infers the MIME type for a Microsoft output format.
 *
 * Before:
 * - `"audio-24khz-48kbitrate-mono-mp3"`
 * - `"ogg-24khz-16bit-mono-opus"`
 * - `"riff-24khz-16bit-mono-pcm"`
 *
 * After:
 * - `"audio/mpeg"`
 * - `"audio/ogg"`
 * - `"audio/wav"`
 */
export function inferMicrosoftContentType(format: string): string {
  if (format.includes('mp3'))
    return 'audio/mpeg'
  if (format.includes('opus'))
    return 'audio/ogg'
  if (format.includes('pcm') || format.startsWith('riff'))
    return 'audio/wav'
  return 'application/octet-stream'
}

/**
 * Normalizes a speech speed multiplier into Microsoft SSML prosody rate.
 *
 * Before:
 * - `1.0`
 * - `1.2`
 * - `0.8`
 *
 * After:
 * - `""`
 * - `"+20%"`
 * - `"-20%"`
 */
export function microsoftSpeedToProsodyRate(speed: number | undefined): string {
  if (speed == null || speed === 1)
    return ''

  const delta = Math.round((speed - 1) * 100)
  if (delta === 0)
    return ''

  return delta > 0 ? `+${delta}%` : `${delta}%`
}

/**
 * Escapes plain text before inserting it into Microsoft SSML.
 *
 * Before:
 * - `"Tom & <Jerry>"`
 *
 * After:
 * - `"Tom &amp; &lt;Jerry&gt;"`
 */
export function escapeMicrosoftSsmlText(text: string): string {
  return text
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll('\'', '&apos;')
}

/**
 * Builds a Microsoft-compatible SSML envelope.
 *
 * Use when:
 * - Sending plain text through unspeech's Microsoft backend while preserving
 *   speed/prosody control.
 *
 * Expects:
 * - `text` is plain text. Already-built SSML should be passed through with
 *   `disableSsml`.
 *
 * Returns:
 * - A `<speak>` document that Microsoft accepts as `application/ssml+xml`.
 */
export function buildMicrosoftSsml(text: string, voice: string, speed: number | undefined): string {
  const safe = escapeMicrosoftSsmlText(text)
  const rate = microsoftSpeedToProsodyRate(speed)
  const inner = rate
    ? `<prosody rate='${rate}'>${safe}</prosody>`
    : safe

  return `<speak version='1.0' xml:lang='en-US'><voice name='${voice}'>${inner}</voice></speak>`
}

/**
 * [Microsoft / Azure AI](https://speech.microsoft.com/portal) provider for [UnSpeech](https://github.com/moeru-ai/unspeech)
 * only.
 *
 * [UnSpeech](https://github.com/moeru-ai/unspeech) is a open-source project that provides a
 * OpenAI-compatible audio & speech related API that can be used with various providers such
 * as ElevenLabs, Azure TTS, Google TTS, etc.
 *
 * @param apiKey - Microsoft / Azure AI subscription key
 * @param baseURL - UnSpeech Instance URL
 * @returns SpeechProviderWithExtraOptions
 */
export function createUnMicrosoft(apiKey: string, baseURL = 'http://localhost:5933/v1/') {
  const toUnSpeechOptions = (options: UnMicrosoftOptions): UnSpeechOptions => {
    const { deploymentId, region, sampleRate } = options

    const extraBody: Record<string, unknown> = {
      deploymentId,
      region,
      sampleRate,
    }

    if ('disableSsml' in options) {
      extraBody.disableSsml = options.disableSsml
    }
    else if ('lang' in options) {
      extraBody.lang = options.lang
      extraBody.gender = options.gender
      extraBody.voice = options.voice
    }

    return { extraBody: objCamelToSnake(extraBody) }
  }

  const speechProvider: SpeechProviderWithExtraOptions<
    /** @see Currently, cognitive services are on v1 */
    'microsoft/v1',
    UnMicrosoftOptions
  > = {
    speech: (model, options) => ({
      ...(options ? toUnSpeechOptions(options) : {}),
      apiKey,
      baseURL,
      model: `microsoft/${model}`,
    }),
  }

  const voiceProvider: VoiceProviderWithExtraOptions<
    UnMicrosoftOptions
  > = {
    voice: (options) => {
      if (baseURL.endsWith('v1/')) {
        baseURL = baseURL.slice(0, -3)
      }
      else if (baseURL.endsWith('v1')) {
        baseURL = baseURL.slice(0, -2)
      }

      return {
        query: `region=${options?.region}&provider=microsoft`,
        ...(options ? toUnSpeechOptions(options) : {}),
        apiKey,
        baseURL,
      }
    },
  }

  return merge(
    speechProvider,
    voiceProvider,
  )
}
