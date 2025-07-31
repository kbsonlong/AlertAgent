<template>
  <a-layout class="main-layout">
    <!-- 侧边栏 -->
    <a-layout-sider
      v-model:collapsed="appStore.collapsed"
      :trigger="null"
      collapsible
      :width="240"
      :collapsed-width="80"
      class="layout-sider"
    >
      <!-- Logo -->
      <div class="logo">
        <img src="/vite.svg" alt="Logo" class="logo-img" />
        <span v-if="!appStore.collapsed" class="logo-text">告警管理系统</span>
      </div>
      
      <!-- 菜单 -->
      <a-menu
        v-model:selectedKeys="selectedKeys"
        v-model:openKeys="openKeys"
        mode="inline"
        :theme="appStore.theme === 'dark' ? 'dark' : 'light'"
        class="layout-menu"
      >
        <template v-for="item in appStore.menuItems" :key="item.key">
          <a-sub-menu v-if="item.children" :key="item.key">
            <template #icon>
              <component :is="getIcon(item.icon)" />
            </template>
            <template #title>{{ item.label }}</template>
            <a-menu-item
              v-for="child in item.children"
              :key="child.key"
              @click="() => navigateTo(child.path!)"
            >
              {{ child.label }}
            </a-menu-item>
          </a-sub-menu>
          <a-menu-item v-else :key="item.key" @click="() => navigateTo(item.path!)">
            <template #icon>
              <component :is="getIcon(item.icon)" />
            </template>
            {{ item.label }}
          </a-menu-item>
        </template>
      </a-menu>
    </a-layout-sider>
    
    <a-layout class="layout-content">
      <!-- 头部 -->
      <a-layout-header class="layout-header">
        <div class="header-left">
          <a-button
            type="text"
            :icon="appStore.collapsed ? h(MenuUnfoldOutlined) : h(MenuFoldOutlined)"
            @click="appStore.toggleCollapsed"
            class="trigger"
          />
          
          <!-- 面包屑 -->
          <a-breadcrumb class="breadcrumb">
            <a-breadcrumb-item v-for="item in breadcrumbItems" :key="item.path">
              <router-link v-if="item.path && item.path !== $route.path" :to="item.path">
                {{ item.title }}
              </router-link>
              <span v-else>{{ item.title }}</span>
            </a-breadcrumb-item>
          </a-breadcrumb>
        </div>
        
        <div class="header-right">
          <!-- 主题切换 -->
          <a-tooltip :title="appStore.theme === 'dark' ? '切换到亮色主题' : '切换到暗色主题'">
            <a-button
              type="text"
              :icon="appStore.theme === 'dark' ? h(SunOutlined) : h(MoonOutlined)"
              @click="toggleTheme"
              class="header-action"
            />
          </a-tooltip>
          
          <!-- 全屏切换 -->
          <a-tooltip :title="isFullscreen ? '退出全屏' : '全屏'">
            <a-button
              type="text"
              :icon="isFullscreen ? h(FullscreenExitOutlined) : h(FullscreenOutlined)"
              @click="toggleFullscreen"
              class="header-action"
            />
          </a-tooltip>
          
          <!-- 刷新 -->
          <a-tooltip title="刷新页面">
            <a-button
              type="text"
              :icon="h(ReloadOutlined)"
              @click="refreshPage"
              class="header-action"
            />
          </a-tooltip>
          
          <!-- 用户菜单 -->
          <a-dropdown>
            <a-button type="text" class="user-info">
              <a-avatar size="small" :src="userStore.user?.avatar">
                {{ userStore.user?.name?.charAt(0) || 'U' }}
              </a-avatar>
              <span class="username">{{ userStore.user?.name || '用户' }}</span>
              <DownOutlined />
            </a-button>
            <template #overlay>
              <a-menu>
                <a-menu-item key="profile" @click="() => navigateTo('/profile')">
                  <UserOutlined />
                  个人资料
                </a-menu-item>
                <a-menu-item key="settings" @click="() => navigateTo('/settings')">
                  <SettingOutlined />
                  系统设置
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item key="logout" @click="handleLogout">
                  <LogoutOutlined />
                  退出登录
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>
      
      <!-- 内容区域 -->
      <a-layout-content class="main-content">
        <div class="content-wrapper">
          <router-view v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </div>
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { ref, computed, watch, h, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  Layout,
  Menu,
  Button,
  Breadcrumb,
  Dropdown,
  Avatar,
  Tooltip,
  message
} from 'ant-design-vue'
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  AlertOutlined,
  SettingOutlined,
  BookOutlined,
  DatabaseOutlined,
  BellOutlined,
  ControlOutlined,
  UserOutlined,
  LogoutOutlined,
  DownOutlined,
  SunOutlined,
  MoonOutlined,
  FullscreenOutlined,
  FullscreenExitOutlined,
  ReloadOutlined
} from '@ant-design/icons-vue'
import { useAppStore, useUserStore } from '@/stores'

const ALayout = Layout
const ALayoutSider = Layout.Sider
const ALayoutHeader = Layout.Header
const ALayoutContent = Layout.Content
const AMenu = Menu
const AMenuItem = Menu.Item
const ASubMenu = Menu.SubMenu
const AButton = Button
const ABreadcrumb = Breadcrumb
const ABreadcrumbItem = Breadcrumb.Item
const ADropdown = Dropdown
const AAvatar = Avatar
const ATooltip = Tooltip
const AMenuDivider = Menu.Divider

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const userStore = useUserStore()

