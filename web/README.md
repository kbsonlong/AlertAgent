# AlertAgent Web 前端项目

## 项目简介

AlertAgent Web 是一个基于 Vue 3 + TypeScript + Ant Design Vue 的现代化告警管理系统前端项目。该项目采用 RBAC（基于角色的访问控制）模式，提供完整的用户权限管理功能。

## 技术栈

- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript
- **UI 组件库**: Ant Design Vue
- **路由**: Vue Router 4
- **状态管理**: Pinia
- **构建工具**: Vite
- **代码规范**: ESLint + Prettier

## 项目结构

```
src/
├── components/          # 公共组件
│   ├── MainLayout.vue   # 主布局组件
│   └── ...
├── views/               # 页面组件
│   ├── alerts/          # 告警管理
│   ├── rules/           # 规则管理
│   ├── knowledge/       # 知识库
│   ├── notifications/   # 渠道管理
│   ├── queue/           # 队列监控
│   ├── users/           # 用户管理 (RBAC)
│   │   ├── UserList.vue     # 用户列表
│   │   ├── RoleList.vue     # 角色管理
│   │   └── PermissionList.vue # 权限管理
│   ├── settings/        # 系统设置
│   └── auth/            # 认证相关
├── services/            # API 服务
│   ├── api.ts           # API 基础配置
│   └── userService.ts   # 用户相关 API
├── router/              # 路由配置
├── stores/              # Pinia 状态管理
└── types/               # TypeScript 类型定义
```

## RBAC 权限管理系统

### 功能概述

本项目实现了完整的 RBAC（Role-Based Access Control）权限管理系统，包含以下核心功能：

#### 1. 用户管理 (`/users`)

**功能特性：**
- 用户列表展示与分页
- 用户信息的增删改查
- 用户状态管理（启用/禁用）
- 批量操作（批量启用、禁用、删除）
- 用户搜索与筛选
- 用户数据导入导出
- 用户统计信息展示

**主要操作：**
- 新建用户：填写用户基本信息，分配角色
- 编辑用户：修改用户信息，调整角色权限
- 删除用户：支持单个删除和批量删除
- 状态切换：快速启用或禁用用户账户
- 角色分配：为用户分配一个或多个角色

#### 2. 角色管理 (`/roles`)

**功能特性：**
- 角色列表管理
- 角色权限配置
- 角色用户关联查看
- 系统角色与自定义角色区分
- 角色状态管理
- 角色数据导出

**主要操作：**
- 创建角色：定义角色名称、描述和权限范围
- 编辑角色：修改角色信息和权限配置
- 删除角色：删除自定义角色（系统角色不可删除）
- 权限配置：为角色分配具体的功能权限
- 用户查看：查看拥有该角色的所有用户

#### 3. 权限管理 (`/permissions`)

**功能特性：**
- 权限列表展示
- 权限分类管理
- 权限与角色关联查看
- 系统权限与自定义权限区分
- 权限状态管理
- 批量权限操作

**主要操作：**
- 创建权限：定义权限代码、名称、描述和资源
- 编辑权限：修改权限信息和分类
- 删除权限：删除自定义权限
- 关联查看：查看使用该权限的所有角色
- 批量管理：批量更新权限状态

### 数据模型

#### 用户 (User)
```typescript
interface User {
  id: number
  username: string
  email: string
  phone?: string
  real_name?: string
  avatar?: string
  status: 'active' | 'inactive'
  last_login_at?: string
  created_at: string
  updated_at: string
  roles: Role[]
}
```

#### 角色 (Role)
```typescript
interface Role {
  id: number
  name: string
  code: string
  description?: string
  type: 'system' | 'custom'
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
  permissions: Permission[]
}
```

#### 权限 (Permission)
```typescript
interface Permission {
  id: number
  name: string
  code: string
  description?: string
  resource: string
  action: string
  category: string
  type: 'system' | 'custom'
  status: 'active' | 'inactive'
  is_system: boolean
  created_at: string
  updated_at: string
}
```

### API 接口

#### 用户管理 API
- `GET /api/users` - 获取用户列表
- `POST /api/users` - 创建用户
- `PUT /api/users/:id` - 更新用户
- `DELETE /api/users/:id` - 删除用户
- `POST /api/users/batch-update` - 批量更新用户
- `POST /api/users/import` - 导入用户
- `GET /api/users/export` - 导出用户

#### 角色管理 API
- `GET /api/roles` - 获取角色列表
- `POST /api/roles` - 创建角色
- `PUT /api/roles/:id` - 更新角色
- `DELETE /api/roles/:id` - 删除角色
- `GET /api/roles/:id/users` - 获取角色用户
- `POST /api/roles/:id/users` - 分配用户到角色

#### 权限管理 API
- `GET /api/permissions` - 获取权限列表
- `POST /api/permissions` - 创建权限
- `PUT /api/permissions/:id` - 更新权限
- `DELETE /api/permissions/:id` - 删除权限
- `GET /api/permissions/:id/roles` - 获取权限关联角色

## 开发指南

### 环境要求

- Node.js >= 16.0.0
- npm >= 8.0.0

### 安装依赖

```bash
npm install
```

### 开发模式

```bash
npm run dev
```

访问 http://localhost:5173 查看应用

### 构建生产版本

```bash
npm run build
```

### 代码检查

```bash
npm run lint
```

### 类型检查

```bash
npm run type-check
```

## 使用说明

### 1. 登录系统

使用管理员账户登录系统，默认会跳转到告警管理页面。

### 2. 访问用户权限管理

在左侧导航栏中，点击「用户权限」菜单，可以看到三个子菜单：
- 用户管理：管理系统用户
- 角色管理：管理用户角色
- 权限管理：管理系统权限

### 3. 用户管理操作流程

1. **创建用户**：点击「新建用户」按钮，填写用户信息并分配角色
2. **编辑用户**：点击用户列表中的「编辑」按钮，修改用户信息
3. **分配角色**：在编辑用户时，可以为用户分配多个角色
4. **管理状态**：可以快速启用或禁用用户账户

### 4. 角色管理操作流程

1. **创建角色**：点击「新建角色」按钮，定义角色信息
2. **配置权限**：为角色分配具体的功能权限
3. **查看用户**：查看拥有该角色的所有用户
4. **编辑角色**：修改角色信息和权限配置

### 5. 权限管理操作流程

1. **查看权限**：浏览系统中所有可用权限
2. **创建权限**：定义新的功能权限
3. **管理分类**：按功能模块对权限进行分类
4. **查看关联**：查看使用该权限的所有角色

## 注意事项

1. **系统权限保护**：系统内置的角色和权限不能被删除，只能修改状态
2. **角色依赖检查**：删除角色前会检查是否有用户正在使用
3. **权限依赖检查**：删除权限前会检查是否有角色正在使用
4. **数据一致性**：所有关联操作都会保证数据的一致性
5. **操作日志**：重要操作会记录操作日志，便于审计

## 扩展开发

### 添加新的权限

1. 在权限管理页面创建新权限
2. 在对应的角色中分配该权限
3. 在前端代码中添加权限检查逻辑

### 自定义角色

1. 在角色管理页面创建自定义角色
2. 为角色分配所需权限
3. 将角色分配给相应用户

### 权限控制实现

在组件中使用权限检查：

```vue
<template>
  <a-button v-if="hasPermission('user:create')">
    新建用户
  </a-button>
</template>

<script setup>
import { usePermission } from '@/composables/usePermission'

const { hasPermission } = usePermission()
</script>
```

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交代码变更
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License