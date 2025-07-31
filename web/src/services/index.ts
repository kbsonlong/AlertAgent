/**
 * API 服务统一导出文件
 * 提供所有API服务的统一入口
 */

// 导出基础API服务
export { default as ApiService } from './api'
export type { ApiResponse, PaginatedResponse } from './api'

// 导出知识库服务
export { default as KnowledgeService } from './knowledgeService'
export type {
  Knowledge,
  CreateKnowledgeRequest,
  UpdateKnowledgeRequest,
  KnowledgeQueryParams
} from './knowledgeService'

// 导出告警服务
export { default as AlertService } from './alertService'
export type {
  Alert,
  CreateAlertRequest,
  UpdateAlertRequest,
  AlertQueryParams,
  AlertAnalysis
} from './alertService'

// 导出规则服务
export { default as RuleService } from './ruleService'
export type {
  Rule,
  CreateRuleRequest,
  UpdateRuleRequest,
  RuleQueryParams,
  RuleVersion,
  RuleAuditLog,
  RuleValidationResult,
  RuleTestResult
} from './ruleService'

// 导出数据源服务
export { default as ProviderService } from './providerService'
export type {
  Provider,
  CreateProviderRequest,
  UpdateProviderRequest,
  ProviderQueryParams,
  TestProviderRequest,
  ProviderTestResult,
  ProviderHealthCheck,
  MetricQueryRequest,
  MetricQueryResult
} from './providerService'

// 导出系统服务
export { default as SystemService } from './systemService'
export type {
  SystemConfig,
  SystemConfigGroup,
  UpdateSystemConfigRequest,
  BatchUpdateSystemConfigRequest,
  SystemInfo,
  SystemStats,
  SystemHealthCheck,
  SystemLogQueryParams,
  SystemLogEntry,
  SystemBackup,
  CreateBackupRequest,
  RestoreBackupRequest
} from './systemService'

// 导出用户服务
export { default as UserService } from './userService'
export type {
  User,
  Group,
  Permission,
  Role,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  CreateUserRequest,
  UpdateUserRequest,
  ChangePasswordRequest,
  ResetPasswordRequest,
  UserQueryParams,
  CreateGroupRequest,
  UpdateGroupRequest,
  GroupQueryParams,
  UserPreferences,
  UserActivity
} from './userService'

// 导出通知服务
export { default as NotificationService } from './notificationService'
export type {
  NotificationTemplate,
  CreateNotificationTemplateRequest,
  UpdateNotificationTemplateRequest,
  NotificationTemplateQueryParams,
  SendNotificationRequest,
  NotificationSendResult,
  NotificationHistory,
  NotificationStats,
  TestTemplateRequest,
  TestTemplateResult,
  NotificationChannel
} from './notificationService'

// 导入所有服务
import ApiService from './api'
import KnowledgeService from './knowledgeService'
import AlertService from './alertService'
import RuleService from './ruleService'
import ProviderService from './providerService'
import SystemService from './systemService'
import UserService from './userService'
import NotificationService from './notificationService'

// 导出所有服务的集合对象
export const Services = {
  Api: ApiService,
  Knowledge: KnowledgeService,
  Alert: AlertService,
  Rule: RuleService,
  Provider: ProviderService,
  System: SystemService,
  User: UserService,
  Notification: NotificationService
}

// 默认导出服务集合
export default Services