import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
	reporters: ['verbose', 'junit'],
	outputFile: './njs-unit-tests.xml',
    coverage: {
      reporter: ['text', 'json', 'html'],
    },
  },
});
