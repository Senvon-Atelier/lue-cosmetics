import { defineConfig } from 'orval';

export default defineConfig({
  rue: {
    input: {
      target: '../backend/docs/swagger.json',
    },
    output: {
      target: './src/lib/api/generated/',
      client: 'react-query',
      httpClient: 'axios',
      override: {
        mutator: {
          path: './src/lib/api/client.ts',
          name: 'customInstance',
        },
      },
    },
    hooks: {
      afterAllFilesWrite: 'prettier --write "src/lib/api/generated/**/*.{ts,tsx}"',
    },
  },
});
