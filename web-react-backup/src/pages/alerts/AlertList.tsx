import React, { useEffect, useState } from 'react';
import { Table, Button, Space, Modal, Tag, message, Descriptions } from 'antd';
import { Alert, getAlerts, getAlert, updateAlertStatus, analyzeAlert, asyncAnalyzeAlert, getAnalysisResult, convertToKnowledge } from '../../services/alert';
import { getSystemConfig, SystemConfig } from '../../services/config';
import { formatDateTime } from '../../utils/datetime';
import ReactMarkdown from 'react-markdown';
import { DownOutlined, UpOutlined, EyeOutlined } from '@ant-design/icons';
import AlertDetailModal from '../../components/AlertDetailModal';

const AlertList: React.FC = () => {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(false);
  const [analysisModalVisible, setAnalysisModalVisible] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [currentAlert, setCurrentAlert] = useState<Alert | null>(null);
  const [selectedAlertId, setSelectedAlertId] = useState<number | null>(null);
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [showThinkContent, setShowThinkContent] = useState(false);
  const [systemConfig, setSystemConfig] = useState<SystemConfig | null>(null);

  const fetchAlerts = async () => {
    setLoading(true);
    try {
      const response = await getAlerts();
      setAlerts(response);
    } catch {
      message.error('获取告警列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchSystemConfig = async () => {
    try {
      const config = await getSystemConfig();
      setSystemConfig(config);

    } catch (error) {
      console.error('获取系统配置失败:', error);
      message.error('获取系统配置失败');
    }
  };

  useEffect(() => {
    fetchAlerts();
    fetchSystemConfig();
  }, []);

  const handleAcknowledge = async (record: Alert) => {
    try {
      await updateAlertStatus(record.id, 'acknowledged');
      message.success('已确认告警');
      fetchAlerts();
    } catch {
      message.error('确认告警失败');
    }
  };

  const handleResolve = async (record: Alert) => {
    try {
      await updateAlertStatus(record.id, 'resolved');
      message.success('已解决告警');
      fetchAlerts();
    } catch {
      message.error('解决告警失败');
    }
  };

  const handleAnalyze = async (record: Alert) => {
    setCurrentAlert(record);
    setAnalysisModalVisible(true);
    setIsAnalyzing(true);

    try {
      const task = await asyncAnalyzeAlert(record.id);
      message.success('分析任务已提交，正在处理中...');
      
      // 轮询分析结果
      const checkResult = async () => {
        try {
          const result = await getAnalysisResult(task.task_id);
          
          if (result.status === 'completed') {
            // 分析完成，停止轮询
            setCurrentAlert(prev => prev ? { ...prev, analysis: result.result || result.message || '分析已完成' } : null);
            setIsAnalyzing(false);
            message.success('分析完成');
          } else if (result.status === 'failed') {
            message.error(result.error || '分析失败');
            setIsAnalyzing(false);
          } else if (result.status === 'processing') {
            // 继续轮询
            setTimeout(checkResult, 3000);
          } else {
            // 未知状态，继续轮询
            setTimeout(checkResult, 3000);
          }
        } catch (error) {
          console.error('获取分析结果失败:', error);
          setTimeout(checkResult, 3000);
        }
      };

      // 延迟开始轮询，给后端处理时间
      setTimeout(checkResult, 2000);
    } catch (error) {
      console.error('启动分析失败:', error);
      message.error('启动分析失败');
      setIsAnalyzing(false);
    }
  };

  const handleViewDetail = (alertId: number) => {
    setSelectedAlertId(alertId);
    setDetailModalVisible(true);
  };

  const handleCloseDetailModal = () => {
    setDetailModalVisible(false);
    setSelectedAlertId(null);
  };

  const handleConvertToKnowledge = async (record: Alert) => {
    try {
      await convertToKnowledge(record.id);
      message.success('已成功转换为知识库记录');
    } catch {
      message.error('转换知识库失败');
    }
  };

  const getSeverityColor = (severity: string | undefined): string => {
    if (!severity) return 'default';
    switch (severity.toLowerCase()) {
      case 'critical':
        return 'red';
      case 'high':
        return 'orange';
      case 'medium':
        return 'yellow';
      case 'low':
        return 'green';
      default:
        return 'default';
    }
  };

  const getStatusColor = (status: string | undefined): string => {
    if (!status) return 'default';
    switch (status.toLowerCase()) {
      case 'new':
        return 'blue';
      case 'acknowledged':
        return 'orange';
      case 'resolved':
        return 'green';
      default:
        return 'default';
    }
  };

  const columns = [
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
    },
    {
      title: '内容',
      dataIndex: 'content',
      key: 'content',
      ellipsis: true,
    },
    {
      title: '严重程度',
      dataIndex: 'severity',
      key: 'severity',
      render: (severity: string | undefined) => {
        const color = getSeverityColor(severity);
        return <Tag color={color}>{severity?.toUpperCase() || '未知'}</Tag>;
      },
    },
    {
      title: '来源',
      dataIndex: 'source',
      key: 'source',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string | undefined) => {
        const color = getStatusColor(status);
        return <Tag color={color}>{status?.toUpperCase() || '未知'}</Tag>;
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (text: string | undefined) => formatDateTime(text || ''),
      sorter: (a: Alert, b: Alert) => {
        if (!a.created_at || !b.created_at) return 0;
        return new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
      },
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: Alert) => (
        <Space size="middle">
          <Button 
            type="link" 
            icon={<EyeOutlined />}
            onClick={() => handleViewDetail(record.id)}
          >
            详情
          </Button>
          {record.status === 'new' && (
            <Button type="link" onClick={() => handleAcknowledge(record)}>
              确认
            </Button>
          )}
          {record.status !== 'resolved' && (
            <Button type="link" onClick={() => handleResolve(record)}>
              解决
            </Button>
          )}
          {systemConfig?.ollama_enabled === true && (
            <Button type="link" onClick={() => handleAnalyze(record)}>
              分析
            </Button>
          )}
          <Button type="link" onClick={() => handleConvertToKnowledge(record)}>
            转为知识库
          </Button>
        </Space>
      ),
    },
  ];

  const renderAnalysisContent = (analysis: string) => {
    // 提取深度思考内容
    const thinkMatch = analysis.match(/<think>([\s\S]*?)<\/think>/);
    // 移除 <think> 标签，获取主要内容
    const mainContent = analysis.replace(/<think>[\s\S]*?<\/think>/g, '').trim();

    return (
      <div style={{ padding: '16px' }}>
        {/* 渲染主要内容 */}
        <div className="markdown-content">
          <ReactMarkdown>{mainContent}</ReactMarkdown>
        </div>

        {/* 如果存在深度思考内容，添加折叠面板 */}
        {thinkMatch && (
          <div style={{ 
            marginTop: '16px', 
            borderTop: '1px solid #f0f0f0', 
            paddingTop: '16px' 
          }}>
            <Button 
              type="link" 
              onClick={() => setShowThinkContent(!showThinkContent)}
              icon={showThinkContent ? <UpOutlined /> : <DownOutlined />}
              style={{ padding: 0 }}
            >
              {showThinkContent ? '收起深度思考' : '查看深度思考'}
            </Button>
            
            {showThinkContent && (
              <div style={{ marginTop: '8px' }}>
                <ReactMarkdown>{thinkMatch[1]}</ReactMarkdown>
              </div>
            )}
          </div>
        )}
      </div>
    );
  };

  return (
    <div>
      <Table
        columns={columns}
        dataSource={alerts}
        loading={loading}
        rowKey="id"
      />
      <Modal
        title="告警分析结果"
        open={analysisModalVisible}
        onCancel={() => {
          setAnalysisModalVisible(false);
          setShowThinkContent(false); // 关闭时重置折叠状态
        }}
        footer={null}
        width={800}
      >
        {currentAlert?.analysis ? (
          renderAnalysisContent(currentAlert.analysis)
        ) : (
          <div>{isAnalyzing ? '正在分析中...' : '暂无分析结果'}</div>
        )}
      </Modal>
      
      <AlertDetailModal
         visible={detailModalVisible}
         alertId={selectedAlertId}
         onClose={handleCloseDetailModal}
       />
    </div>
  );
};

export default AlertList;