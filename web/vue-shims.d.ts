// Vue.js 组件类型声明
declare module '*.vue' {
  const component: any
  export default component
}

// Vue 核心模块
declare module 'vue' {
  export function createApp(rootComponent: any): any
  export function ref<T>(value: T): any
  export function reactive<T extends object>(target: T): any
  export function computed<T>(getter: () => T): any
  export function watch(source: any, callback: any, options?: any): any
  export function onMounted(hook: () => void): void
  export function onUnmounted(hook: () => void): void
  export function nextTick(fn?: () => void): Promise<void>
  export const defineComponent: any
  export const h: any
  export const Fragment: any
  export const Teleport: any
  export const Suspense: any
  export const KeepAlive: any
  export const Transition: any
  export const TransitionGroup: any
  export const defineAsyncComponent: any
  export const defineCustomElement: any
  export const VueElement: any
  export const createSSRApp: any
  export const version: string
}

// Vue Router 模块
declare module 'vue-router' {
  export function createRouter(options: any): any
  export function createWebHistory(base?: string): any
  export function createWebHashHistory(base?: string): any
  export function createMemoryHistory(base?: string): any
  export function useRouter(): any
  export function useRoute(): any
  export function onBeforeRouteLeave(guard: any): void
  export function onBeforeRouteUpdate(guard: any): void
  export const START_LOCATION: any
  export const RouterLink: any
  export const RouterView: any
}

// Pinia 状态管理
declare module 'pinia' {
  export function createPinia(): any
  export function defineStore(id: string, setup: any): any
  export function defineStore(options: any): any
  export function storeToRefs(store: any): any
  export function acceptHMRUpdate(definition: any, hot: any): any
  export function mapActions(store: any, actions: any): any
  export function mapState(store: any, state: any): any
  export function mapStores(...stores: any[]): any
  export function mapWritableState(store: any, state: any): any
  export function setActivePinia(pinia: any): void
  export function getActivePinia(): any
}

// Ant Design Vue
declare module 'ant-design-vue' {
  const Antd: any
  export default Antd
  export const message: any
  export const notification: any
  export const Modal: any
  export const Button: any
  export const Input: any
  export const Form: any
  export const Table: any
  export const Select: any
  export const DatePicker: any
  export const Upload: any
  export const Spin: any
  export const Card: any
  export const Layout: any
  export const Menu: any
  export const Breadcrumb: any
  export const Dropdown: any
  export const Tabs: any
  export const Switch: any
  export const Radio: any
  export const Checkbox: any
  export const Pagination: any
  export const Steps: any
  export const Progress: any
  export const Tag: any
  export const Badge: any
  export const Avatar: any
  export const Tooltip: any
  export const Popover: any
  export const Popconfirm: any
  export const Alert: any
  export const Drawer: any
  export const Space: any
  export const Divider: any
  export const Typography: any
  export const Row: any
  export const Col: any
  export const Affix: any
  export const Anchor: any
  export const BackTop: any
  export const ConfigProvider: any
}

// Ant Design Vue CSS
declare module 'ant-design-vue/dist/reset.css' {
  const css: any
  export default css
}

// Ant Design Icons
declare module '@ant-design/icons-vue' {
  export const MenuFoldOutlined: any
  export const MenuUnfoldOutlined: any
  export const BookOutlined: any
  export const SettingOutlined: any
  export const UserOutlined: any
  export const DownOutlined: any
  export const LogoutOutlined: any
  export const PlusOutlined: any
  export const EditOutlined: any
  export const DeleteOutlined: any
  export const SearchOutlined: any
  export const ReloadOutlined: any
  export const UploadOutlined: any
  export const DownloadOutlined: any
  export const EyeOutlined: any
  export const CopyOutlined: any
  export const CheckOutlined: any
  export const CloseOutlined: any
  export const ExclamationCircleOutlined: any
  export const InfoCircleOutlined: any
  export const QuestionCircleOutlined: any
  export const WarningOutlined: any
  export const LoadingOutlined: any
}

// 本地模块声明
declare module '@/utils/datetime' {
  export function formatDate(date: Date | string): string
  export function formatDateTime(date: Date | string): string
  export function parseDate(dateString: string): Date
  export function isValidDate(date: any): boolean
  export function getRelativeTime(date: Date | string): string
}

declare module '@/services/*' {
  const service: any
  export default service
}

declare module '@/components/*' {
  const comp: any
  export default comp
}

declare module '@/views/*' {
  const view: any
  export default view
}

declare module '@/router' {
  const router: any
  export default router
}

declare module '@/store' {
  const store: any
  export default store
}

// CSS 模块
declare module '*.css' {
  const css: any
  export default css
}

declare module '*.scss' {
  const scss: any
  export default scss
}

declare module '*.sass' {
  const sass: any
  export default sass
}

declare module '*.less' {
  const less: any
  export default less
}

declare module '*.styl' {
  const stylus: any
  export default stylus
}

// 图片资源
declare module '*.png' {
  const pngSrc: string
  export default pngSrc
}

declare module '*.jpg' {
  const jpgSrc: string
  export default jpgSrc
}

declare module '*.jpeg' {
  const jpegSrc: string
  export default jpegSrc
}

declare module '*.gif' {
  const gifSrc: string
  export default gifSrc
}

declare module '*.svg' {
  const svgSrc: string
  export default svgSrc
}

declare module '*.webp' {
  const webpSrc: string
  export default webpSrc
}

// 其他资源
declare module '*.json' {
  const value: any
  export default value
}

declare module '*.txt' {
  const txtContent: string
  export default txtContent
}

// ES2015+ Polyfills 和全局声明
declare global {
  interface PromiseConstructor {
    new <T>(executor: (resolve: (value?: T | PromiseLike<T>) => void, reject: (reason?: any) => void) => void): Promise<T>
  }

  interface ObjectConstructor {
    assign<T, U>(target: T, source: U): T & U
    assign<T, U, V>(target: T, source1: U, source2: V): T & U & V
    assign<T, U, V, W>(target: T, source1: U, source2: V, source3: W): T & U & V & W
    assign(target: object, ...sources: any[]): any
    keys(o: object): string[]
    values<T>(o: { [s: string]: T } | ArrayLike<T>): T[]
    entries<T>(o: { [s: string]: T } | ArrayLike<T>): [string, T][]
  }

  interface Array<T> {
    includes(searchElement: T, fromIndex?: number): boolean
    find<S extends T>(predicate: (this: void, value: T, index: number, obj: T[]) => value is S, thisArg?: any): S | undefined
    find(predicate: (value: T, index: number, obj: T[]) => boolean, thisArg?: any): T | undefined
    findIndex(predicate: (value: T, index: number, obj: T[]) => boolean, thisArg?: any): number
  }

  interface String {
    includes(searchString: string, position?: number): boolean
    startsWith(searchString: string, position?: number): boolean
    endsWith(searchString: string, length?: number): boolean
    repeat(count: number): string
    padStart(targetLength: number, padString?: string): string
    padEnd(targetLength: number, padString?: string): string
  }

  // Node.js 全局变量
  const process: any
  const global: any
  const Buffer: any
  const __dirname: string
  const __filename: string
  const exports: any
  const module: any
  const require: any
}