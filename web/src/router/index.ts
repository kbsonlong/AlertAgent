import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

// 路由配置
const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: {
      title: '登录',
      requiresAuth: false
    }
  },
  {
    path: '/',
    redirect: '/knowledge'
  },
  {
    path: '/',
    component: () => import('@/components/MainLayout.vue'),
    meta: {
      requiresAuth: true
    },
    children: [
      {
        path: 'knowledge',
        name: 'KnowledgeList',
        component: () => import('@/views/KnowledgeList.vue'),
        meta: {
          title: '知识库',
          icon: 'book'
        }
      },
      {
        path: 'queue-monitor',
        name: 'QueueMonitor',
        component: () => import('@/views/QueueMonitor.vue'),
        meta: {
          title: '队列监控',
          icon: 'monitor'
        }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/settings/Settings.vue'),
        meta: {
          title: '系统设置',
          icon: 'setting'
        }
      }
    ]
  }
]

// 创建路由实例
const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  if (to.meta?.title) {
    document.title = `${to.meta.title} - 告警管理系统`
  }
  
  // 检查是否需要认证
  const requiresAuth = to.matched.some(record => record.meta?.requiresAuth !== false)
  const token = localStorage.getItem('token')
  
  if (requiresAuth && !token) {
    // 需要认证但没有token，跳转到登录页
    next('/login')
  } else if (to.path === '/login' && token) {
    // 已登录用户访问登录页，跳转到首页
    next('/')
  } else {
    next()
  }
})

export default router