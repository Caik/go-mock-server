import { defineConfig } from 'vitest/config';
import tsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig({
  plugins: [tsconfigPaths()],
  test: {
    environment: 'jsdom',
    environmentOptions: {
      jsdom: {
        storageQuota: 10000000,
      },
    },
    globals: true,
    setupFiles: ['./vitest.setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov', 'html'],
      include: ['app/**/*.{ts,tsx}'],
      exclude: [
        'app/routes/**',
        'app/root.tsx',
        '**/*.d.ts',
      ],
    },
  },
});
