<template>
  <a-layout class="main-layout">
    <!-- 侧边栏 -->
    <a-layout-sider
      v-model:collapsed="collapsed"
      :trigger="null"
      collapsible
      class="sidebar"
    >
      <div class="logo">
        <h3 v-if="!collapsed">告警管理系统</h3>
        <h3 v-else>AMS</h3>
      </div>
      
      <a-menu
        v-model:selectedKeys="selectedKeys"
        theme="dark"
        mode="inline"
        @click="handleMenuClick"
      >
        <a-menu-item key="alerts">
          <template #icon>
            <AlertOutlined />
          </template>
          告警管理
        </a-menu-item>
        
        <a-menu-item key="rules">
          <template #icon>
            <FileTextOutlined />
          </template>
          规则管理
        </a-menu-item>
        
        <a-menu-item key="knowledge">
          <template #icon>
            <BookOutlined />
          </template>
          知识库
        </a-menu-item>
        
        <a-menu-item key="notifications">
          <template #icon>
            <BellOutlined />
          </template>
          渠道管理
        </a-menu-item>
        
        <a-menu-item key="queue-monitor">
          <template #icon>
            <MonitorOutlined />
          </template>
          队列监控
        </a-menu-item>
        
        <a-sub-menu key="user-management">
          <template #icon>
            <TeamOutlined />
          </template>
          <template #title>用户权限</template>
          <a-menu-item key="users">
            <template #icon>
              <UserOutlined />
            </template>
            用户管理
          </a-menu-item>
          <a-menu-item key="roles">
            <template #icon>
              <TeamOutlined />
            </template>
            角色管理
          </a-menu-item>
          <a-menu-item key="permissions">
            <template #icon>
              <SafetyOutlined />
            </template>
            权限管理
          </a-menu-item>
        </a-sub-menu>
        
        <a-menu-item key="settings">
          <template #icon>
            <SettingOutlined />
          </template>
          系统设置
        </a-menu-item>
      </a-menu>
    </a-layout-sider>
    
    <!-- 主内容区 -->
    <a-layout>
      <!-- 顶部导航 -->
      <a-layout-header class="header">
        <div class="header-left">
          <a-button
            type="text"
            @click="collapsed = !collapsed"
            class="trigger"
          >
            <MenuUnfoldOutlined v-if="collapsed" />
            <MenuFoldOutlined v-else />
          </a-button>
          
          <a-breadcrumb class="breadcrumb">
            <a-breadcrumb-item>首页</a-breadcrumb-item>
            <a-breadcrumb-item v-if="currentTitle">{{ currentTitle }}</a-breadcrumb-item>
          </a-breadcrumb>
        </div>
        
        <div class="header-right">
          <a-dropdown>
            <a-button type="text" class="user-info">
              <UserOutlined />
              管理员
              <DownOutlined />
            </a-button>
            <template #overlay>
              <a-menu>
                <a-menu-item key="profile">
                  <UserOutlined />
                  个人信息
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item key="logout">
                  <LogoutOutlined />
                  退出登录
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>
      
      <!-- 内容区域 -->
      <a-layout-content class="content">
        <div class="content-wrapper">
          <router-view />
        </div>
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  AlertOutlined,
  BookOutlined,
  FileTextOutlined,
  BellOutlined,
  MonitorOutlined,
  SettingOutlined,
  UserOutlined,
  DownOutlined,
  LogoutOutlined,
  TeamOutlined,
  SafetyOutlined
} from '@ant-design/icons-vue'

const router = useRouter()
const route = useRoute()

// 侧边栏折叠状态
const collapsed = ref(false)

// 路由名称映射
const routeMap: Record<string, string> = {
  'alerts': 'AlertList',
  'rules': 'RuleList',
  'knowledge': 'KnowledgeList',
  'notifications': 'NotificationList',
  'queue-monitor': 'QueueMonitor',
  'users': 'UserList',
  'roles': 'RoleList',
  'permissions': 'PermissionList',
  'settings': 'Settings'
}

// 反向映射：从路由名称到菜单key
const reverseRouteMap: Record<string, string> = {
  'AlertList': 'alerts',
  'RuleList': 'rules',
  'KnowledgeList': 'knowledge',
  'NotificationList': 'notifications',
  'QueueMonitor': 'queue-monitor',
  'UserList': 'users',
  'RoleList': 'roles',
  'PermissionList': 'permissions',
  'Settings': 'settings'
}

// 获取当前菜单key
const getCurrentMenuKey = (routeName: string) => {
  return reverseRouteMap[routeName] || routeName
}

// 当前选中的菜单项
const selectedKeys = ref([getCurrentMenuKey(route.name as string)])

// 当前页面标题
const currentTitle = computed(() => {
  return route.meta?.title as string || ''
})

// 菜单点击处理
const handleMenuClick = ({ key }: { key: string }) => {
  selectedKeys.value = [key]
  
  const routeName = routeMap[key] || key
  router.push({ name: routeName })
}

// 监听路由变化更新选中状态
router.afterEach((to) => {
  selectedKeys.value = [getCurrentMenuKey(to.name as string)]
})
</script>

<style scoped>
.main-layout {
  min-height: 100vh;
}

.sidebar {
  position: fixed;
  height: 100vh;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 100;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.1);
  margin: 16px;
  border-radius: 6px;
}

.logo h3 {
  color: white;
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.header {
  background: #fff;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  margin-left: 200px;
  transition: margin-left 0.2s;
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
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.content {
  margin-left: 200px;
  transition: margin-left 0.2s;
  min-height: calc(100vh - 64px);
}

.content-wrapper {
  margin: 24px;
  padding: 24px;
  background: #fff;
  border-radius: 6px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  min-height: calc(100vh - 112px);
}

/* 响应式布局 */
@media (max-width: 768px) {
  .header,
  .content {
    margin-left: 80px;
  }
}

/* 折叠状态样式调整 */
.ant-layout-sider-collapsed + .ant-layout .header,
.ant-layout-sider-collapsed + .ant-layout .content {
  margin-left: 80px;
}
</style>