import { defineConfig } from '@moeru/eslint-config'

export default defineConfig({
  masknet: false,
  preferArrow: false,
  perfectionist: false,
  sonarjs: false,
  typescript: true,
}, {
  ignores: [
    'cspell.config.yaml',
    'cspell.config.yml',
    '.vscode/settings.json',
    '.golangci.yml',
  ],
}, {
  rules: {
    'antfu/import-dedupe': 'error',
    // TODO: remove this
    'depend/ban-dependencies': 'warn',
    'import/order': 'off',
    'no-console': ['error', { allow: ['warn', 'error', 'info'] }],
    'perfectionist/sort-imports': [
      'error',
      {
        groups: [
          'type-builtin',
          'type-import',
          'type-internal',
          ['type-parent', 'type-sibling', 'type-index'],
          'default-value-builtin',
          'named-value-builtin',
          'value-builtin',
          'default-value-external',
          'named-value-external',
          'value-external',
          'default-value-internal',
          'named-value-internal',
          'value-internal',
          ['default-value-parent', 'default-value-sibling', 'default-value-index'],
          ['named-value-parent', 'named-value-sibling', 'named-value-index'],
          ['wildcard-value-parent', 'wildcard-value-sibling', 'wildcard-value-index'],
          ['value-parent', 'value-sibling', 'value-index'],
          'side-effect',
          'style',
        ],
        newlinesBetween: 'always',
      },
    ],
    'style/padding-line-between-statements': 'error',
    'yaml/plain-scalar': 'off',
  },
})
