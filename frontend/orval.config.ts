import { defineConfig } from 'orval';

export default defineConfig({
  rue: {
    output: './src/lib/api/generated/',
    input: {
      target: '../backend/docs/swagger.json',
    },
    hooks: {
      afterAllFilesWrite: 'prettier --write "src/lib/api/generated/**/*.{ts,tsx}"',
    },
    definitions: {
      query: {
        useInfinite: true,
      },
    },
    tags: {
      exclude: ['healthz'],
    },
    operations: {
      credentials: 'include',
    },
    // Override axios instance
    override: {
      axios: {
        output: './src/lib/api/client.ts',
        name: 'apiClient',
      },
    },
  },
});
