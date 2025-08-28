import { GLOB_MARKDOWN, GLOB_MARKDOWN_CODE, GLOB_TESTS } from '@antfu/eslint-config'
import { defineConfig } from '@moeru/eslint-config'

export default defineConfig()
  .append({
    rules: {
      '@masknet/no-default-error': 'off',
      '@masknet/no-then': 'off',
      '@masknet/unicode-specific-set': 'off',
      'sonarjs/todo-tag': 'warn',
      'ts/strict-boolean-expressions': 'off',
    },
  })
  .append({
    files: [...GLOB_TESTS, GLOB_MARKDOWN_CODE, GLOB_MARKDOWN],
    rules: {
      '@masknet/no-top-level': 'off',
      '@masknet/unicode-specific-set': 'off',
      'sonarjs/unused-import': 'off',
    },
  })
  .append({
    ignores: [
      'cspell.config.yaml',
      'cspell.config.yml',
      '.vscode/settings.json',
      '.golangci.yml',
    ],
  })
