/// <reference types="vite/client" />

// eslint-disable-next-line @typescript-eslint/no-unused-vars
import * as React from 'react';

interface ImportMetaEnv {
  readonly VITE_API_URL?: string;
  // Add more env variables as needed
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

// React 18's types lack the native `inert` attribute. '' = inert, undefined = interactive.
declare module 'react' {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface HTMLAttributes<T> {
    inert?: '';
  }
}
