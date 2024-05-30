import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vitejs.dev/config/
export default defineConfig(({mode}) => {
    const isProduction = mode === 'production';

    return {
        plugins: [svelte()],
        define: {
            'import.meta.env.VITE_BASE_URL': JSON.stringify( isProduction ? '': 'http://localhost:8189'),
        },
        build: {
            rollupOptions: {
                output: {
                    dir: "../static",
                    entryFileNames: `assets/[name].js`,
                    chunkFileNames: `assets/[name].js`,
                    assetFileNames: `assets/[name].[ext]`
                }
            }
        },
        server: {
            host: '0.0.0.0'
        }
    }
})