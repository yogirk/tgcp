// @ts-check
import { defineConfig } from 'astro/config';
import tailwindcss from '@tailwindcss/vite';

// https://astro.build/config
export default defineConfig({
  site: 'https://yogirk.github.io',
  base: '/tgcp',
  vite: {
    plugins: [tailwindcss()],
  },
});
