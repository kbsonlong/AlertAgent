import React, { useEffect, useState } from 'react';
import { Table, Button, Space, Modal, Tag, message } from 'antd';
import { Alert, getAlerts, updateAlertStatus, asyncAnalyzeAlert, getAnalysisStatus, convertToKnowledge } from '../../services/alert';
import { formatDateTime } from '../../utils/datetime';
import ReactMarkdown from 'react-markdown';
import { DownOutlined, UpOutlined } from '@ant-design/icons';

const AlertList: React.FC = () => {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(false);
  const [analysisModalVisible, setAnalysisModalVisible] = useState(false);
  const [currentAlert, setCurrentAlert] = useState<Alert | null>(null);
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [showThinkContent, setShowThinkContent] = useState(false);

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

  useEffect(() => {
    fetchAlerts();
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
      await asyncAnalyzeAlert(record.id);
      
      // 轮询分析状态
      const checkStatus = async () => {
        const analysis = await getAnalysisStatus(record.id);
        if (analysis.status === 'completed' && analysis.result) {
          setCurrentAlert(prev => prev ? { ...prev, analysis: analysis.result } : null);
          setIsAnalyzing(false);
        } else if (analysis.status === 'failed') {
          message.error('分析失败');
          setIsAnalyzing(false);
        } else {
          setTimeout(checkStatus, 2000);
        }
      };

      checkStatus();
    } catch {
      message.error('启动分析失败');
      setIsAnalyzing(false);
    }
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
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (text: string | undefined) => formatDateTime(text || ''),
      sorter: (a: Alert, b: Alert) => {
        if (!a.createdAt || !b.createdAt) return 0;
        return new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
      },
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: Alert) => (
        <Space size="middle">
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
          <Button type="link" onClick={() => handleAnalyze(record)}>
            分析
          </Button>
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
    </div>
  );
};

export default AlertList;