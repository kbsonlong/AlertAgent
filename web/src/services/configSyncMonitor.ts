import axios from 'axios';

const API_BASE = '/api/v1/config/sync';

// 同步指标接口
export interface SyncMetrics {
  cluster_id: string;
  config_type: string;
  last_sync_time?: string;
  sync_status: string;
  sync_delay_seconds: number;
  success_rate: number;
  failure_count: number;
  average_duration_ms: number;
  error_message?: string;
  is_healthy: boolean;
  config_hash: string;
  config_size: number;
}

export interface SyncSummary {
  total_clusters: number;
  healthy_clusters: number;
  unhealthy_clusters: number;
  config_types: string[];
  cluster_metrics: Record<string, SyncMetrics[]>;
  overall_health: string;
  last_update_time: string;
}

export interface SyncDelayPoint {
  timestamp: string;
  duration_ms: number;
  success_rate: number;
  sample_count: number;
}

export interface FailureRatePoint {
  timestamp: string;
  failure_rate: number;
  total_count: number;
  failure_count: number;
}

// 异常接口
export interface SyncException {
  id: string;
  cluster_id: string;
  config_type: string;
  exception_type: string;
  error_message: string;
  severity: string;
  status: string;
  first_occurred: string;
  last_occurred: string;
  occurrence_count: number;
  auto_retry_count: number;
  max_auto_retry: number;
  next_retry_at?: string;
  resolved_at?: string;
  resolved_by?: string;
  created_at: string;
  updated_at: string;
}

export interface ExceptionAnalysis {
  exception_id: string;
  root_cause: string;
  possible_causes: string[];
  suggested_actions: string[];
  related_exceptions: string[];
  confidence: number;
  analysis_time: string;
  metadata: Record<string, any>;
}

// 版本管理接口
export interface ConfigVersion {
  id: string;
  cluster_id: string;
  config_type: string;
  version: string;
  config_hash: string;
  config_content: string;
  config_size: number;
  description: string;
  created_by: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ConfigDiff {
  cluster_id: string;
  config_type: string;
  from_version: string;
  to_version: string;
  diff_type: string;
  changes: DiffLine[];
  summary: DiffSummary;
  generated_at: string;
}

export interface DiffLine {
  line_number: number;
  type: string;
  old_content?: string;
  new_content?: string;
}

export interface DiffSummary {
  added_lines: number;
  removed_lines: number;
  modified_lines: number;
  total_changes: number;
}

export interface ConsistencyCheck {
  cluster_id: string;
  config_type: string;
  expected_hash: string;
  actual_hash: string;
  is_consistent: boolean;
  inconsistencies?: InconsistencyDetail[];
  check_time: string;
  recommendations?: string[];
}

export interface InconsistencyDetail {
  type: string;
  description: string;
  severity: string;
  impact: string;
}

// 配置同步监控服务
export class ConfigSyncMonitorService {
  // 获取同步指标
  static async getSyncMetrics(): Promise<SyncSummary> {
    const response = await axios.get(`${API_BASE}/metrics`);
    return response.data.data.metrics;
  }

  // 获取同步延迟指标
  static async getSyncDelayMetrics(
    clusterId?: string,
    configType?: string,
    hours: number = 24
  ): Promise<SyncDelayPoint[]> {
    const params = new URLSearchParams();
    if (clusterId) params.append('cluster_id', clusterId);
    if (configType) params.append('config_type', configType);
    params.append('hours', hours.toString());

    const response = await axios.get(`${API_BASE}/metrics/delay?${params}`);
    return response.data.data.delay_metrics;
  }

  // 获取失败率指标
  static async getFailureRateMetrics(
    clusterId?: string,
    configType?: string,
    hours: number = 24
  ): Promise<FailureRatePoint[]> {
    const params = new URLSearchParams();
    if (clusterId) params.append('cluster_id', clusterId);
    if (configType) params.append('config_type', configType);
    params.append('hours', hours.toString());

    const response = await axios.get(`${API_BASE}/metrics/failure-rate?${params}`);
    return response.data.data.failure_rate_metrics;
  }

  // 记录同步历史
  static async recordSyncHistory(data: {
    cluster_id: string;
    config_type: string;
    config_hash: string;
    config_size?: number;
    sync_status: string;
    sync_duration_ms?: number;
    error_message?: string;
  }): Promise<void> {
    await axios.post(`${API_BASE}/history`, data);
  }

