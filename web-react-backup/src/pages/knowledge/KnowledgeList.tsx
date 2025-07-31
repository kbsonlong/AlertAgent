import React, { useEffect, useState } from 'react';
import { Card, Input, Tag, Space, Table, message, Button } from 'antd';
import { Knowledge, KnowledgeListParams, getKnowledgeList } from '../../services/knowledge';
import { useNavigate } from 'react-router-dom';
import AlertDetailModal from '../../components/AlertDetailModal';

const { Search } = Input;

const KnowledgeList: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [list, setList] = useState<Knowledge[]>([]);
  const [searchParams, setSearchParams] = useState<KnowledgeListParams>({
    page: 1,
    pageSize: 10,
  });
  const [alertDetailVisible, setAlertDetailVisible] = useState(false);
  const [selectedAlertId, setSelectedAlertId] = useState<number | null>(null);

  const fetchList = async (params: KnowledgeListParams) => {
    try {
      setLoading(true);
      const response = await getKnowledgeList(params);
      if (response && response.code === 200) {
        setList(response.data.list || []);
        setTotal(response.data.total || 0);
      }
    } catch (error) {
      console.error(error);
      message.error('获取知识列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchList(searchParams);
  }, [searchParams]);

  const handleSearch = (value: string) => {
    setSearchParams((prev: KnowledgeListParams) => ({ ...prev, keyword: value, page: 1 }));
  };

  const handleCategoryFilter = (category: string) => {
    setSearchParams(prev => ({ ...prev, category, page: 1 }));
  };

  const handleViewAlert = (alertId: number) => {
    setSelectedAlertId(alertId);
    setAlertDetailVisible(true);
  };

  const handleCloseAlertDetail = () => {
    setAlertDetailVisible(false);
    setSelectedAlertId(null);
  };

  const columns = [
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      width: '40%',
      render: (text: string, record: Knowledge) => (
        <a onClick={() => navigate(`/knowledge/${record.id}`)}>{text}</a>
      ),
    },
    {
      title: '来源',
      dataIndex: 'source',
      key: 'source',
      width: '20%',
    },
    {
      title: '关联告警',
      dataIndex: 'source_id',
      key: 'source_id',
      width: '20%',
      render: (sourceId: number, record: Knowledge) => {
        if (!sourceId || record.source !== 'alert') return '-';
        return (
          <Button 
            type="link" 
            size="small"
            onClick={() => handleViewAlert(sourceId)}
          >
            告警#{sourceId}
          </Button>
        );
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: '20%',
      render: (text: string) => {
        if (!text) return '-';
        return new Date(text).toLocaleString('zh-CN');
      },
    },
  ];

  return (
    <Card>
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        <Search
          placeholder="搜索知识库"
          allowClear
          enterButton
          onSearch={handleSearch}
          style={{ width: 300 }}
        />
        <Table
          loading={loading}
          columns={columns}
          dataSource={list}
          rowKey="id"
          pagination={{
            total,
            current: searchParams.page,
            pageSize: searchParams.pageSize,
            onChange: (page, pageSize) =>
              setSearchParams(prev => ({ ...prev, page, pageSize })),
          }}
        />
      </Space>
      
      <AlertDetailModal
        visible={alertDetailVisible}
        alertId={selectedAlertId}
        onClose={handleCloseAlertDetail}
      />
    </Card>
  );
};

export default KnowledgeList;