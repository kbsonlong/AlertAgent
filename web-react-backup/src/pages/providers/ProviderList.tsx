import React, { useEffect, useState } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  message,
  Popconfirm,
  Input,
  Select,
  Modal,
  Tooltip,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import {
  Provider,
  ProviderListParams,
  getProviderList,
  deleteProvider,
  PROVIDER_TYPES,
  PROVIDER_STATUS,
} from '../../services/provider';
import ProviderForm from './ProviderForm';

const { Search } = Input;
const { Option } = Select;

const ProviderList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [list, setList] = useState<Provider[]>([]);
  const [total, setTotal] = useState(0);
  const [searchParams, setSearchParams] = useState<ProviderListParams>({
    page: 1,
    pageSize: 10,
  });
  const [formVisible, setFormVisible] = useState(false);
  const [editingProvider, setEditingProvider] = useState<Provider | null>(null);

  const fetchList = async (params: ProviderListParams) => {
    try {
      setLoading(true);
      const response = await getProviderList(params);
      const data = response.data || [];
      setList(data);
      setTotal(data.length);
    } catch (error) {
      console.error(error);
      message.error('获取数据源列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchList(searchParams);
  }, [searchParams]);

  const handleSearch = (value: string) => {
    setSearchParams(prev => ({ ...prev, keyword: value, page: 1 }));
  };

  const handleTypeFilter = (value: string) => {
    setSearchParams(prev => ({ ...prev, type: value || undefined, page: 1 }));
  };

  const handleStatusFilter = (value: string) => {
    setSearchParams(prev => ({ ...prev, status: value || undefined, page: 1 }));
  };

  const handleCreate = () => {
    setEditingProvider(null);
    setFormVisible(true);
  };

  const handleEdit = (record: Provider) => {
    setEditingProvider(record);
    setFormVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteProvider(id);
      message.success('删除成功');
      fetchList(searchParams);
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleFormSuccess = () => {
    setFormVisible(false);
    setEditingProvider(null);
    fetchList(searchParams);
  };

  const handleFormCancel = () => {
    setFormVisible(false);
    setEditingProvider(null);
  };

  const getStatusTag = (status: string) => {
    const statusConfig = {
      active: { color: 'green', icon: <CheckCircleOutlined />, text: '活跃' },
      inactive: { color: 'red', icon: <ExclamationCircleOutlined />, text: '非活跃' },
    };
    const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.inactive;
    return (
      <Tag color={config.color} icon={config.icon}>
        {config.text}
      </Tag>
    );
  };

  const getTypeTag = (type: string) => {
    const typeConfig = {
      prometheus: { color: 'blue', text: 'Prometheus' },
      victoriametrics: { color: 'purple', text: 'VictoriaMetrics' },
    };
    const config = typeConfig[type as keyof typeof typeConfig] || { color: 'default', text: type };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: '20%',
      render: (text: string, record: Provider) => (
        <div>
          <div style={{ fontWeight: 500 }}>{text}</div>
          {record.description && (
            <div style={{ fontSize: '12px', color: '#666', marginTop: '2px' }}>
              {record.description}
            </div>
          )}
        </div>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: '15%',
      render: (type: string) => getTypeTag(type),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: '10%',
      render: (status: string) => getStatusTag(status),
    },
    {
      title: '端点地址',
      dataIndex: 'endpoint',
      key: 'endpoint',
      width: '25%',
      render: (endpoint: string) => (
        <Tooltip title={endpoint}>
          <div style={{ maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis' }}>
            {endpoint}
          </div>
        </Tooltip>
      ),
    },
    {
      title: '最后检查',
      dataIndex: 'last_check',
      key: 'last_check',
      width: '15%',
      render: (lastCheck: string) => {
        if (!lastCheck) return '-';
        return new Date(lastCheck).toLocaleString('zh-CN');
      },
    },
    {
      title: '操作',
      key: 'action',
      width: '15%',
      render: (_: any, record: Provider) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个数据源吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card>
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        {/* 搜索和筛选 */}
        <Space wrap>
          <Search
            placeholder="搜索数据源名称"
            allowClear
            style={{ width: 250 }}
            onSearch={handleSearch}
          />
          <Select
            placeholder="选择类型"
            allowClear
            style={{ width: 150 }}
            onChange={handleTypeFilter}
          >
            {PROVIDER_TYPES.map(type => (
              <Option key={type.value} value={type.value}>
                {type.label}
              </Option>
            ))}
          </Select>
          <Select
            placeholder="选择状态"
            allowClear
            style={{ width: 120 }}
            onChange={handleStatusFilter}
          >
            {PROVIDER_STATUS.map(status => (
              <Option key={status.value} value={status.value}>
                {status.label}
              </Option>
            ))}
          </Select>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            新建数据源
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => fetchList(searchParams)}
          >
            刷新
          </Button>
        </Space>

        {/* 表格 */}
        <Table
          loading={loading}
          columns={columns}
          dataSource={list}
          rowKey="id"
          pagination={{
            total,
            current: searchParams.page,
            pageSize: searchParams.pageSize,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
            onChange: (page, pageSize) =>
              setSearchParams(prev => ({ ...prev, page, pageSize })),
          }}
        />
      </Space>

      {/* 表单弹窗 */}
      <Modal
        title={editingProvider ? '编辑数据源' : '新建数据源'}
        open={formVisible}
        onCancel={handleFormCancel}
        footer={null}
        width={600}
        destroyOnClose
      >
        <ProviderForm
          provider={editingProvider}
          onSuccess={handleFormSuccess}
          onCancel={handleFormCancel}
        />
      </Modal>
    </Card>
  );
};

export default ProviderList;