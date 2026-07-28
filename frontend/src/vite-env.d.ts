/// <reference types="vite/client" />

import type { BackendAPI } from './types';

declare global {
  interface Window {
    go?: {
      main?: {
        App?: BackendAPI;
      };
    };
    runtime?: {
      EventsOn(eventName: string, callback: (payload: unknown) => void): () => void;
    };
  }
}

export {};
