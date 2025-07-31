<template>
  <a-config-provider :locale="locale">
    <div id="app" :class="{ 'dark-theme': appStore.theme === 'dark' }">
      <router-view />
    </div>
  </a-config-provider>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { ConfigProvider } from 'ant-design-vue'
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import enUS from 'ant-design-vue/es/locale/en_US'
import { useAppStore } from '@/stores'
import 'dayjs/locale/zh-cn'
import dayjs from 'dayjs'

const AConfigProvider = ConfigProvider

const appStore = useAppStore()

// 根据当前语言设置获取对应的 locale
const locale = computed(() => {
  return appStore.locale === 'zh-CN' ? zhCN : enUS
})

// 设置 dayjs 语言
const setDayjsLocale = () => {
  if (appStore.locale === 'zh-CN') {
    dayjs.locale('zh-cn')
  } else {
    dayjs.locale('en')
  }
}

onMounted(() => {
  // 初始化应用设置
  appStore.initializeApp()
  
  // 设置 dayjs 语言
  setDayjsLocale()
  
  // 监听主题变化，更新 CSS 变量
  const updateTheme = () => {
    const root = document.documentElement
    if (appStore.theme === 'dark') {
      root.classList.add('dark')
    } else {
      root.classList.remove('dark')
    }
  }
  
  updateTheme()
  
  // 监听语言变化
  const unwatchLocale = appStore.$subscribe((mutation, state) => {
    if (mutation.events?.some((event: any) => event.key === 'locale')) {
      setDayjsLocale()
    }
    if (mutation.events?.some((event: any) => event.key === 'theme')) {
      updateTheme()
    }
  })
  
  // 组件卸载时清理监听
  return () => {
    unwatchLocale()
  }
})
</script>

<style>
/* 全局样式 */
#app {
  min-height: 100vh;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

/* 暗色主题 */
.dark-theme {
  background-color: #141414;
  color: #ffffff;
}

.dark-theme .ant-layout {
  background-color: #141414;
}

.dark-theme .ant-layout-sider {
  background-color: #001529;
}

.dark-theme .ant-layout-header {
  background-color: #001529;
}

.dark-theme .ant-layout-content {
  background-color: #141414;
}

.dark-theme .ant-card {
  background-color: #1f1f1f;
  border-color: #303030;
}

.dark-theme .ant-table {
  background-color: #1f1f1f;
}

.dark-theme .ant-table-thead > tr > th {
  background-color: #262626;
  border-color: #303030;
}

.dark-theme .ant-table-tbody > tr > td {
  border-color: #303030;
}

.dark-theme .ant-table-tbody > tr:hover > td {
  background-color: #262626;
}

.dark-theme .ant-form-item-label > label {
  color: #ffffff;
}

.dark-theme .ant-input {
  background-color: #1f1f1f;
  border-color: #303030;
  color: #ffffff;
}

.dark-theme .ant-input:hover {
  border-color: #1890ff;
}

.dark-theme .ant-input:focus {
  border-color: #1890ff;
  box-shadow: 0 0 0 2px rgba(24, 144, 255, 0.2);
}

.dark-theme .ant-select-selector {
  background-color: #1f1f1f !important;
  border-color: #303030 !important;
  color: #ffffff !important;
}

.dark-theme .ant-select-arrow {
  color: #ffffff;
}

.dark-theme .ant-btn {
  border-color: #303030;
}

.dark-theme .ant-btn-default {
  background-color: #1f1f1f;
  color: #ffffff;
}

.dark-theme .ant-btn-default:hover {
  background-color: #262626;
  border-color: #1890ff;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .ant-layout-sider {
    position: fixed !important;
    height: 100vh;
    z-index: 999;
  }
  
  .ant-layout-content {
    margin-left: 0 !important;
  }
}

/* 自定义滚动条 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

.dark-theme ::-webkit-scrollbar-track {
  background: #2f2f2f;
}

.dark-theme ::-webkit-scrollbar-thumb {
  background: #555;
}

.dark-theme ::-webkit-scrollbar-thumb:hover {
  background: #777;
}

/* 加载动画 */
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

/* 状态标签样式 */
.status-tag {
  font-weight: 500;
  border-radius: 4px;
  padding: 2px 8px;
  font-size: 12px;
}

.status-firing {
  background-color: #fff2e8;
  color: #fa541c;
  border: 1px solid #ffbb96;
}

.status-resolved {
  background-color: #f6ffed;
  color: #52c41a;
  border: 1px solid #b7eb8f;
}

.status-acknowledged {
  background-color: #e6f7ff;
  color: #1890ff;
  border: 1px solid #91d5ff;
}

.status-active {
  background-color: #f6ffed;
  color: #52c41a;
  border: 1px solid #b7eb8f;
}

.status-inactive {
  background-color: #fff2e8;
  color: #fa541c;
  border: 1px solid #ffbb96;
}

.dark-theme .status-firing {
  background-color: rgba(250, 84, 28, 0.1);
  border-color: rgba(250, 84, 28, 0.3);
}

.dark-theme .status-resolved {
  background-color: rgba(82, 196, 26, 0.1);
  border-color: rgba(82, 196, 26, 0.3);
}

.dark-theme .status-acknowledged {
  background-color: rgba(24, 144, 255, 0.1);
  border-color: rgba(24, 144, 255, 0.3);
}

.dark-theme .status-active {
  background-color: rgba(82, 196, 26, 0.1);
  border-color: rgba(82, 196, 26, 0.3);
}

.dark-theme .status-inactive {
  background-color: rgba(250, 84, 28, 0.1);
  border-color: rgba(250, 84, 28, 0.3);
}
</style>