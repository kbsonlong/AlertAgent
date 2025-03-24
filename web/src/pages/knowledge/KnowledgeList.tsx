import React, { useEffect, useState } from 'react';
import { Card, Input, Tag, Space, Table, message } from 'antd';
import { Knowledge, KnowledgeListParams, getKnowledgeList } from '../../services/knowledge';

const { Search } = Input;

const KnowledgeList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [list, setList] = useState<Knowledge[]>([]);
  const [searchParams, setSearchParams] = useState<KnowledgeListParams>({
    page: 1,
    pageSize: 10,
  });

  const fetchList = async (params: KnowledgeListParams) => {
    try {
      setLoading(true);
      const res = await getKnowledgeList(params);
      setList(res.items);
      setTotal(res.total);
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

  const columns = [
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      width: '30%',
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      render: (category: string) => (
        <Tag color="blue" onClick={() => handleCategoryFilter(category)}>
          {category}
        </Tag>
      ),
    },
    {
      title: '标签',
      dataIndex: 'tags',
      key: 'tags',
      render: (tags: string[]) => (
        <Space>
          {tags.map(tag => (
            <Tag key={tag}>{tag}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '来源',
      dataIndex: 'source',
      key: 'source',
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
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
    </Card>
  );
};

export default KnowledgeList;