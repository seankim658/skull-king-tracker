/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
  readonly VITE_BACKEND_ASSET_BASE_URL: string;
  readonly VITE_APP_TITLE?: string;
  readonly VITE_SSE_BASE_URL: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
