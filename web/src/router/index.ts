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
    component: () => import('@/components/MainLayout.vue'),
    redirect: '/alerts',
    meta: {
      requiresAuth: true
    },
    children: [
      {
        path: 'alerts',
        name: 'AlertList',
        component: () => import('@/views/AlertList.vue'),
        meta: {
          title: '告警管理',
          icon: 'alert'
        }
      },
      {
        path: 'rules',
        name: 'RuleList',
        component: () => import('@/views/RuleList.vue'),
        meta: {
          title: '规则管理',
          icon: 'rule'
        }
      },
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
        path: 'notifications',
        name: 'NotificationList',
        component: () => import('@/views/NotificationList.vue'),
        meta: {
          title: '渠道管理',
          icon: 'notification'
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
        path: 'users',
        name: 'UserList',
        component: () => import('@/views/UserList.vue'),
        meta: {
          title: '用户管理',
          icon: 'user'
        }
      },
      {
        path: 'roles',
        name: 'RoleList',
        component: () => import('@/views/RoleList.vue'),
        meta: {
          title: '角色管理',
          icon: 'team'
        }
      },
      {
        path: 'permissions',
        name: 'PermissionList',
        component: () => import('@/views/PermissionList.vue'),
        meta: {
          title: '权限管理',
          icon: 'safety'
        }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/Settings.vue'),
        meta: {
          title: '系统设置',
          icon: 'setting'
        }
      },
      {
        path: 'config-sync',
        name: 'ConfigSyncMonitor',
        component: () => import('@/views/ConfigSyncMonitor.vue'),
        meta: {
          title: '配置同步监控',
          icon: 'sync'
        }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/Profile.vue'),
        meta: {
          title: '个人资料',
          icon: 'user'
        }
      },
      {
        path: 'test-alert-detail',
        name: 'TestAlertDetail',
        component: () => import('@/views/TestAlertDetail.vue'),
        meta: {
          title: '告警详情测试',
          icon: 'experiment'
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
    document.title = `${to.meta.title} - AlertAgent`
  }
  
  // 检查是否需要认证
  const requiresAuth = to.matched.some(record => record.meta?.requiresAuth === true)
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