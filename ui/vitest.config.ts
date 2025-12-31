import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  test: {
    // Test environment
    environment: 'jsdom',

    // Setup files
    setupFiles: ['./src/setupTests.ts'],

    // Global test utilities
    globals: true,

    // Coverage configuration
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      exclude: [
        'node_modules/',
        'src/setupTests.ts',
        '**/*.d.ts',
        '**/*.config.*',
        '**/mockData',
        'src/main.tsx',
        'dist/',
        'src/utils/sampleData.ts', // Exclude large sample data file
        'src/utils/sampleDrifts.ts'
      ],
      // Target 60% coverage
      thresholds: {
        lines: 60,
        functions: 60,
        branches: 60,
        statements: 60
      }
    },

    // Test file patterns
    include: ['src/**/*.{test,spec}.{ts,tsx}'],

    // Exclude patterns
    exclude: [
      'node_modules',
      'dist',
      '.idea',
      '.git',
      '.cache'
    ],

    // Test timeout
    testTimeout: 10000,

    // Hooks timeout
    hookTimeout: 10000,

    // Watch mode
    watch: false,

    // Reporters
    reporters: ['verbose'],

    // Mock options
    mockReset: true,
    restoreMocks: true,
    clearMocks: true
  },

  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  }
});
