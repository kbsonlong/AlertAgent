import React, { useEffect, useState } from 'react';
import { Card, Tag, Space, Typography, Spin, message } from 'antd';
import { useParams } from 'react-router-dom';
import { Knowledge, getKnowledgeById } from '../../services/knowledge';
import ReactMarkdown from 'react-markdown';

const { Title, Text } = Typography;

const KnowledgeDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [loading, setLoading] = useState(false);
  const [detail, setDetail] = useState<Knowledge>();

  const fetchDetail = async () => {
    if (!id) return;
    try {
      setLoading(true);
      const res = await getKnowledgeById(Number(id));
      const data = res.data;
      setDetail(data);
    } catch (error) {
      console.error(error);
      message.error('获取知识详情失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDetail();
  }, [id]);

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!detail) {
    return null;
  }

  return (
    <Card>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Title level={2}>{detail.title}</Title>
        <Space wrap>
          <Space>
            <Text type="secondary">来源：</Text>
            <Text>{detail.source}</Text>
          </Space>
          <Space>
            <Text type="secondary">关联告警ID：</Text>
            <Text>{detail.source_id}</Text>
          </Space>
          <Space>
            <Text type="secondary">创建时间：</Text>
            <Text>{new Date(detail.created_at).toLocaleString('zh-CN')}</Text>
          </Space>
        </Space>
        <div>
          <div className="markdown-content">
            <ReactMarkdown>{detail.content}</ReactMarkdown>
          </div>
        </div>
      </Space>
      <style>
        {`
          .markdown-content {
            padding: 16px;
            background: #fafafa;
            border-radius: 4px;
          }
          .markdown-content h1 { font-size: 24px; margin-top: 24px; margin-bottom: 16px; }
          .markdown-content h2 { font-size: 20px; margin-top: 24px; margin-bottom: 16px; }
          .markdown-content h3 { font-size: 18px; margin-top: 24px; margin-bottom: 16px; }
          .markdown-content h4 { font-size: 16px; margin-top: 24px; margin-bottom: 16px; }
          .markdown-content p { margin-bottom: 16px; line-height: 1.6; }
          .markdown-content ul, .markdown-content ol { margin-bottom: 16px; padding-left: 24px; }
          .markdown-content li { margin-bottom: 8px; }
          .markdown-content code {
            background: #f0f0f0;
            padding: 2px 4px;
            border-radius: 2px;
            font-family: monospace;
          }
          .markdown-content pre {
            background: #f0f0f0;
            padding: 16px;
            border-radius: 4px;
            overflow-x: auto;
          }
          .markdown-content blockquote {
            margin: 16px 0;
            padding: 0 16px;
            color: #666;
            border-left: 4px solid #ddd;
          }
          .markdown-content img {
            max-width: 100%;
            height: auto;
          }
          .markdown-content table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 16px;
          }
          .markdown-content th, .markdown-content td {
            border: 1px solid #ddd;
            padding: 8px;
          }
          .markdown-content th {
            background: #f5f5f5;
          }
        `}
      </style>
    </Card>
  );
};

export default KnowledgeDetail;