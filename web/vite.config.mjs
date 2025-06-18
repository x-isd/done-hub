// https://github.com/vitejs/vite/discussions/3448
import path from 'path';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import jsconfigPaths from 'vite-jsconfig-paths';

// ----------------------------------------------------------------------

export default defineConfig({
  plugins: [react(), jsconfigPaths()],
  // https://github.com/jpuri/react-draft-wysiwyg/issues/1317
  //   define: {
  //     global: 'window'
  //   },
  css: {
    preprocessorOptions: {
      scss: {
        // 使用现代 Sass API 解决 Legacy JS API 警告
        api: 'modern-compiler',
        // 静默弃用警告
        silenceDeprecations: ['legacy-js-api', 'import']
      }
    }
  },
  resolve: {
    alias: [
      {
        find: /^~(.+)/,
        replacement: path.join(process.cwd(), 'node_modules/$1')
      },
      {
        find: /^src(.+)/,
        replacement: path.join(process.cwd(), 'src/$1')
      }
    ]
  },
  server: {
    // this ensures that the browser opens upon server start
    open: true,
    // this sets a default port to 3000
    host: true,
    port: 3010,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:3000', // 设置代理的目标服务器
        changeOrigin: true
      }
    }
  },
  preview: {
    // this ensures that the browser opens upon preview start
    open: true,
    // this sets a default port to 3000
    port: 3010
  }
});
