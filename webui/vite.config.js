import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ command, mode, ssrBuild }) => {
  const ret = {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      }
    },
  }
  ret.define = {
    "__API_URL__": JSON.stringify(process.env.VITE_API_URL || "http://localhost:3000") 
    // as the prof said, you should not modify this constant here - "__API_URL__": JSON.stringify("http://localhost:3000")
    // although so my deploy will work, i modified it to the 15th line
    // you have to put the constant back, if you use my code.
  }
  return ret;
})
