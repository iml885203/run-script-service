import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./tests/setup.ts'],
    // Exclude E2E tests from Vitest (they should only run with Playwright)
    exclude: [
      '**/node_modules/**',
      '**/dist/**',
      '**/tests/e2e/**',
      '**/*.e2e.{test,spec}.{js,ts}'
    ]
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  }
})
