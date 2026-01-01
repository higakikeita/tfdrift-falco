/**
 * Lighthouse CI Configuration
 * Performance and quality audits for the TFDrift Falco UI
 */

module.exports = {
  ci: {
    collect: {
      // Start the dev server before collecting
      startServerCommand: 'npm run preview',
      url: ['http://localhost:4173'],
      numberOfRuns: 3,
    },
    assert: {
      preset: 'lighthouse:recommended',
      assertions: {
        // Performance thresholds
        'categories:performance': ['error', { minScore: 0.8 }],

        // Accessibility thresholds (WCAG compliance)
        'categories:accessibility': ['error', { minScore: 0.9 }],

        // Best practices
        'categories:best-practices': ['error', { minScore: 0.9 }],

        // SEO
        'categories:seo': ['error', { minScore: 0.8 }],

        // Core Web Vitals
        'first-contentful-paint': ['warn', { maxNumericValue: 2000 }],
        'largest-contentful-paint': ['warn', { maxNumericValue: 2500 }],
        'cumulative-layout-shift': ['warn', { maxNumericValue: 0.1 }],
        'total-blocking-time': ['warn', { maxNumericValue: 300 }],

        // Resource sizes
        'total-byte-weight': ['warn', { maxNumericValue: 1000000 }], // 1MB
        'dom-size': ['warn', { maxNumericValue: 1500 }],

        // Specific audits
        'uses-responsive-images': 'off',
        'offscreen-images': 'warn',
        'uses-optimized-images': 'warn',
        'uses-webp-images': 'warn',
        'modern-image-formats': 'warn',
        'uses-text-compression': 'warn',
        'unused-javascript': 'warn',
        'unused-css-rules': 'warn',
      },
    },
    upload: {
      target: 'temporary-public-storage',
    },
  },
};
