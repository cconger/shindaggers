import { defineConfig } from 'vite';
import solidPlugin from 'vite-plugin-solid';
// import devtools from 'solid-devtools/vite';

let upstream = 'http://localhost:8080';
if (process.env["SHINDAGGERS_UPSTREAM"]) {
  upstream = process.env["SHINDAGGERS_UPSTREAM"];
}

export default defineConfig({
  plugins: [
    /* 
    Uncomment the following line to enable solid-devtools.
    For more info see https://github.com/thetarnav/solid-devtools/tree/main/packages/extension#readme
    */
    // devtools(),
    solidPlugin(),
  ],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: upstream,
        changeOrigin: true,
      },
      '/oauth/redirect': {
        target: upstream,
        changeOrigin: true,
      },
      '/oauth/login': {
        target: upstream,
        changeOrigin: true,
      },
    },
  },
  build: {
    target: 'esnext',
  },
});
