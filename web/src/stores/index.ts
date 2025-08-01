import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, MenuItem } from '@/types'
import { userApi } from '@/api/user'
import { message } from 'ant-design-vue'

// 用户状态管理
export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))
  
  const isLoggedIn = computed(() => !!token.value && !!user.value)
  
  const setUser = (userData: User) => {
    user.value = userData
  }
  
  const setToken = (tokenValue: string) => {
    token.value = tokenValue
    localStorage.setItem('token', tokenValue)
  }
  
  // 获取当前用户信息
  const fetchCurrentUser = async () => {
    try {
      if (!token.value) {
        throw new Error('未找到认证令牌')
      }
      const response = await userApi.getCurrentUser()
      user.value = response.data
      return response.data
    } catch (error) {
      console.error('获取用户信息失败:', error)
      // 如果获取用户信息失败，清除无效的token
      logout()
      throw error
    }
  }
  
  // 检查用户认证状态
  const checkAuth = async () => {
    if (!token.value) {
      throw new Error('未找到认证令牌')
    }
    
    // 如果已有用户信息，直接返回
    if (user.value) {
      return user.value
    }
    
    // 否则获取用户信息
    return await fetchCurrentUser()
  }
  
  const logout = () => {
    user.value = null
    token.value = null
    localStorage.removeItem('token')
  }
  
  return {
    user,
    token,
    isLoggedIn,
    setUser,
    setToken,
    fetchCurrentUser,
    checkAuth,
    logout
  }
})

// 应用状态管理
export const useAppStore = defineStore('app', () => {
  const collapsed = ref(false)
  const loading = ref(false)
  const theme = ref<'light' | 'dark'>('light')
  const locale = ref('zh-CN')
  
  // 菜单项
  const menuItems = ref<MenuItem[]>([
    {
      key: '/alerts',
      icon: 'AlertOutlined',
      label: '告警管理',
      path: '/alerts'
    },
    {
      key: '/rules',
      icon: 'SettingOutlined',
      label: '规则管理',
      path: '/rules'
    },
    {
      key: '/knowledge',
      icon: 'BookOutlined',
      label: '知识库',
      path: '/knowledge'
    },
    {
      key: '/providers',
      icon: 'DatabaseOutlined',
      label: '数据源',
      path: '/providers'
    },
    {
      key: '/notifications',
      icon: 'BellOutlined',
      label: '通知管理',
      children: [
        {
          key: '/notifications/groups',
          label: '通知组',
          path: '/notifications/groups'
        },
        {
          key: '/notifications/templates',
          label: '通知模板',
          path: '/notifications/templates'
        }
      ]
    },
    {
      key: '/settings',
      icon: 'ControlOutlined',
      label: '系统设置',
      path: '/settings'
    }
  ])
  
  const toggleCollapsed = () => {
    collapsed.value = !collapsed.value
  }
  
  const setLoading = (value: boolean) => {
    loading.value = value
  }
  
  const setTheme = (value: 'light' | 'dark') => {
    theme.value = value
    localStorage.setItem('theme', value)
  }
  
  const setLocale = (value: string) => {
    locale.value = value
    localStorage.setItem('locale', value)
  }
  
  // 初始化主题和语言
  const initializeApp = () => {
    const savedTheme = localStorage.getItem('theme') as 'light' | 'dark'
    if (savedTheme) {
      theme.value = savedTheme
    }
    
    const savedLocale = localStorage.getItem('locale')
    if (savedLocale) {
      locale.value = savedLocale
    }
  }
  
  return {
    collapsed,
    loading,
    theme,
    locale,
    menuItems,
    toggleCollapsed,
    setLoading,
    setTheme,
    setLocale,
    initializeApp
  }
})

// 告警状态管理
export const useAlertStore = defineStore('alert', () => {
  const alerts = ref([])
  const alertStats = ref({
    total: 0,
    firing: 0,
    resolved: 0,
    acknowledged: 0
  })
  
  const setAlerts = (data: any[]) => {
    alerts.value = data
  }
  
  const setAlertStats = (stats: any) => {
    alertStats.value = stats
  }
  
  const updateAlert = (id: number, data: any) => {
    const index = alerts.value.findIndex((alert: any) => alert.id === id)
    if (index !== -1) {
      alerts.value[index] = { ...alerts.value[index], ...data }
    }
  }
  
  return {
    alerts,
    alertStats,
    setAlerts,
    setAlertStats,
    updateAlert
  }
})

// 规则状态管理
export const useRuleStore = defineStore('rule', () => {
  const rules = ref([])
  const ruleGroups = ref([])
  
  const setRules = (data: any[]) => {
    rules.value = data
  }
  
  const setRuleGroups = (data: any[]) => {
    ruleGroups.value = data
  }
  
  const addRule = (rule: any) => {
    rules.value.unshift(rule)
  }
  
  const updateRule = (id: number, data: any) => {
    const index = rules.value.findIndex((rule: any) => rule.id === id)
    if (index !== -1) {
      rules.value[index] = { ...rules.value[index], ...data }
    }
  }
  
  const removeRule = (id: number) => {
    const index = rules.value.findIndex((rule: any) => rule.id === id)
    if (index !== -1) {
      rules.value.splice(index, 1)
    }
  }
  
  return {
    rules,
    ruleGroups,
    setRules,
    setRuleGroups,
    addRule,
    updateRule,
    removeRule
  }
})