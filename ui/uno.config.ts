import { Preset, defineConfig, presetWind, transformerDirectives } from 'unocss'

import { FileSystemIconLoader } from '@iconify/utils/lib/loader/node-loaders'
import { presetIcons } from '@unocss/preset-icons'

export default defineConfig({
  presets: [
    presetWind(),
    presetIcons({
      collections: {
        local: FileSystemIconLoader('./src/assets/icons', (svg) =>
          svg.replace(/#fff/, 'currentColor')
        ),
      },
    }) as Preset,
  ],
  transformers: [transformerDirectives()],
  rules: [
    [
      'center',
      {
        display: 'flex',
        'justify-content': 'center',
        'align-items': 'center',
      },
    ],
  ],
})
