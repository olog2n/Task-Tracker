/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

// 🔧 CSS модули и обычные стили
declare module '*.css' {
  const css: string
  export default css
}

declare module '*.scss'
declare module '*.sass'
declare module '*.less'
