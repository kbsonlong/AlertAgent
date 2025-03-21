import React, { useEffect, useState, useRef } from 'react';
import { Table, Card, Button, Tag, Space, message, Modal, Progress } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { AlertOutlined, RobotOutlined } from '@ant-design/icons';
// 导入告警相关的API接口
import { getAlerts, updateAlertStatus, asyncAnalyzeAlert, getAnalysisStatus } from '../../services/alert';

interface Alert {
  id: number;
  name: string;
  level: string;
  status: string;
  source: string;
  content: string;
  analysis: string;
  created_at: string;
  updated_at: string;
}

const AlertList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [analysisModalVisible, setAnalysisModalVisible] = useState(false);
  const [currentAlert, setCurrentAlert] = useState<Alert | null>(null);
  const [analyzing, setAnalyzing] = useState(false);
  const [analysisProgress, setAnalysisProgress] = useState(0);
  const [analysisStatus, setAnalysisStatus] = useState<'idle' | 'processing' | 'completed' | 'failed'>('idle');
  const statusCheckInterval = useRef<number | null>(null);

  const columns: ColumnsType<Alert> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '告警名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '级别',
      dataIndex: 'level',
      key: 'level',
      render: (level: string) => {
        const color = level === 'high' ? 'red' : level === 'medium' ? 'orange' : 'green';
        return <Tag color={color}>{level.toUpperCase()}</Tag>;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const color = status === 'active' ? 'red' : status === 'acknowledged' ? 'blue' : 'green';
        return <Tag color={color}>{status.toUpperCase()}</Tag>;
      },
    },
    {
      title: '来源',
      dataIndex: 'source',
      key: 'source',
    },
    {
      title: '内容',
      dataIndex: 'content',
      key: 'content',
      ellipsis: true,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Button type="link" onClick={() => handleAcknowledge(record.id)}>确认</Button>
          <Button type="link" onClick={() => handleResolve(record.id)}>解决</Button>
          <Button 
            type="link" 
            icon={<RobotOutlined />}
            onClick={() => handleAnalyze(record)}
          >
            AI分析
          </Button>
        </Space>
      ),
    },
  ];

  const fetchAlerts = async () => {
    try {
      setLoading(true);
      const data = await getAlerts();
      if (data.code === 200) {
        setAlerts(data.data);
      } else {
        message.error(data.msg || '获取告警列表失败');
      }
    } catch (error) {
      message.error('获取告警列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleAcknowledge = async (id: number) => {
    try {
      const data = await updateAlertStatus(id, 'acknowledged');
      if (data.code === 200) {
        message.success('告警已确认');
        fetchAlerts();
      } else {
        message.error(data.msg || '确认告警失败');
      }
    } catch (error) {
      message.error('确认告警失败');
    }
  };

  const handleResolve = async (id: number) => {
    try {
      const data = await updateAlertStatus(id, 'resolved');
      if (data.code === 200) {
        message.success('告警已解决');
        fetchAlerts();
      } else {
        message.error(data.msg || '解决告警失败');
      }
    } catch (error) {
      message.error('解决告警失败');
    }
  };

  // 清除状态检查定时器
  const clearStatusCheckInterval = () => {
    if (statusCheckInterval.current) {
      window.clearInterval(statusCheckInterval.current);
      statusCheckInterval.current = null;
    }
  };

  // 检查分析状态
  const checkAnalysisStatus = async (alertId: number) => {
    try {
      const data = await getAnalysisStatus(alertId);
      if (data.code === 200) {
        const { status, analysis, error } = data.data;
        
        if (status === 'completed' && analysis) {
          // 分析完成，更新结果
          setAnalysisStatus('completed');
          setAnalysisProgress(100);
          setAnalyzing(false);
          
          // 更新告警列表中的分析结果
          setAlerts(alerts.map(a => 
            a.id === alertId ? { ...a, analysis } : a
          ));
          
          if (currentAlert && currentAlert.id === alertId) {
            setCurrentAlert({ ...currentAlert, analysis });
          }
          
          clearStatusCheckInterval();
          message.success('分析完成');
        } else if (status === 'failed') {
          // 分析失败
          setAnalysisStatus('failed');
          setAnalyzing(false);
          clearStatusCheckInterval();
          message.error(error || 'AI分析失败');
        } else if (status === 'processing') {
          // 分析中，更新进度
          setAnalysisStatus('processing');
          // 模拟进度，最多到95%
          setAnalysisProgress(prev => Math.min(prev + 5, 95));
        }
      } else {
        message.error(data.msg || '获取分析状态失败');
      }
    } catch (error) {
      message.error('获取分析状态失败');
    }
  };

  const handleAnalyze = async (alert: Alert) => {
    setCurrentAlert(alert);
    setAnalysisModalVisible(true);
    
    // 如果已有分析结果，直接显示
    if (alert.analysis) {
      setAnalysisStatus('completed');
      setAnalysisProgress(100);
      return;
    }
    
    // 重置状态
    setAnalysisStatus('processing');
    setAnalysisProgress(0);
    setAnalyzing(true);
    
    try {
      // 调用异步分析接口
      const data = await asyncAnalyzeAlert(alert.id);
      if (data.code === 200) {
        message.success(data.msg || '分析任务已提交');
        
        // 设置定时检查分析状态
        clearStatusCheckInterval();
        statusCheckInterval.current = window.setInterval(() => {
          checkAnalysisStatus(alert.id);
        }, 2000); // 每2秒检查一次
        
        // 初始进度设为10%
        setAnalysisProgress(10);
      } else {
        setAnalyzing(false);
        setAnalysisStatus('failed');
        message.error(data.msg || '提交分析任务失败');
      }
    } catch (error) {
      setAnalyzing(false);
      setAnalysisStatus('failed');
      message.error('提交分析任务失败');
    }
  };
  
  // 组件卸载时清除定时器
  useEffect(() => {
    return () => clearStatusCheckInterval();
  }, []);

  useEffect(() => {
    fetchAlerts();
  }, []);

  return (
    <>
      <Card
        title={
          <Space>
            <AlertOutlined />
            告警列表
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={alerts}
          rowKey="id"
          loading={loading}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
        />
      </Card>

      <Modal
        title={
          <Space>
            <RobotOutlined />
            AI分析结果
          </Space>
        }
        open={analysisModalVisible}
        onCancel={() => {
          setAnalysisModalVisible(false);
          // 关闭Modal时清除定时器，避免后台继续更新状态
          if (analyzing) {
            clearStatusCheckInterval();
            setAnalyzing(false);
          }
        }}
        footer={[
          <Button key="close" onClick={() => {
            setAnalysisModalVisible(false);
            // 关闭Modal时清除定时器，避免后台继续更新状态
            if (analyzing) {
              clearStatusCheckInterval();
              setAnalyzing(false);
            }
          }}>
            关闭
          </Button>
        ]}
        width={800}
      >
        {analyzing ? (
          <div style={{ textAlign: 'center', padding: '20px' }}>
            <div style={{ marginBottom: '20px' }}>正在分析中，请稍候...</div>
            <Progress percent={analysisProgress} status="active" />
          </div>
        ) : analysisStatus === 'failed' ? (
          <div style={{ textAlign: 'center', color: '#ff4d4f', padding: '20px' }}>
            分析失败，请稍后重试
          </div>
        ) : currentAlert?.analysis ? (
          <div style={{ whiteSpace: 'pre-wrap' }}>
            {currentAlert.analysis}
          </div>
        ) : (
          <div style={{ textAlign: 'center', color: '#999' }}>
            暂无分析结果
          </div>
        )}
      </Modal>
    </>
  );
};

export default AlertList;