import React, { useEffect, useState } from 'react';
import { Card, Tag, Space, Typography, Spin, message } from 'antd';
import { useParams } from 'react-router-dom';
import { Knowledge, getKnowledgeById } from '../../services/knowledge';

const { Title, Paragraph, Text } = Typography;

const KnowledgeDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [loading, setLoading] = useState(false);
  const [detail, setDetail] = useState<Knowledge>();

  const fetchDetail = async () => {
    if (!id) return;
    try {
      setLoading(true);
      const res = await getKnowledgeById(Number(id));
      setDetail(res);
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
        <Space>
          <Text type="secondary">分类：</Text>
          <Tag color="blue">{detail.category}</Tag>
        </Space>
        {detail.tags && detail.tags.length > 0 && (
          <Space>
            <Text type="secondary">标签：</Text>
            {detail.tags.map(tag => (
              <Tag key={tag}>{tag}</Tag>
            ))}
          </Space>
        )}
        <Space>
          <Text type="secondary">来源：</Text>
          <Text>{detail.source}</Text>
        </Space>
        <Space>
          <Text type="secondary">创建时间：</Text>
          <Text>{detail.createdAt}</Text>
        </Space>
        {detail.summary && (
          <div>
            <Text type="secondary">摘要：</Text>
            <Paragraph>{detail.summary}</Paragraph>
          </div>
        )}
        <div>
          <Text type="secondary">内容：</Text>
          <Paragraph style={{ whiteSpace: 'pre-wrap' }}>{detail.content}</Paragraph>
        </div>
      </Space>
    </Card>
  );
};

export default KnowledgeDetail;