// 菜单状态
const selectedKeys = ref<string[]>([])
const openKeys = ref<string[]>([])
const isFullscreen = ref(false)

// 图标映射
const iconMap = {
  AlertOutlined,
  SettingOutlined,
  BookOutlined,
  DatabaseOutlined,
  BellOutlined,
  ControlOutlined
}

// 获取图标组件
const getIcon = (iconName?: string) => {
  if (!iconName) return null
  return iconMap[iconName as keyof typeof iconMap] || null
}

// 面包屑数据
const breadcrumbItems = computed(() => {
  const pathSegments = route.path.split('/').filter(Boolean)
  const items = [{ title: '首页', path: '/' }]
  
  let currentPath = ''
  for (const segment of pathSegments) {
    currentPath += `/${segment}`
    
    // 根据路径生成面包屑标题
    let title = segment
    switch (segment) {
      case 'alerts':
        title = '告警管理'
        break
      case 'rules':
        title = '规则管理'
        break
      case 'knowledge':
        title = '知识库'
        break
      case 'providers':
        title = '数据源'
        break
      case 'notifications':
        title = '通知管理'
        break
      case 'groups':
        title = '通知组'
        break
      case 'templates':
        title = '通知模板'
        break
      case 'settings':
        title = '系统设置'
        break
    }
    
    items.push({ title, path: currentPath })
  }
  
  return items
})

// 导航到指定路径
const navigateTo = (path: string) => {
  router.push(path)
}

// 切换主题
const toggleTheme = () => {
  const newTheme = appStore.theme === 'dark' ? 'light' : 'dark'
  appStore.setTheme(newTheme)
  message.success(`已切换到${newTheme === 'dark' ? '暗色' : '亮色'}主题`)
}

// 全屏切换
const toggleFullscreen = () => {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen()
    isFullscreen.value = true
  } else {
    document.exitFullscreen()
    isFullscreen.value = false
  }
}

// 刷新页面
const refreshPage = () => {
  window.location.reload()
}

// 退出登录
const handleLogout = () => {
  userStore.logout()
  router.push('/login')
  message.success('已退出登录')
}

// 监听路由变化，更新菜单选中状态
watch(
  () => route.path,
  (newPath) => {
    selectedKeys.value = [newPath]
    
    // 自动展开包含当前路径的子菜单
    const pathSegments = newPath.split('/').filter(Boolean)
    if (pathSegments.length > 1) {
      const parentPath = `/${pathSegments[0]}`
      if (!openKeys.value.includes(parentPath)) {
        openKeys.value.push(parentPath)
      }
    }
  },
  { immediate: true }
)

// 监听全屏状态变化
const handleFullscreenChange = () => {
  isFullscreen.value = !!document.fullscreenElement
}

onMounted(() => {
  document.addEventListener('fullscreenchange', handleFullscreenChange)
})

onUnmounted(() => {
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
})
</script>

<style scoped>
.main-layout {
  min-height: 100vh;
}

.layout-sider {
  position: fixed;
  height: 100vh;
  left: 0;
  top: 0;
  z-index: 100;
  overflow: auto;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 16px;
  border-bottom: 1px solid #f0f0f0;
}

.logo-img {
  width: 32px;
  height: 32px;
  margin-right: 8px;
}

.logo-text {
  font-size: 16px;
  font-weight: 600;
  color: #1890ff;
  white-space: nowrap;
}

.layout-menu {
  border-right: none;
}

.layout-content {
  margin-left: 240px;
  transition: margin-left 0.2s;
}

.layout-content.collapsed {
  margin-left: 80px;
}

.layout-header {
  background: #fff;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #f0f0f0;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
}

.header-left {
  display: flex;
  align-items: center;
}

.trigger {
  font-size: 18px;
  margin-right: 24px;
}

.breadcrumb {
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-action {
  font-size: 16px;
  color: #666;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  height: 40px;
}

.username {
  margin-left: 8px;
  margin-right: 4px;
}

.main-content {
  margin: 24px;
  padding: 24px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  min-height: calc(100vh - 112px);
}

.content-wrapper {
  min-height: 100%;
}

/* 页面切换动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .layout-sider {
    transform: translateX(-100%);
    transition: transform 0.3s;
  }
  
  .layout-sider.ant-layout-sider-collapsed {
    transform: translateX(0);
  }
  
  .layout-content {
    margin-left: 0;
  }
  
  .main-content {
    margin: 16px;
    padding: 16px;
  }
  
  .breadcrumb {
    display: none;
  }
}

/* 暗色主题适配 */
:deep(.ant-layout-sider-dark) {
  background: #001529;
}

:deep(.ant-layout-sider-dark .logo) {
  border-bottom-color: #303030;
}

:deep(.ant-layout-sider-dark .logo-text) {
  color: #fff;
}

:deep(.dark-theme .layout-header) {
  background: #001529;
  border-bottom-color: #303030;
}

:deep(.dark-theme .main-content) {
  background: #1f1f1f;
}
</style>