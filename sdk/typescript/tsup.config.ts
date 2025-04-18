import { join } from 'node:path'
import { defineConfig } from 'tsup'

export default defineConfig({
  dts: true,
  entry: [join('src', 'index.ts')],
  format: 'esm',
  tsconfig: 'tsconfig.lib.json',
})
