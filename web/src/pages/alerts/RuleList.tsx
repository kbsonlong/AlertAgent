import React, { useEffect, useState } from 'react';
import { Table, Card, Button, Tag, Space, message, Modal, Form, Input, Select, Switch } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { AlertOutlined, PlusOutlined } from '@ant-design/icons';

interface Rule {
  id: number;
  name: string;
  description: string;
  level: string;
  enabled: boolean;
  condition: string;
  created_at: string;
  updated_at: string;
}

const RuleList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [rules, setRules] = useState<Rule[]>([]);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [form] = Form.useForm();

  const columns: ColumnsType<Rule> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '规则名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
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
      dataIndex: 'enabled',
      key: 'enabled',
      render: (_, record) => (
        <Switch
          checked={record.enabled}
          onChange={(checked) => handleToggleStatus(record.id, checked)}
        />
      ),
    },
    {
      title: '条件',
      dataIndex: 'condition',
      key: 'condition',
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
          <Button type="link" onClick={() => handleEdit(record)}>编辑</Button>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ];

  const fetchRules = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/v1/rules', {
        headers: {
          'Accept': 'application/json; charset=utf-8',
          'Content-Type': 'application/json; charset=utf-8',
          'Accept-Charset': 'utf-8'
        }
      });
      const text = await response.text();
      const data = JSON.parse(text);
      if (data.code === 200) {
        setRules(data.data);
      } else {
        message.error(data.msg || '获取规则列表失败');
      }
    } catch (error) {
      message.error('获取规则列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleToggleStatus = async (id: number, enabled: boolean) => {
    try {
      const response = await fetch(`/api/v1/rules/${id}`, {
        method: 'PUT',
        headers: {
          'Accept': 'application/json; charset=utf-8',
          'Content-Type': 'application/json; charset=utf-8',
          'Accept-Charset': 'utf-8'
        },
        body: JSON.stringify({ enabled }),
      });
      const text = await response.text();
      const data = JSON.parse(text);
      if (data.code === 200) {
        message.success(`规则已${enabled ? '启用' : '禁用'}`);
        fetchRules();
      } else {
        message.error(data.msg || '操作失败');
      }
    } catch (error) {
      message.error('操作失败');
    }
  };

  const handleEdit = (record: Rule) => {
    form.setFieldsValue(record);
    setIsModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      const response = await fetch(`/api/v1/rules/${id}`, {
        method: 'DELETE',
        headers: {
          'Accept': 'application/json; charset=utf-8',
          'Content-Type': 'application/json; charset=utf-8',
          'Accept-Charset': 'utf-8'
        }
      });
      const text = await response.text();
      const data = JSON.parse(text);
      if (data.code === 200) {
        message.success('规则已删除');
        fetchRules();
      } else {
        message.error(data.msg || '删除失败');
      }
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      const url = values.id ? `/api/v1/rules/${values.id}` : '/api/v1/rules';
      const method = values.id ? 'PUT' : 'POST';
      const response = await fetch(url, {
        method,
        headers: {
          'Accept': 'application/json; charset=utf-8',
          'Content-Type': 'application/json; charset=utf-8',
          'Accept-Charset': 'utf-8'
        },
        body: JSON.stringify(values),
      });
      const text = await response.text();
      const data = JSON.parse(text);
      if (data.code === 200) {
        message.success(`${values.id ? '更新' : '创建'}成功`);
        setIsModalVisible(false);
        form.resetFields();
        fetchRules();
      } else {
        message.error(data.msg || `${values.id ? '更新' : '创建'}失败`);
      }
    } catch (error) {
      message.error(`${values.id ? '更新' : '创建'}失败`);
    }
  };

  useEffect(() => {
    fetchRules();
  }, []);

  return (
    <>
      <Card
        title={
          <Space>
            <AlertOutlined />
            告警规则
          </Space>
        }
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsModalVisible(true)}>
            新建规则
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={rules}
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
        title={form.getFieldValue('id') ? '编辑规则' : '新建规则'}
        open={isModalVisible}
        onOk={() => form.submit()}
        onCancel={() => {
          setIsModalVisible(false);
          form.resetFields();
        }}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item name="id" hidden>
            <Input />
          </Form.Item>
          <Form.Item
            name="name"
            label="规则名称"
            rules={[{ required: true, message: '请输入规则名称' }]}
          >
            <Input placeholder="请输入规则名称" />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
            rules={[{ required: true, message: '请输入规则描述' }]}
          >
            <Input.TextArea rows={4} placeholder="请输入规则描述" />
          </Form.Item>
          <Form.Item
            name="level"
            label="级别"
            rules={[{ required: true, message: '请选择告警级别' }]}
          >
            <Select>
              <Select.Option value="high">高</Select.Option>
              <Select.Option value="medium">中</Select.Option>
              <Select.Option value="low">低</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="condition"
            label="条件"
            rules={[{ required: true, message: '请输入告警条件' }]}
          >
            <Input.TextArea rows={4} placeholder="请输入告警条件" />
          </Form.Item>
          <Form.Item
            name="enabled"
            label="状态"
            valuePropName="checked"
            initialValue={true}
          >
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default RuleList; 