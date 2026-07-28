import { createReadStream, statSync } from 'node:fs';
import { homedir } from 'node:os';
import { extname, join, resolve, sep } from 'node:path';

import { defineConfig, type Plugin } from 'vite';
import react from '@vitejs/plugin-react';

function localAssetCache(): Plugin {
  const configRoot = process.env.APPDATA ?? join(homedir(), 'AppData', 'Roaming');
  const cacheRoot = resolve(configRoot, 'WaveArchive', 'assets');
  const contentTypes: Record<string, string> = {
    '.gif': 'image/gif',
    '.jpeg': 'image/jpeg',
    '.jpg': 'image/jpeg',
    '.png': 'image/png',
    '.svg': 'image/svg+xml',
    '.webp': 'image/webp',
  };

  return {
    name: 'wavearchive-local-asset-cache',
    configureServer(server) {
      server.middlewares.use('/cache', (request, response, next) => {
        let pathname: string;
        try {
          pathname = decodeURIComponent(
            new URL(request.url ?? '/', 'http://wavearchive.local').pathname
          );
        } catch {
          response.statusCode = 400;
          response.end('Invalid asset path');
          return;
        }

        const assetPath = resolve(cacheRoot, pathname.replace(/^[/\\]+/, ''));
        if (assetPath !== cacheRoot && !assetPath.startsWith(`${cacheRoot}${sep}`)) {
          response.statusCode = 403;
          response.end('Asset path is outside the cache');
          return;
        }

        try {
          if (!statSync(assetPath).isFile()) {
            next();
            return;
          }
        } catch {
          next();
          return;
        }

        response.setHeader(
          'Content-Type',
          contentTypes[extname(assetPath).toLowerCase()] ?? 'application/octet-stream'
        );
        response.setHeader('Cache-Control', 'no-cache');
        createReadStream(assetPath).on('error', next).pipe(response);
      });
    },
  };
}

export default defineConfig({
  plugins: [react(), localAssetCache()],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
});