  // 清理旧历史记录
  static async cleanupOldHistory(retentionDays: number = 30): Promise<void> {
    await axios.delete(`${API_BASE}/history/cleanup?retention_days=${retentionDays}`);
  }

  // 检测同步异常
  static async detectExceptions(): Promise<void> {
    await axios.post(`${API_BASE}/exceptions/detect`);
  }

  // 获取活跃异常
  static async getActiveExceptions(
    clusterId?: string,
    configType?: string
  ): Promise<SyncException[]> {
    const params = new URLSearchParams();
    if (clusterId) params.append('cluster_id', clusterId);
    if (configType) params.append('config_type', configType);

    const response = await axios.get(`${API_BASE}/exceptions?${params}`);
    return response.data.data.exceptions;
  }

  // 分析异常
  static async analyzeException(exceptionId: string): Promise<ExceptionAnalysis> {
    const response = await axios.get(`${API_BASE}/exceptions/${exceptionId}/analysis`);
    return response.data.data.analysis;
  }

  // 解决异常
  static async resolveException(
    exceptionId: string,
    resolvedBy: string,
    resolution?: string
  ): Promise<void> {
    await axios.post(`${API_BASE}/exceptions/${exceptionId}/resolve`, {
      resolved_by: resolvedBy,
      resolution,
    });
  }

  // 触发手动重试
  static async triggerManualRetry(
    exceptionId: string,
    retryBy: string,
    force: boolean = false
  ): Promise<void> {
    await axios.post(`${API_BASE}/exceptions/${exceptionId}/retry`, {
      retry_by: retryBy,
      force,
    });
  }

  // 获取异常统计
  static async getExceptionStatistics(
    clusterId?: string,
    configType?: string
  ): Promise<any> {
    const params = new URLSearchParams();
    if (clusterId) params.append('cluster_id', clusterId);
    if (configType) params.append('config_type', configType);

    const response = await axios.get(`${API_BASE}/exceptions/statistics?${params}`);
    return response.data.data.statistics;
  }
}

// 配置版本管理服务
export class ConfigVersionService {
  // 创建版本
  static async createVersion(data: {
    cluster_id: string;
    config_type: string;
    description?: string;
    created_by: string;
  }): Promise<ConfigVersion> {
    const response = await axios.post('/api/v1/config/versions', data);
    return response.data.data.version;
  }

  // 获取版本列表
  static async getVersions(
    clusterId: string,
    configType: string,
    limit: number = 50
  ): Promise<ConfigVersion[]> {
    const response = await axios.get('/api/v1/config/versions', {
      params: { cluster_id: clusterId, config_type: configType, limit },
    });
    return response.data.data.versions;
  }

  // 获取指定版本
  static async getVersion(versionId: string): Promise<ConfigVersion> {
    const response = await axios.get(`/api/v1/config/versions/${versionId}`);
    return response.data.data.version;
  }

  // 比较版本
  static async compareVersions(fromVersionId: string, toVersionId: string): Promise<ConfigDiff> {
    const response = await axios.get('/api/v1/config/versions/compare', {
      params: { from: fromVersionId, to: toVersionId },
    });
    return response.data.data.diff;
  }

  // 回滚版本
  static async rollbackToVersion(versionId: string, rollbackBy: string): Promise<void> {
    await axios.post(`/api/v1/config/versions/${versionId}/rollback`, {
      rollback_by: rollbackBy,
    });
  }

  // 检查一致性
  static async checkConsistency(
    clusterId: string,
    configType: string
  ): Promise<ConsistencyCheck> {
    const response = await axios.get('/api/v1/config/versions/consistency', {
      params: { cluster_id: clusterId, config_type: configType },
    });
    return response.data.data.consistency_check;
  }

  // 获取活跃版本
  static async getActiveVersion(
    clusterId: string,
    configType: string
  ): Promise<ConfigVersion> {
    const response = await axios.get('/api/v1/config/versions/active', {
      params: { cluster_id: clusterId, config_type: configType },
    });
    return response.data.data.active_version;
  }

  // 删除版本
  static async deleteVersion(versionId: string, deletedBy: string): Promise<void> {
    await axios.delete(`/api/v1/config/versions/${versionId}`, {
      data: { deleted_by: deletedBy },
    });
  }

  // 清理旧版本
  static async cleanupOldVersions(
    clusterId: string,
    configType: string,
    keepCount: number = 10
  ): Promise<void> {
    await axios.post('/api/v1/config/versions/cleanup', {
      cluster_id: clusterId,
      config_type: configType,
      keep_count: keepCount,
    });
  }
}