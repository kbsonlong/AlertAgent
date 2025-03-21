import React, { useEffect, useState } from 'react';
import { Table, Card, Button, Space, message, Modal, Form, Input, Select } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { MailOutlined, PlusOutlined } from '@ant-design/icons';
import { getTemplates, createTemplate, updateTemplate, deleteTemplate } from '../../services/notification';

interface Template {
  id: number;
  name: string;
  type: string;
  content: string;
  created_at: string;
  updated_at: string;
}

const TemplateList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [form] = Form.useForm();

  const columns: ColumnsType<Template> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '模板名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
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
          <Button type="link" onClick={() => handleEdit(record)}>编辑</Button>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ];

  const fetchTemplates = async () => {
    try {
      setLoading(true);
      const data = await getTemplates();
      if (data.code === 200) {
        setTemplates(data.data);
      } else {
        message.error(data.msg || '获取模板列表失败');
      }
    } catch (error) {
      message.error('获取模板列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (record: Template) => {
    form.setFieldsValue(record);
    setIsModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      const data = await deleteTemplate(id);
      if (data.code === 200) {
        message.success('模板已删除');
        fetchTemplates();
      } else {
        message.error(data.msg || '删除失败');
      }
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      let data;
      if (values.id) {
        data = await updateTemplate(values.id, values);
      } else {
        data = await createTemplate(values);
      }
      
      if (data.code === 200) {
        message.success(`${values.id ? '更新' : '创建'}成功`);
        setIsModalVisible(false);
        form.resetFields();
        fetchTemplates();
      } else {
        message.error(data.msg || `${values.id ? '更新' : '创建'}失败`);
      }
    } catch (error) {
      message.error(`${values.id ? '更新' : '创建'}失败`);
    }
  };

  useEffect(() => {
    fetchTemplates();
  }, []);

  return (
    <>
      <Card
        title={
          <Space>
            <MailOutlined />
            通知模板
          </Space>
        }
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsModalVisible(true)}>
            新建模板
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={templates}
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
        title={form.getFieldValue('id') ? '编辑模板' : '新建模板'}
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
            label="模板名称"
            rules={[{ required: true, message: '请输入模板名称' }]}
          >
            <Input placeholder="请输入模板名称" />
          </Form.Item>
          <Form.Item
            name="type"
            label="类型"
            rules={[{ required: true, message: '请选择通知类型' }]}
          >
            <Select>
              <Select.Option value="email">邮件</Select.Option>
              <Select.Option value="sms">短信</Select.Option>
              <Select.Option value="webhook">Webhook</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="content"
            label="内容"
            rules={[{ required: true, message: '请输入模板内容' }]}
          >
            <Input.TextArea
              rows={6}
              placeholder="请输入模板内容，支持以下变量：&#13;&#10;${alert.name} - 告警名称&#13;&#10;${alert.level} - 告警级别&#13;&#10;${alert.content} - 告警内容"
            />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default TemplateList;