import React, { useState, useEffect } from 'react';
import { Modal, Descriptions, Tag, message } from 'antd';
import { Alert, getAlert } from '../services/alert';
import { formatDateTime } from '../utils/datetime';
import ReactMarkdown from 'react-markdown';

interface AlertDetailModalProps {
  visible: boolean;
  alertId: number | null;
  onClose: () => void;
}

const AlertDetailModal: React.FC<AlertDetailModalProps> = ({ visible, alertId, onClose }) => {
  const [alert, setAlert] = useState<Alert | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (visible && alertId) {
      fetchAlertDetail();
    }
  }, [visible, alertId]);

  const fetchAlertDetail = async () => {
    if (!alertId) return;
    
    setLoading(true);
    try {
      const alertData = await getAlert(alertId);
      setAlert(alertData);
    } catch (error) {
      message.error('获取告警详情失败');
    } finally {
      setLoading(false);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'red';
      case 'high': return 'orange';
      case 'medium': return 'yellow';
      case 'low': return 'green';
      default: return 'default';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'new': return 'red';
      case 'acknowledged': return 'orange';
      case 'resolved': return 'green';
      default: return 'default';
    }
  };

  const handleClose = () => {
    setAlert(null);
    onClose();
  };

  return (
    <Modal
      title="告警详情"
      open={visible}
      onCancel={handleClose}
      footer={null}
      width={800}
      loading={loading}
    >
      {alert && (
        <Descriptions column={1} bordered>
          <Descriptions.Item label="标题">{alert.title}</Descriptions.Item>
          <Descriptions.Item label="内容">{alert.content}</Descriptions.Item>
        </Descriptions>
      )}
      
      {alert && (
        <Descriptions column={3} bordered style={{ marginTop: 16 }}>
          <Descriptions.Item label="严重程度">
            <Tag color={getSeverityColor(alert.severity)}>
              {alert.severity?.toUpperCase() || '未知'}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="来源">{alert.source}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={getStatusColor(alert.status)}>
              {alert.status?.toUpperCase() || '未知'}
            </Tag>
          </Descriptions.Item>
        </Descriptions>
      )}
      
      {alert && (
        <Descriptions column={2} bordered style={{ marginTop: 16 }}>
          <Descriptions.Item label="创建时间">
            {formatDateTime(alert.created_at || '')}
          </Descriptions.Item>
          <Descriptions.Item label="更新时间">
            {formatDateTime(alert.updated_at || '')}
          </Descriptions.Item>
        </Descriptions>
      )}
      
      {alert && alert.analysis && (
        <Descriptions column={1} bordered style={{ marginTop: 16 }}>
          <Descriptions.Item label="分析结果">
            <div style={{ maxHeight: '200px', overflow: 'auto' }}>
              <ReactMarkdown>{alert.analysis}</ReactMarkdown>
            </div>
          </Descriptions.Item>
        </Descriptions>
      )}
    </Modal>
  );
};

export default AlertDetailModal;